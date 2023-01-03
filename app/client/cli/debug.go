//go:build debug
// +build debug

package cli

import (
	"os"
	"sync"

	"github.com/manifoldco/promptui"
	"github.com/pokt-network/pocket/consensus"
	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/p2p"
	"github.com/pokt-network/pocket/rpc"
	"github.com/pokt-network/pocket/runtime"
	"github.com/pokt-network/pocket/shared"
	pocketCrypto "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/telemetry"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/anypb"
)

// TECHDEBT: Lowercase variables / constants that do not need to be exported.
const (
	PromptResetToGenesis         string = "ResetToGenesis"
	PromptPrintNodeState         string = "PrintNodeState"
	PromptTriggerNextView        string = "TriggerNextView"
	PromptTogglePacemakerMode    string = "TogglePacemakerMode"
	PromptShowLatestBlockInStore string = "ShowLatestBlockInStore"

	defaultConfigPath  = "build/config/config1.json"
	defaultGenesisPath = "build/config/genesis.json"
)

var (
	// A P2P module is initialized in order to broadcast a message to the local network
	p2pMod modules.P2PModule

	// A consensus module is initialized in order to get a list of the validator network
	consensusMod modules.ConsensusModule
	modInitOnce  sync.Once

	items = []string{
		PromptResetToGenesis,
		PromptPrintNodeState,
		PromptTriggerNextView,
		PromptTogglePacemakerMode,
		PromptShowLatestBlockInStore,
	}
)

func init() {
	rootCmd.AddCommand(NewDebugCommand())
}

func NewDebugCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "debug",
		Short: "Debug utility for rapid development",
		Args:  cobra.ExactArgs(0),
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			initDebug(remoteCLIURL)
		},
		RunE: runDebug,
	}
}

func runDebug(cmd *cobra.Command, args []string) (err error) {
	for {
		if selection, err := promptGetInput(); err == nil {
			handleSelect(selection)
		} else {
			return err
		}
	}
}

func promptGetInput() (string, error) {
	prompt := promptui.Select{
		Label: "Select an action",
		Items: items,
		Size:  len(items),
	}

	_, result, err := prompt.Run()

	if err == promptui.ErrInterrupt {
		os.Exit(0)
	}

	if err != nil {
		logger.Global.Logger.Error().Err(err).Msg("Prompt failed")
		return "", err
	}

	return result, nil
}

func handleSelect(selection string) {
	switch selection {
	case PromptResetToGenesis:
		m := &messaging.DebugMessage{
			Action:  messaging.DebugMessageAction_DEBUG_CONSENSUS_RESET_TO_GENESIS,
			Message: nil,
		}
		broadcastDebugMessage(m)
	case PromptPrintNodeState:
		m := &messaging.DebugMessage{
			Action:  messaging.DebugMessageAction_DEBUG_CONSENSUS_PRINT_NODE_STATE,
			Message: nil,
		}
		broadcastDebugMessage(m)
	case PromptTriggerNextView:
		m := &messaging.DebugMessage{
			Action:  messaging.DebugMessageAction_DEBUG_CONSENSUS_TRIGGER_NEXT_VIEW,
			Message: nil,
		}
		broadcastDebugMessage(m)
	case PromptTogglePacemakerMode:
		m := &messaging.DebugMessage{
			Action:  messaging.DebugMessageAction_DEBUG_CONSENSUS_TOGGLE_PACE_MAKER_MODE,
			Message: nil,
		}
		broadcastDebugMessage(m)
	case PromptShowLatestBlockInStore:
		m := &messaging.DebugMessage{
			Action:  messaging.DebugMessageAction_DEBUG_SHOW_LATEST_BLOCK_IN_STORE,
			Message: nil,
		}
		sendDebugMessage(m)
	default:
		logger.Global.Logger.Error().Msg("Selection not yet implemented...")
	}
}

// Broadcast to the entire validator set
func broadcastDebugMessage(debugMsg *messaging.DebugMessage) {
	anyProto, err := anypb.New(debugMsg)
	if err != nil {
		logger.Global.Logger.Error().Err(err).Msg("Failed to create Any proto")
	}

	// TODO(olshansky): Once we implement the cleanup layer in RainTree, we'll be able to use
	// broadcast. The reason it cannot be done right now is because this client is not in the
	// address book of the actual validator nodes, so `node1.consensus` never receives the message.
	// p2pMod.Broadcast(anyProto, messaging.PocketTopic_DEBUG_TOPIC)

	for _, val := range consensusMod.ValidatorMap() {
		addr, err := pocketCrypto.NewAddress(val.GetAddress())
		if err != nil {
			logger.Global.Logger.Fatal().Err(err).Msg("Failed to convert validator address into pocketCrypto.Address")
		}
		p2pMod.Send(addr, anyProto)
	}
}

// Send to just a single (i.e. first) validator in the set
func sendDebugMessage(debugMsg *messaging.DebugMessage) {
	anyProto, err := anypb.New(debugMsg)
	if err != nil {
		logger.Global.Logger.Error().Err(err).Msg("Failed to create Any proto")
	}

	var validatorAddress []byte
	for _, val := range consensusMod.ValidatorMap() {
		validatorAddress, err = pocketCrypto.NewAddress(val.GetAddress())
		if err != nil {
			logger.Global.Logger.Fatal().Err(err).Msg("Failed to convert validator address into pocketCrypto.Address")
		}
		break
	}

	p2pMod.Send(validatorAddress, anyProto)
}

func initDebug(remoteCLIURL string) {
	modInitOnce.Do(func() {
		var err error
		runtimeMgr := runtime.NewManagerFromFiles(defaultConfigPath, defaultGenesisPath, runtime.WithRandomPK())

		loggerM, err := logger.Create(runtimeMgr)
		if err != nil {
			logger.Global.Fatal().Err(err).Msg("Failed to create logger module")
		}
		loggerMod := loggerM.(modules.LoggerModule)

		consM, err := consensus.Create(runtimeMgr)
		if err != nil {
			logger.Global.Fatal().Err(err).Msg("Failed to create consensus module")
		}
		consensusMod = consM.(modules.ConsensusModule)

		p2pM, err := p2p.Create(runtimeMgr)
		if err != nil {
			logger.Global.Fatal().Err(err).Msg("Failed to create p2p module")
		}
		p2pMod = p2pM.(modules.P2PModule)

		// This telemetry module instance is a NOOP because the 'enable_telemetry' flag in the `cfg` above is set to false.
		// Since this client mimics partial - networking only - functionality of a full node, some of the telemetry-related
		// code paths are executed. To avoid those messages interfering with the telemetry data collected, a non-nil telemetry
		// module that NOOPs (per the configs above) is injected.
		telemetryM, err := telemetry.Create(runtimeMgr)
		if err != nil {
			logger.Global.Fatal().Err(err).Msg("Failed to create telemetry module")
		}
		telemetryMod := telemetryM.(modules.TelemetryModule)

		rpcM, err := rpc.Create(runtimeMgr)
		if err != nil {
			logger.Global.Fatal().Err(err).Msg("Failed to create rpc module")
		}
		rpcMod := rpcM.(modules.RPCModule)

		_ = shared.CreateBusWithOptionalModules(runtimeMgr, nil, p2pMod, nil, consensusMod, telemetryMod, loggerMod, rpcMod) // REFACTOR: use the `WithXXXModule()` pattern accepting a slice of IntegratableModule

		p2pMod.Start()
	})
}
