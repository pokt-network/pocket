package consensus

import (
	"fmt"
	"sort"
	"sync"

	"github.com/pokt-network/pocket/consensus/leader_election"
	"github.com/pokt-network/pocket/consensus/pacemaker"
	"github.com/pokt-network/pocket/consensus/state_sync"
	consensusTelemetry "github.com/pokt-network/pocket/consensus/telemetry"
	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/runtime/genesis"
	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	DefaultLogPrefix = "NODE" // TODO(#164): Make implicit when logging is standardized
)

var (
	_ modules.ConsensusModule = &consensusModule{}
	_ ConsensusDebugModule    = &consensusModule{}
)

type consensusModule struct {
	modules.BaseIntegratableModule

	privateKey cryptoPocket.Ed25519PrivateKey

	consCfg      *configs.ConsensusConfig
	genesisState *genesis.GenesisState

	// m is a mutex used to control synchronization when multiple goroutines are accessing the struct and its fields / properties.
	//
	// The idea is that you want to acquire a Lock when you are writing values and a RLock when you want to make sure that no other goroutine is changing the values you are trying to read concurrently.
	//
	// Locking context should be the smallest possible but not smaller than a single "unit of work".
	m sync.RWMutex

	// Hotstuff
	height uint64
	round  uint64
	step   typesCons.HotstuffStep
	block  *coreTypes.Block // The current block being proposed / voted on; it has not been committed to finality
	// TODO(#315): Move the statefulness of `TxResult` to the persistence module
	TxResults []modules.TxResult // The current block applied transaction results / voted on; it has not been committed to finality

	prepareQC *typesCons.QuorumCertificate // Highest QC for which replica voted PRECOMMIT
	lockedQC  *typesCons.QuorumCertificate // Highest QC for which replica voted COMMIT

	// Leader Election
	leaderId *typesCons.NodeId
	nodeId   typesCons.NodeId

	nodeAddress string

	// Module Dependencies
	// IMPROVE(#283): Investigate whether the current approach to how the `utilityContext` should be
	//                managed or changed. Also consider exposing a function that exposes the context
	//                to streamline how its accessed in the module (see the ticket).

	utilityContext    modules.UtilityContext
	paceMaker         pacemaker.Pacemaker
	leaderElectionMod leader_election.LeaderElectionModule

	logger    modules.Logger
	logPrefix string

	stateSync state_sync.StateSyncModule

	hotstuffMempool map[typesCons.HotstuffStep]*hotstuffFIFOMempool
}

// Functions exposed by the debug interface should only be used for testing puposes.
type ConsensusDebugModule interface {
	SetHeight(uint64)
	SetRound(uint64)
	// REFACTOR: This should accept typesCons.HotstuffStep.
	SetStep(uint8)
	SetBlock(*coreTypes.Block)
	SetLeaderId(*typesCons.NodeId)
	SetUtilityContext(modules.UtilityContext)
}

func (m *consensusModule) SetHeight(height uint64) {
	m.height = height
	m.publishNewHeightEvent(height)
}

func (m *consensusModule) SetRound(round uint64) {
	m.round = round
}

func (m *consensusModule) SetStep(step uint8) {
	m.step = typesCons.HotstuffStep(step)
}

func (m *consensusModule) SetBlock(block *coreTypes.Block) {
	m.block = block
}

func (m *consensusModule) SetLeaderId(leaderId *typesCons.NodeId) {
	m.leaderId = leaderId
}

func (m *consensusModule) SetUtilityContext(utilityContext modules.UtilityContext) {
	m.utilityContext = utilityContext
}

// Implementations of the ConsensusStateSync interface

func (m *consensusModule) GetNodeIdFromNodeAddress(peerId string) (uint64, error) {
	validators, err := m.getValidatorsAtHeight(m.CurrentHeight())
	if err != nil {
		// REFACTOR(#434): As per issue #434, once the new id is sorted out, this return statement must be changed
		return 0, err
	}

	valAddrToIdMap := typesCons.NewActorMapper(validators).GetValAddrToIdMap()
	return uint64(valAddrToIdMap[peerId]), nil
}

func (m *consensusModule) GetNodeAddress() string {
	return m.nodeAddress
}

// Implementations of the type PaceMakerAccessModule interface
// SetHeight, SeetRound, SetStep are implemented for ConsensusDebugModule
func (m *consensusModule) ClearLeaderMessagesPool() {
	m.clearLeader()
	m.clearMessagesPool()
}

func Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	return new(consensusModule).Create(bus, options...)
}

