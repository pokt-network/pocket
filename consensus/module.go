package consensus

import (
	"fmt"
	"log"
	"sort"
	"sync"

	"github.com/pokt-network/pocket/consensus/leader_election"
	consensusTelemetry "github.com/pokt-network/pocket/consensus/telemetry"
	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/runtime/genesis"
	"github.com/pokt-network/pocket/shared/codec"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	DefaultLogPrefix    = "NODE" // TODO(#164): Make implicit when logging is standardized
	consensusModuleName = "consensus"
)

var (
	_ modules.ConsensusModule = &consensusModule{}
	_ ConsensusDebugModule    = &consensusModule{}
)

type consensusModule struct {
	bus        modules.Bus
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
	block  *typesCons.Block // The current block being proposed / voted on; it has not been committed to finality
	//    TODO(#315):  Move the statefulness of `TxResult` to the persistence module
	TxResults []modules.TxResult // The current block applied transaction results / voted on; it has not been committed to finality

	// IMPROVE: Consider renaming `highPrepareQC` to simply `prepareQC`
	highPrepareQC *typesCons.QuorumCertificate // Highest QC for which replica voted PRECOMMIT
	lockedQC      *typesCons.QuorumCertificate // Highest QC for which replica voted COMMIT

	// Leader Election
	leaderId       *typesCons.NodeId
	nodeId         typesCons.NodeId
	valAddrToIdMap typesCons.ValAddrToIdMap // TODO: This needs to be updated every time the ValMap is modified
	idToValAddrMap typesCons.IdToValAddrMap // TODO: This needs to be updated every time the ValMap is modified

	// Consensus State
	validatorMap typesCons.ValidatorMap

	// Module Dependencies
	// TODO(#283): Improve how `utilityContext` is managed
	utilityContext    modules.UtilityContext
	paceMaker         Pacemaker
	leaderElectionMod leader_election.LeaderElectionModule

	// DEPRECATE: Remove later when we build a shared/proper/injected logger
	logPrefix string

	// TECHDEBT: Rename this to `consensusMessagePool` or something similar
	//           and reconsider if an in-memory map is the best approach
	messagePool map[typesCons.HotstuffStep][]*typesCons.HotstuffMessage
}

// Functions exposed by the debug interface should only be used for testing puposes.
type ConsensusDebugModule interface {
	SetHeight(uint64)
	SetRound(uint64)
	SetStep(typesCons.HotstuffStep)
	SetBlock(*typesCons.Block)
	SetLeaderId(*typesCons.NodeId)
	SetUtilityContext(modules.UtilityContext)
}

func (c *consensusModule) SetHeight(height uint64) {
	c.height = height
}

func (c *consensusModule) SetRound(round uint64) {
	c.round = round
}

func (c *consensusModule) SetStep(step typesCons.HotstuffStep) {
	c.step = step
}

func (c *consensusModule) SetBlock(block *typesCons.Block) {
	c.block = block
}

func (c *consensusModule) SetLeaderId(leaderId *typesCons.NodeId) {
	c.leaderId = leaderId
}

func (c *consensusModule) SetUtilityContext(utilityContext modules.UtilityContext) {
	c.utilityContext = utilityContext
}

func Create(runtimeMgr modules.RuntimeMgr) (modules.Module, error) {
	return new(consensusModule).Create(runtimeMgr)
}

func (*consensusModule) Create(runtimeMgr modules.RuntimeMgr) (modules.Module, error) {
	var m *consensusModule

	consensusCfg := runtimeMgr.GetConfig().Consensus

	genesisState := runtimeMgr.GetGenesis()
	if err := m.ValidateGenesis(genesisState); err != nil {
		return nil, fmt.Errorf("genesis validation failed: %w", err)
	}

	leaderElectionMod, err := leader_election.Create(runtimeMgr)
	if err != nil {
		return nil, err
	}

	// TODO(olshansky): Can we make this a submodule?
	paceMakerMod, err := CreatePacemaker(runtimeMgr)
	if err != nil {
		return nil, err
	}

	valMap := typesCons.ActorListToValidatorMap(genesisState.GetValidators())

	privateKey, err := cryptoPocket.NewPrivateKey(consensusCfg.PrivateKey)
	if err != nil {
		return nil, err
	}
	address := privateKey.Address().String()
	valIdMap, idValMap := typesCons.GetValAddrToIdMap(valMap)

	paceMaker := paceMakerMod.(Pacemaker)

	m = &consensusModule{
		bus: nil,

		privateKey:   privateKey.(cryptoPocket.Ed25519PrivateKey),
		consCfg:      consensusCfg,
		genesisState: genesisState,

		height: 0,
		round:  0,
		step:   NewRound,
		block:  nil,

		highPrepareQC: nil,
		lockedQC:      nil,

		nodeId:         valIdMap[address],
		leaderId:       nil,
		valAddrToIdMap: valIdMap,
		idToValAddrMap: idValMap,

		validatorMap: valMap,

		utilityContext:    nil,
		paceMaker:         paceMaker,
		leaderElectionMod: leaderElectionMod.(leader_election.LeaderElectionModule),

		logPrefix:   DefaultLogPrefix,
		messagePool: make(map[typesCons.HotstuffStep][]*typesCons.HotstuffMessage),
	}

	// TODO(olshansky): Look for a way to avoid doing this.
	// TODO(goku): remove tight connection of pacemaker and consensus.
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
	return consensusModuleName
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

func (*consensusModule) ValidateGenesis(genesis *genesis.GenesisState) error {
	// Sort the validators by their generic param (i.e. service URL)
	vals := genesis.GetValidators()
	sort.Slice(vals, func(i, j int) bool {
		return vals[i].GetGenericParam() < vals[j].GetGenericParam()
	})

	// Sort the validators by their address
	vals2 := vals[:]
	sort.Slice(vals, func(i, j int) bool {
		return vals[i].GetAddress() < vals[j].GetAddress()
	})

	for i := 0; i < len(vals); i++ {
		if vals[i].GetAddress() != vals2[i].GetAddress() {
			// There is an implicit dependency because of how RainTree works and how the validator map
			// is currently managed to make sure that the ordering of the address and the service URL
			// are the same. This will be addressed once the # of validators will scale.
			panic("HACK(olshansky): service url and address must be sorted the same way")
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

func (m *consensusModule) ValidatorMap() modules.ValidatorMap { // TODO: This needs to be dynamically updated during various operations and network changes.
	return typesCons.ValidatorMapToModulesValidatorMap(m.validatorMap)
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

	m.height = uint64(latestHeight) + 1 // +1 because the height of the consensus module is where it is actively participating in consensus

	m.nodeLog(fmt.Sprintf("Starting node at height %d", latestHeight))

	return nil
}
