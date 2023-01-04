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
	DefaultLogPrefix = "NODE" // TODO(#164): Make implicit when logging is standardized
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
	// TODO(#315): Move the statefulness of `TxResult` to the persistence module
	TxResults []modules.TxResult // The current block applied transaction results / voted on; it has not been committed to finality

	prepareQC *typesCons.QuorumCertificate // Highest QC for which replica voted PRECOMMIT
	lockedQC  *typesCons.QuorumCertificate // Highest QC for which replica voted COMMIT

	// Leader Election
	leaderId *typesCons.NodeId
	nodeId   typesCons.NodeId

	// Module Dependencies
	// IMPROVE(#283): Investigate whether the current approach to how the `utilityContext` should be
	//                managed or changed. Also consider exposing a function that exposes the context
	//                to streamline how its accessed in the module (see the ticket).

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
	SetStep(uint64)
	SetBlock(*typesCons.Block)
	SetLeaderId(*typesCons.NodeId)
	SetUtilityContext(modules.UtilityContext)
}

func (m *consensusModule) SetHeight(height uint64) {
	m.height = height
}

func (m *consensusModule) SetRound(round uint64) {
	m.round = round
}

func (m *consensusModule) SetStep(step uint64) {
	m.step = typesCons.HotstuffStep(step)
}

func (m *consensusModule) SetBlock(block *typesCons.Block) {
	m.block = block
}

func (m *consensusModule) SetLeaderId(leaderId *typesCons.NodeId) {
	m.leaderId = leaderId
}

func (m *consensusModule) SetUtilityContext(utilityContext modules.UtilityContext) {
	m.utilityContext = utilityContext
}

// implementations of the type PaceMakerAccessModule interface
// SetHeight, SeetRound, SetStep are implemented for ConsensusDebugModule

func (m *consensusModule) ClearLeaderMessagesPool() {
	m.clearLeader()
	m.clearMessagesPool()
}

func (m *consensusModule) ResetForNewHeight() {
	m.resetForNewHeight()
}

func (m *consensusModule) ReleaseUtilityContext() error {
	if m.utilityContext != nil {
		if err := m.utilityContext.Release(); err != nil {
			log.Println("[WARN] Failed to release utility context: ", err)
			return err
		}
		m.utilityContext = nil
	}

	return nil
}

func (m *consensusModule) BroadcastMessageToNodes(msg *anypb.Any) error {
	msgCodec, err := codec.GetCodec().FromAny(msg)
	if err != nil {
		return err
	}

	broadcastMessage, ok := msgCodec.(*typesCons.HotstuffMessage)
	if !ok {
		return fmt.Errorf("failed to cast message to HotstuffMessage")
	}
	m.broadcastToNodes(broadcastMessage)

	return nil
}

func (m *consensusModule) IsLeader() bool {
	return m.isLeader()
}

func (m *consensusModule) IsLeaderSet() bool {
	// if m.leaderId == nil {
	// 	return false
	// }
	return m.leaderId != nil
}

func (m *consensusModule) ElectNextLeader(msg *anypb.Any) error {
	msgCodec, err := codec.GetCodec().FromAny(msg)
	if err != nil {
		return err
	}

	message, ok := msgCodec.(*typesCons.HotstuffMessage)
	if !ok {
		return fmt.Errorf("failed to cast message to HotstuffMessage")
	}

	return m.electNextLeader(message)
}

func (m *consensusModule) GetPrepareQC() *anypb.Any {
	anyProto, err := anypb.New(m.prepareQC)
	if err != nil {
		log.Println("[WARN] NewHeight: Failed to convert paceMaker message to proto: ", err)
		return nil
	}
	return anyProto

}

func Create(bus modules.Bus) (modules.Module, error) {
	return new(consensusModule).Create(bus)
}

func (*consensusModule) Create(bus modules.Bus) (modules.Module, error) {
	leaderElectionMod, err := leader_election.Create(bus)
	if err != nil {
		return nil, err
	}

	// TODO(olshansky): Can we make this a submodule?
	paceMakerMod, err := CreatePacemaker(bus)
	if err != nil {
		return nil, err
	}

	pacemaker := paceMakerMod.(Pacemaker)
	m := &consensusModule{
		paceMaker:         pacemaker,
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

		messagePool: make(map[typesCons.HotstuffStep][]*typesCons.HotstuffMessage),
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

	// TODO(olshansky): Look for a way to avoid doing this.
	// TODO(goku): remove tight connection of pacemaker and consensus.
	//paceMaker.SetConsensusModule(m)

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
	return modules.ConsensusModuleName
}

func (m *consensusModule) GetBus() modules.Bus {
	if m.bus == nil {
		log.Fatalf("PocketBus is not initialized")
	}
	return m.bus
}

func (m *consensusModule) SetBus(pocketBus modules.Bus) {
	m.bus = pocketBus
	if m.paceMaker != nil {
		m.paceMaker.SetBus(pocketBus)
	}
	if m.leaderElectionMod != nil {
		m.leaderElectionMod.SetBus(pocketBus)
	}
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

// TODO: Populate the entire state from the persistence module: validator set, quorum cert, last block hash, etc...
func (m *consensusModule) loadPersistedState() error {
	persistenceContext, err := m.GetBus().GetPersistenceModule().NewReadContext(-1) // Unknown height
	if err != nil {
		return nil
	}
	defer persistenceContext.Close()

	latestHeight, err := persistenceContext.GetLatestBlockHeight()
	if err != nil || latestHeight == 0 {
		// TODO: Proper state sync not implemented yet
		return nil
	}

	m.height = uint64(latestHeight) + 1 // +1 because the height of the consensus module is where it is actively participating in consensus

	m.nodeLog(fmt.Sprintf("Starting consensus module at height %d", latestHeight))

	return nil
}
