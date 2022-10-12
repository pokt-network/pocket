package consensus

import (
	"fmt"
	"log"
	"sync"

	"github.com/pokt-network/pocket/consensus/leader_election"
	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/codec"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"google.golang.org/protobuf/types/known/anypb"

	consensusTelemetry "github.com/pokt-network/pocket/consensus/telemetry"
	"github.com/pokt-network/pocket/shared/modules"
)

const (
	DefaultLogPrefix    = "NODE" // TODO(#164): Make implicit when logging is standardized
	ConsensusModuleName = "consensus"
)

var (
	_ modules.ConsensusModule       = &consensusModule{}
	_ modules.ConsensusConfig       = &typesCons.ConsensusConfig{}
	_ modules.ConsensusGenesisState = &typesCons.ConsensusGenesisState{}
)

// TODO(#256): Do not export the `ConsensusModule` struct or the fields inside of it.
type consensusModule struct {
	bus        modules.Bus
	privateKey cryptoPocket.Ed25519PrivateKey

	consCfg     *typesCons.ConsensusConfig
	consGenesis *typesCons.ConsensusGenesisState

	// m is a mutex used to control synchronization when multiple goroutines are accessing the struct and its fields / properties.
	//
	// The idea is that you want to acquire a Lock when you are writing values and a RLock when you want to make sure that no other goroutine is changing the values you are trying to read concurrently.
	//
	// Locking context should be the smallest possible but not smaller than a single "unit of work".
	m sync.RWMutex

	// Hotstuff
	Height uint64
	Round  uint64
	Step   typesCons.HotstuffStep
	Block  *typesCons.Block // The current block being proposed / voted on; it has not been committed to finality

	highPrepareQC *typesCons.QuorumCertificate // Highest QC for which replica voted PRECOMMIT
	lockedQC      *typesCons.QuorumCertificate // Highest QC for which replica voted COMMIT

	// Leader Election
	LeaderId       *typesCons.NodeId
	nodeId         typesCons.NodeId
	valAddrToIdMap typesCons.ValAddrToIdMap // TODO: This needs to be updated every time the ValMap is modified
	idToValAddrMap typesCons.IdToValAddrMap // TODO: This needs to be updated every time the ValMap is modified

	// Consensus State
	lastAppHash  string // TODO: Always retrieve this variable from the persistence module and simplify this struct
	validatorMap typesCons.ValidatorMap

	// Module Dependencies
	// TODO(#283): Improve how `utilityContext` is managed
	utilityContext    modules.UtilityContext
	paceMaker         Pacemaker
	leaderElectionMod leader_election.LeaderElectionModule

	// DEPRECATE: Remove later when we build a shared/proper/injected logger
	logPrefix string

	// TECHDEBT: Move this over to use the txIndexer
	messagePool map[typesCons.HotstuffStep][]*typesCons.HotstuffMessage
}

func Create(runtimeMgr modules.RuntimeMgr) (modules.Module, error) {
	return new(consensusModule).Create(runtimeMgr)
}

