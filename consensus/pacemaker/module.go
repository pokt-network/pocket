package pacemaker

import (
	"context"
	"fmt"
	"time"

	consensusTelemetry "github.com/pokt-network/pocket/consensus/telemetry"
	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	defaultLogPrefix    = "NODE"
	pacemakerModuleName = "pacemaker"

	// A buffer around the pacemaker timeout to avoid race condition; 100ms was arbitrarily chosen
	timeoutBuffer = 100 * time.Millisecond

	newRound = typesCons.HotstuffStep_HOTSTUFF_STEP_NEWROUND
	propose  = typesCons.HotstuffMessageType_HOTSTUFF_MESSAGE_PROPOSE
)

var (
	_ modules.Module = &pacemaker{}
	_ PacemakerDebug = &pacemaker{}
	_ Pacemaker      = &pacemaker{}
)

type Pacemaker interface {
	modules.Module
	PacemakerDebug

	ShouldHandleMessage(message *typesCons.HotstuffMessage) (bool, error)

	RestartTimer()
	NewHeight()
	InterruptRound(reason string)
}

type pacemaker struct {
	base_modules.IntegratableModule
	base_modules.InterruptableModule

	pacemakerCfg    *configs.PacemakerConfig
	roundTimeout    time.Duration
	roundCancelFunc context.CancelFunc

	// Only used for development and debugging.
	debug pacemakerDebug

	logger *modules.Logger
	// REFACTOR: logPrefix should be removed in exchange for setting a namespace directly with the logger
	logPrefix string
}

func CreatePacemaker(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	return new(pacemaker).Create(bus, options...)
}

func (*pacemaker) Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	m := &pacemaker{
		logPrefix: defaultLogPrefix,
	}

	for _, option := range options {
		option(m)
	}

	bus.RegisterModule(m)

	runtimeMgr := bus.GetRuntimeMgr()
	cfg := runtimeMgr.GetConfig()

	m.pacemakerCfg = cfg.Consensus.PacemakerConfig
	m.roundTimeout = m.getRoundTimeout()
	m.debug = pacemakerDebug{
		manualMode:                m.pacemakerCfg.GetManual(),
		debugTimeBetweenStepsMsec: m.pacemakerCfg.GetDebugTimeBetweenStepsMsec(),
		quorumCertificate:         nil,
	}

	return m, nil
}

func (m *pacemaker) Start() error {
	m.logger = logger.Global.CreateLoggerForModule(m.GetModuleName())
	m.RestartTimer()
	return nil
}

func (*pacemaker) GetModuleName() string {
	return pacemakerModuleName
}

func (m *pacemaker) ShouldHandleMessage(msg *typesCons.HotstuffMessage) (bool, error) {
	consensusMod := m.GetBus().GetConsensusModule()

	currentHeight := consensusMod.CurrentHeight()
	currentRound := consensusMod.CurrentRound()
	currentStep := typesCons.HotstuffStep(consensusMod.CurrentStep())

	// Consensus message is from the past
	if currentHeight > msg.Height {
		m.logger.Warn().Msgf("⚠️ [DISCARDING] ⚠️ Node (ahead) at height %d > message height %d", currentHeight, msg.Height)
		return false, nil
	}

	// If this case happens, there are two possibilities:
	// 1. The node is behind and needs to catch up, node must start syncing,
	// 2. The leader is sending a malicious proposal.
	// There, for both cases, node rejects the proposal, because:
	// 1. If node is out of sync, node can't verify the block proposal, so rejects it. But node will eventually sync with the rest of the network and add the block.
	// 2. If node is synced, node must reject the proposal because proposal is not valid.
	if msg.Height > currentHeight {
		m.logger.Info().Msgf("⚠️ [WARN] ⚠️ Node at height %d < message height %d", currentHeight, msg.Height)
		err := m.GetBus().GetStateMachineModule().SendEvent(coreTypes.StateMachineEvent_Consensus_IsUnsynced)
		return false, err
	}

	// TODO(olshansky): This code branch is a result of the optimization in the leader
	//                  handlers. Since the leader also acts as a replica but doesn't use the replica's
	//                  handlers given the current implementation, it is safe to drop proposal that the leader made to itself.
	// Do not handle messages if it is a self proposal
	if consensusMod.IsLeader() && msg.Type == propose && msg.Step != newRound {
		return false, nil
	}

	// Message is from the past
	if msg.Round < currentRound || (msg.Round == currentRound && msg.Step < currentStep) {
		m.logger.Warn().Msgf("⚠️ [DISCARDING] ⚠️ Node at (height, step, round) (%d, %d, %d) > message at (%d, %d, %d)", currentHeight, currentStep, currentRound, msg.Height, msg.Step, msg.Round)
		return false, nil
	}

	// Everything checks out!
	if msg.Height == currentHeight && msg.Step == currentStep && msg.Round == currentRound {
		return true, nil
	}

	// pacemaker catch up! Node is synced to the right height, but on a previous step/round so we just jump to the latest state.
	if msg.Round > currentRound || (msg.Round == currentRound && msg.Step > currentStep) {
		m.logger.Info().Msg(pacemakerCatchupLog(currentHeight, uint64(currentStep), currentRound, msg.Height, uint64(msg.Step), msg.Round))
		consensusMod.SetStep(uint8(msg.Step))
		consensusMod.SetRound(msg.Round)

		// TODO: Add tests for this. When we catch up to a later step, the leader is still the same.
		// However, when we catch up to a later round, the leader at the same height will be different.
		if currentRound != msg.Round || !consensusMod.IsLeaderSet() {
			anyProto, err := anypb.New(msg)
			if err != nil {
				m.logger.Warn().Err(err).Msg("Failed to convert pacemaker message to proto.")
				return false, err
			}
			// TODO: Add new custom error
			if err := consensusMod.NewLeader(anyProto); err != nil {
				return false, err
			}
		}

		return true, nil
	}

	return false, typesCons.ErrUnexpectedPacemakerCase
}

