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
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
	"google.golang.org/protobuf/types/known/anypb"
)

var _ modules.ConsensusModule = &consensusModule{}

type consensusModule struct {
	base_modules.IntegratableModule

	logger *modules.Logger

	// General configs
	consCfg      *configs.ConsensusConfig
	genesisState *genesis.GenesisState

	// The key used for participating in consensus
	privateKey  cryptoPocket.Ed25519PrivateKey
	nodeAddress string

	// m is a mutex used to control synchronization when multiple goroutines are accessing the struct and its fields / properties.
	// The idea is that you want to acquire a Lock when you are writing values and a RLock when you want to make sure that no other
	// goroutine is changing the values you are trying to read concurrently. Locking context should be the smallest possible but not
	// smaller than a single "unit of work".
	m sync.RWMutex

	// Hotstuff
	height uint64
	round  uint64
	step   typesCons.HotstuffStep
	block  *coreTypes.Block // The current block being proposed / voted on; it has not been committed to finality

	// Stores messages aggregated during a single consensus round from other validators
	hotstuffMempool map[typesCons.HotstuffStep]*hotstuffFIFOMempool

	// Hotstuff safety
	prepareQC *typesCons.QuorumCertificate // Highest QC for which replica voted PRECOMMIT
	lockedQC  *typesCons.QuorumCertificate // Highest QC for which replica voted COMMIT

	// Leader Election
	leaderId *typesCons.NodeId
	nodeId   typesCons.NodeId

	// Module Dependencies
	utilityUnitOfWork modules.UtilityUnitOfWork
	paceMaker         pacemaker.Pacemaker
	leaderElectionMod leader_election.LeaderElectionModule

	// State Sync
	stateSync state_sync.StateSyncModule
}

func Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	return new(consensusModule).Create(bus, options...)
}

func (*consensusModule) Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	leaderElectionMod, err := leader_election.Create(bus)
	if err != nil {
		return nil, err
	}

	paceMakerMod, err := pacemaker.Create(bus)
	if err != nil {
		return nil, err
	}
	pm := paceMakerMod.(pacemaker.Pacemaker)

	stateSyncMod, err := state_sync.Create(bus)
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

		utilityUnitOfWork: nil,
		hotstuffMempool:   make(map[typesCons.HotstuffStep]*hotstuffFIFOMempool),
	}

	for _, option := range options {
		option(m)
	}
	bus.RegisterModule(m)

	runtimeMgr := bus.GetRuntimeMgr()
	m.consCfg = runtimeMgr.GetConfig().Consensus

	genesisState := runtimeMgr.GetGenesis()
	if err := m.ValidateGenesis(genesisState); err != nil {
		return nil, fmt.Errorf("genesis validation failed: %w", err)
	}
	m.genesisState = genesisState

	// TECHDEBT: Should we use the same private key everywhere (top level config, consensus config, etc...) or should we consolidate them?
	privateKey, err := cryptoPocket.NewPrivateKey(m.consCfg.GetPrivateKey())
	if err != nil {
		return nil, err
	}
	m.privateKey = privateKey.(cryptoPocket.Ed25519PrivateKey)

	m.nodeAddress = privateKey.Address().String()
	if m.updateNodeId() != nil {
		return nil, err
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

	if err := m.leaderElectionMod.Start(); err != nil {
		return err
	}

	if err := m.stateSync.Start(); err != nil {
		return err
	}

	return nil
}

func (m *consensusModule) Stop() error {
	m.logger.Info().Msg("Stopping consensus module")
	return nil
}

func (m *consensusModule) GetModuleName() string {
	return modules.ConsensusModuleName
}

func (m *consensusModule) SetBus(pocketBus modules.Bus) {
	m.IntegratableModule.SetBus(pocketBus)
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
		return vals[i].GetServiceUrl() < vals[j].GetServiceUrl()
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

	case messaging.HotstuffMessageContentType:
		msg, err := codec.GetCodec().FromAny(message)
		if err != nil {
			return err
		}
		hotstuffMessage, ok := msg.(*typesCons.HotstuffMessage)
		if !ok {
			return fmt.Errorf("failed to cast message to HotstuffMessage")
		}
		return m.handleHotstuffMessage(hotstuffMessage)

	default:
		return typesCons.ErrUnknownConsensusMessageType(message.MessageName())
	}
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

func (m *consensusModule) GetNodeAddress() string {
	return m.nodeAddress
}

func (m *consensusModule) updateNodeId() error {
	validators, err := m.getValidatorsAtHeight(m.CurrentHeight())
	if err != nil {
		return err
	}
	valAddrToIdMap := typesCons.NewActorMapper(validators).GetValAddrToIdMap()
	m.nodeId = valAddrToIdMap[m.nodeAddress]
	return nil
}

// TODO: Populate the entire state from the persistence module: validator set, quorum cert, last block hash, etc...
func (m *consensusModule) loadPersistedState() error {
	readCtx, err := m.GetBus().GetPersistenceModule().NewReadContext(-1) // Unknown height
	if err != nil {
		return nil
	}
	defer readCtx.Release()

	latestHeight, err := readCtx.GetMaximumBlockHeight()
	if err != nil || latestHeight == 0 {
		// TODO: Proper state sync not implemented yet
		return nil
	}

	m.height = uint64(latestHeight) + 1 // +1 because the height of the consensus module is where it is actively participating in consensus

	m.logger.Info().Uint64("height", m.height).Msg("Starting consensus module")

	return nil
}