func (*consensusModule) Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	leaderElectionMod, err := leader_election.Create(bus)
	if err != nil {
		return nil, err
	}

	paceMakerMod, err := pacemaker.CreatePacemaker(bus)
	if err != nil {
		return nil, err
	}
	pm := paceMakerMod.(pacemaker.Pacemaker)

	stateSyncMod, err := state_sync.CreateStateSync(bus)
	if err != nil {
		return nil, err
	}
	stateSync := stateSyncMod.(state_sync.StateSyncModule)

	m := &consensusModule{
		paceMaker:         pm,
		stateSync:         stateSync,
		leaderElectionMod: leaderElectionMod.(leader_election.LeaderElectionModule),

		height: 0,
		round:  0,
		step:   NewRound,
		block:  nil,

		prepareQC: nil,
		lockedQC:  nil,

		leaderId: nil,

		utilityContext: nil,

		logPrefix: DefaultLogPrefix,

		hotstuffMempool: make(map[typesCons.HotstuffStep]*hotstuffFIFOMempool),
	}
	bus.RegisterModule(m)

	runtimeMgr := bus.GetRuntimeMgr()

	consensusCfg := runtimeMgr.GetConfig().Consensus

	genesisState := runtimeMgr.GetGenesis()
	if err := m.ValidateGenesis(genesisState); err != nil {
		return nil, fmt.Errorf("genesis validation failed: %w", err)
	}

	privateKey, err := cryptoPocket.NewPrivateKey(consensusCfg.GetPrivateKey())
	if err != nil {
		return nil, err
	}
	address := privateKey.Address().String()

	validators, err := m.getValidatorsAtHeight(m.CurrentHeight())
	if err != nil {
		return nil, err
	}

	valAddrToIdMap := typesCons.NewActorMapper(validators).GetValAddrToIdMap()

	m.privateKey = privateKey.(cryptoPocket.Ed25519PrivateKey)
	m.consCfg = consensusCfg
	m.genesisState = genesisState

	m.nodeId = valAddrToIdMap[address]
	m.nodeAddress = address

	if consensusCfg.ServerModeEnabled {
		m.stateSync.EnableServerMode()
	}

	m.initMessagesPool()

	return m, nil
}

func (m *consensusModule) Start() error {
	m.GetBus().
		GetTelemetryModule().
		GetTimeSeriesAgent().
		CounterRegister(
			consensusTelemetry.CONSENSUS_BLOCKCHAIN_HEIGHT_COUNTER_NAME,
			consensusTelemetry.CONSENSUS_BLOCKCHAIN_HEIGHT_COUNTER_DESCRIPTION,
		)

	m.logger = logger.Global.CreateLoggerForModule(m.GetModuleName())

	if err := m.loadPersistedState(); err != nil {
		return err
	}

	if err := m.paceMaker.Start(); err != nil {
		return err
	}

	if err := m.stateSync.Start(); err != nil {
		return err
	}

	if err := m.leaderElectionMod.Start(); err != nil {
		return err
	}

	return nil
}

func (m *consensusModule) Stop() error {
	return nil
}

func (m *consensusModule) GetModuleName() string {
	return modules.ConsensusModuleName
}

func (m *consensusModule) GetBus() modules.Bus {
	bus := m.BaseIntegratableModule.GetBus()
	if bus == nil {
		logger.Global.Fatal().Msg("PocketBus is not initialized")
	}
	return bus
}

func (m *consensusModule) SetBus(pocketBus modules.Bus) {
	m.BaseIntegratableModule.SetBus(pocketBus)
	if m.paceMaker != nil {
		m.paceMaker.SetBus(pocketBus)
	}
	if m.leaderElectionMod != nil {
		m.leaderElectionMod.SetBus(pocketBus)
	}
}

func (*consensusModule) ValidateGenesis(gen *genesis.GenesisState) error {
	// Sort the validators by their generic param (i.e. service URL)
	vals := gen.GetValidators()
	sort.Slice(vals, func(i, j int) bool {
		return vals[i].GetGenericParam() < vals[j].GetGenericParam()
	})

	// Sort the validators by their address
	vals2 := vals[:] // nolint:gocritic // Make a copy of the slice to retain order
	sort.Slice(vals, func(i, j int) bool {
		return vals[i].GetAddress() < vals[j].GetAddress()
	})

	for i := 0; i < len(vals); i++ {
		if vals[i].GetAddress() != vals2[i].GetAddress() {
			// There is an implicit dependency because of how RainTree works and how the validator map
			// is currently managed to make sure that the ordering of the address and the service URL
			// are the same. This will be addressed once the # of validators will scale.
			logger.Global.Panic().Msg("HACK(olshansky): service url and address must be sorted the same way")
		}
	}

	return nil
}

func (m *consensusModule) GetPrivateKey() (cryptoPocket.PrivateKey, error) {
	return cryptoPocket.NewPrivateKey(m.consCfg.PrivateKey)
}

func (m *consensusModule) HandleMessage(message *anypb.Any) error {
	m.m.Lock()
	defer m.m.Unlock()

	switch message.MessageName() {
	case HotstuffMessageContentType:
		msg, err := codec.GetCodec().FromAny(message)
		if err != nil {
			return err
		}
		hotstuffMessage, ok := msg.(*typesCons.HotstuffMessage)
		if !ok {
			return fmt.Errorf("failed to cast message to HotstuffMessage")
		}
		if err := m.handleHotstuffMessage(hotstuffMessage); err != nil {
			return err
		}
	default:
		return typesCons.ErrUnknownConsensusMessageType(message.MessageName())
	}

	return nil
}

func (m *consensusModule) CurrentHeight() uint64 {
	return m.height
}

func (m *consensusModule) CurrentRound() uint64 {
	return m.round
}

func (m *consensusModule) CurrentStep() uint64 {
	return uint64(m.step)
}

// TODO: Populate the entire state from the persistence module: validator set, quorum cert, last block hash, etc...
func (m *consensusModule) loadPersistedState() error {
	persistenceContext, err := m.GetBus().GetPersistenceModule().NewReadContext(-1) // Unknown height
	if err != nil {
		return nil
	}
	defer persistenceContext.Close()

	latestHeight, err := persistenceContext.GetMaximumBlockHeight()
	if err != nil || latestHeight == 0 {
		// TODO: Proper state sync not implemented yet
		return nil
	}

	m.height = uint64(latestHeight) + 1 // +1 because the height of the consensus module is where it is actively participating in consensus

	m.logger.Info().Uint64("height", m.height).Msg("Starting consensus module")

	return nil
}

func (m *consensusModule) EnableServerMode() {
	m.stateSync.EnableServerMode()
}