func (m *pacemaker) RestartTimer() {
	// NOTE: Not deferring a cancel call because this function is asynchronous.
	if m.roundCancelFunc != nil {
		m.roundCancelFunc()
	}

	clock := m.GetBus().GetRuntimeMgr().GetClock()
	ctx, cancel := clock.WithTimeout(context.TODO(), m.roundTimeout)
	m.roundCancelFunc = cancel
	// NOTE: Not deferring a cancel call because this function is asynchronous.
	go func() {
		select {
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				m.InterruptRound("pacemaker timeout")
			}
		case <-clock.After(m.roundTimeout + timeoutBuffer):
			return
		}
	}()
}

func (m *pacemaker) InterruptRound(reason string) {
	defer m.RestartTimer()

	consensusMod := m.GetBus().GetConsensusModule()
	m.logger.Warn().Fields(m.sharedLoggingFields()).Msgf("⏰ Interrupt ⏰ due to: %s", reason)

	consensusMod.SetRound(consensusMod.CurrentRound() + 1)

	// ADDTEST: check if this is indeed ensured after a successful round
	if m.GetBus().GetConsensusModule().IsPrepareQCNil() {
		m.startNextView(nil, false)
		return
	}

	msgAny, err := consensusMod.GetPrepareQC()
	if err != nil {
		return
	}

	msg, err := codec.GetCodec().FromAny(msgAny)
	if err != nil {
		return
	}

	quorumCertificate, ok := msg.(*typesCons.QuorumCertificate)
	if !ok {
		return
	}

	m.startNextView(quorumCertificate, false)
}

func (m *pacemaker) NewHeight() {
	defer m.RestartTimer()

	consensusMod := m.GetBus().GetConsensusModule()
	consensusMod.ResetRound(true)
	newHeight := consensusMod.CurrentHeight() + 1
	consensusMod.SetHeight(newHeight)
	m.logger.Info().Uint64("height", newHeight).Msg("🏁 Starting 1st round at new height 🏁")

	// CONSIDERATION: We are omitting CommitQC and TimeoutQC here for simplicity, but should we add them?
	m.startNextView(nil, false)

	m.GetBus().
		GetTelemetryModule().
		GetTimeSeriesAgent().
		CounterIncrement(
			consensusTelemetry.CONSENSUS_BLOCKCHAIN_HEIGHT_COUNTER_NAME,
		)
}

func (m *pacemaker) startNextView(qc *typesCons.QuorumCertificate, forceNextView bool) {
	defer m.RestartTimer()

	// DISCUSS: Should we lock the consensus module here?
	consensusMod := m.GetBus().GetConsensusModule()
	consensusMod.ResetRound(false)
	if err := consensusMod.ReleaseUtilityUnitOfWork(); err != nil {
		m.logger.Error().Err(err).Msg("Failed to release utility unit of work.")
	}
	consensusMod.SetStep(uint8(newRound))

	// TECHDEBT: This if structure for debug purposes only; think of a way to externalize it from the main consensus flow...
	if m.debug.manualMode && !forceNextView {
		m.debug.quorumCertificate = qc
		return
	}

	hotstuffMessage := &typesCons.HotstuffMessage{
		Type:          propose,
		Height:        consensusMod.CurrentHeight(),
		Step:          newRound,
		Round:         consensusMod.CurrentRound(),
		Block:         nil,
		Justification: nil, // Set below if qc is not nil
	}

	if qc != nil {
		hotstuffMessage.Justification = &typesCons.HotstuffMessage_QuorumCertificate{
			QuorumCertificate: qc,
		}
	}

	anyProto, err := anypb.New(hotstuffMessage)
	if err != nil {
		m.logger.Error().Err(err).Fields(m.sharedLoggingFields()).Msgf("Failed to convert pacemaker message to proto.")
		return
	}
	if err := consensusMod.BroadcastMessageToValidators(anyProto); err != nil {
		m.logger.Error().Err(err).Fields(m.sharedLoggingFields()).Msgf("Failed to broadcast message to validators.")
		return
	}
}

// TODO: Increase timeout using exponential backoff.
func (m *pacemaker) getRoundTimeout() time.Duration {
	return time.Duration(int64(time.Millisecond) * int64(m.pacemakerCfg.TimeoutMsec))
}

func (m *pacemaker) sharedLoggingFields() map[string]interface{} {
	consensusMod := m.GetBus().GetConsensusModule()
	return map[string]interface{}{
		"height": consensusMod.CurrentHeight(),
		"step":   typesCons.HotstuffStep(consensusMod.CurrentStep()),
		"round":  consensusMod.CurrentRound(),
	}
}

func pacemakerCatchupLog(height1, step1, round1, height2, step2, round2 uint64) string {
	return fmt.Sprintf("🏃 Pacemaker catching 🏃 up (height, step, round) FROM (%d, %d, %d) TO (%d, %d, %d)", height1, step1, round1, height2, step2, round2)
}