func (*consensusModule) Create(runtimeMgr modules.RuntimeMgr) (modules.Module, error) {
	var m *consensusModule

	cfg := runtimeMgr.GetConfig()
	if err := m.ValidateConfig(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}
	consensusCfg := cfg.GetConsensusConfig()

	genesis := runtimeMgr.GetGenesis()
	if err := m.ValidateGenesis(genesis); err != nil {
		return nil, fmt.Errorf("genesis validation failed: %w", err)
	}
	consensusGenesis := genesis.GetConsensusGenesisState()

	leaderElectionMod, err := leader_election.Create(runtimeMgr)
	if err != nil {
		return nil, err
	}

	// TODO(olshansky): Can we make this a submodule?
	paceMakerMod, err := CreatePacemaker(runtimeMgr)
	if err != nil {
		return nil, err
	}

	valMap := typesCons.ActorListToValidatorMap(consensusGenesis.GetVals())

	privateKey, err := cryptoPocket.NewPrivateKey(consensusCfg.GetPrivateKey())
	if err != nil {
		return nil, err
	}
	address := privateKey.Address().String()
	valIdMap, idValMap := typesCons.GetValAddrToIdMap(valMap)

	paceMaker := paceMakerMod.(Pacemaker)

	m = &consensusModule{
		bus: nil,

		privateKey:  privateKey.(cryptoPocket.Ed25519PrivateKey),
		consCfg:     cfg.GetConsensusConfig().(*typesCons.ConsensusConfig),
		consGenesis: genesis.GetConsensusGenesisState().(*typesCons.ConsensusGenesisState),

		Height: 0,
		Round:  0,
		Step:   NewRound,
		Block:  nil,

		highPrepareQC: nil,
		lockedQC:      nil,

		nodeId:         valIdMap[address],
		LeaderId:       nil,
		valAddrToIdMap: valIdMap,
		idToValAddrMap: idValMap,

		lastAppHash:  "",
		validatorMap: valMap,

		utilityContext:    nil,
		paceMaker:         paceMaker,
		leaderElectionMod: leaderElectionMod.(leader_election.LeaderElectionModule),

		logPrefix:   DefaultLogPrefix,
		messagePool: make(map[typesCons.HotstuffStep][]*typesCons.HotstuffMessage),
	}

	// TODO(olshansky): Look for a way to avoid doing this.
	paceMaker.SetConsensusModule(m)

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

	if err := m.loadPersistedState(); err != nil {
		return err
	}

	if err := m.paceMaker.Start(); err != nil {
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
	return ConsensusModuleName
}

func (m *consensusModule) GetBus() modules.Bus {
	if m.bus == nil {
		log.Fatalf("PocketBus is not initialized")
	}
	return m.bus
}

func (m *consensusModule) SetBus(pocketBus modules.Bus) {
	m.bus = pocketBus
	m.paceMaker.SetBus(pocketBus)
	m.leaderElectionMod.SetBus(pocketBus)
}

func (*consensusModule) ValidateConfig(cfg modules.Config) error {
	// DISCUSS (team): we cannot cast if we want to use mocks and rely on interfaces
	// if _, ok := cfg.GetConsensusConfig().(*typesCons.ConsensusConfig); !ok {
	// 	return fmt.Errorf("cannot cast to ConsensusConfig")
	// }
	return nil
}

func (*consensusModule) ValidateGenesis(genesis modules.GenesisState) error {
	// DISCUSS (team): we cannot cast if we want to use mocks and rely on interfaces
	// if _, ok := genesis.GetConsensusGenesisState().(*typesCons.ConsensusGenesisState); !ok {
	// 	return fmt.Errorf("cannot cast to ConsensusGenesisState")
	// }
	return nil
}

func (m *consensusModule) GetPrivateKey() (cryptoPocket.PrivateKey, error) {
	return cryptoPocket.NewPrivateKey(m.consCfg.GetPrivateKey())
}

func (m *consensusModule) HandleMessage(message *anypb.Any) error {
	m.m.Lock()
	defer m.m.Unlock()
	switch message.MessageName() {
	case HotstuffMessage:
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
	case UtilityMessage:
		panic("[WARN] UtilityMessage handling is not implemented by consensus yet...")
	default:
		return typesCons.ErrUnknownConsensusMessageType(message.MessageName())
	}

	return nil
}

func (m *consensusModule) AppHash() string {
	return m.lastAppHash
}

func (m *consensusModule) CurrentHeight() uint64 {
	return m.Height
}

func (m *consensusModule) ValidatorMap() modules.ValidatorMap { // TODO: This needs to be dynamically updated during various operations and network changes.
	return typesCons.ValidatorMapToModulesValidatorMap(m.validatorMap)
}

// TODO(#256): Currently only used for testing purposes
func (m *consensusModule) SetUtilityContext(utilityContext modules.UtilityContext) {
	m.utilityContext = utilityContext
}

// TODO: Populate the entire state from the persistence module: validator set, quorum cert, last block hash, etc...
func (m *consensusModule) loadPersistedState() error {
	persistenceContext, err := m.GetBus().GetPersistenceModule().NewReadContext(-1) // Unknown height
	if err != nil {
		return nil
	}
	defer persistenceContext.Close()

	latestHeight, err := persistenceContext.GetLatestBlockHeight()
	if err != nil || latestHeight == 0 {
		m.nodeLog("TODO: State sync not implemented yet")
		return nil
	}

	appHash, err := persistenceContext.GetBlockHash(int64(latestHeight))
	if err != nil {
		return fmt.Errorf("error getting block hash for height %d even though it's in the database: %s", latestHeight, err)
	}

	m.Height = uint64(latestHeight) + 1 // +1 because the height of the consensus module is where it is actively participating in consensus
	m.lastAppHash = string(appHash)

	m.nodeLog(fmt.Sprintf("Starting node at height %d", latestHeight))
	return nil
}
