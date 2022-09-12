package consensus

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/pokt-network/pocket/consensus/leader_election"
	typesCons "github.com/pokt-network/pocket/consensus/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	consensusTelemetry "github.com/pokt-network/pocket/consensus/telemetry"
	"github.com/pokt-network/pocket/shared/modules"
)

const (
	DefaultLogPrefix    string = "NODE" // Just a default that'll be replaced during consensus operations.
	ConsensusModuleName        = "consensus"
	GenesisStatePostfix        = "_genesis_state"
)

var _ modules.ConsensusGenesisState = &typesCons.ConsensusGenesisState{}
var _ modules.PacemakerConfig = &typesCons.PacemakerConfig{}
var _ modules.ConsensusConfig = &typesCons.ConsensusConfig{}
var _ modules.ConsensusModule = &ConsensusModule{}

// TODO(olshansky): Any reason to make all of these attributes local only (i.e. not exposed outside the struct)?
type ConsensusModule struct {
	bus        modules.Bus
	privateKey cryptoPocket.Ed25519PrivateKey
	consCfg    modules.ConsensusConfig

	// Hotstuff
	Height uint64
	Round  uint64
	Step   typesCons.HotstuffStep
	Block  *typesCons.Block // The current block being voted on prior to committing to finality

	HighPrepareQC *typesCons.QuorumCertificate // Highest QC for which replica voted PRECOMMIT
	LockedQC      *typesCons.QuorumCertificate // Highest QC for which replica voted COMMIT

	// Leader Election
	LeaderId       *typesCons.NodeId
	NodeId         typesCons.NodeId
	ValAddrToIdMap typesCons.ValAddrToIdMap // TODO(design): This needs to be updated every time the ValMap is modified
	IdToValAddrMap typesCons.IdToValAddrMap // TODO(design): This needs to be updated every time the ValMap is modified

	// Consensus State
	appHash      string
	validatorMap typesCons.ValidatorMap

	// Module Dependencies
	utilityContext    modules.UtilityContext
	paceMaker         Pacemaker
	leaderElectionMod leader_election.LeaderElectionModule

	logPrefix   string                                                  // TODO(design): Remove later when we build a shared/proper/injected logger
	MessagePool map[typesCons.HotstuffStep][]*typesCons.HotstuffMessage // TODO(design): Move this over to the persistence module or elsewhere?
	// TODO(andrew): Explain (or remove) why have an explicit `MaxBlockBytes` if we are already storing a reference to `consCfg` above?
	// TODO(andrew): This needs to be updated every time the utility module changes this value. It can be accessed via the "application specific bus" (mimicking the intermodule interface in ABCI)
	MaxBlockBytes uint64
}

func Create(configPath, genesisPath string, useRandomPK bool) (modules.ConsensusModule, error) {
	cm := new(ConsensusModule)
	c, err := cm.InitConfig(configPath)
	if err != nil {
		return nil, err
	}
	g, err := cm.InitGenesis(genesisPath)
	if err != nil {
		return nil, err
	}
	cfg := c.(*typesCons.ConsensusConfig)
	genesis := g.(*typesCons.ConsensusGenesisState)
	leaderElectionMod, err := leader_election.Create(cfg, genesis)
	if err != nil {
		return nil, err
	}

	// TODO(olshansky): Can we make this a submodule?
	paceMaker, err := CreatePacemaker(cfg)
	if err != nil {
		return nil, err
	}

	valMap := typesCons.ValidatorListToMap(genesis.Validators)
	var privateKey cryptoPocket.PrivateKey
	if useRandomPK {
		privateKey, err = cryptoPocket.GeneratePrivateKey()
	} else {
		privateKey, err = cryptoPocket.NewPrivateKey(cfg.PrivateKey)
	}
	if err != nil {
		return nil, err
	}
	address := privateKey.Address().String()
	valIdMap, idValMap := typesCons.GetValAddrToIdMap(valMap)

	m := &ConsensusModule{
		bus: nil,

		privateKey: privateKey.(cryptoPocket.Ed25519PrivateKey),
		consCfg:    cfg,

		Height: 0,
		Round:  0,
		Step:   NewRound,
		Block:  nil,

		HighPrepareQC: nil,
		LockedQC:      nil,

		NodeId:         valIdMap[address],
		LeaderId:       nil,
		ValAddrToIdMap: valIdMap,
		IdToValAddrMap: idValMap,

		appHash:      "",
		validatorMap: valMap,

		utilityContext:    nil,
		paceMaker:         paceMaker,
		leaderElectionMod: leaderElectionMod,

		logPrefix:     DefaultLogPrefix,
		MessagePool:   make(map[typesCons.HotstuffStep][]*typesCons.HotstuffMessage),
		MaxBlockBytes: genesis.GetMaxBlockBytes(),
	}

	// TODO(olshansky): Look for a way to avoid doing this.
	paceMaker.SetConsensusModule(m)

	return m, nil
}

func (m *ConsensusModule) InitConfig(pathToConfigJSON string) (config modules.ConfigI, err error) {
	data, err := ioutil.ReadFile(pathToConfigJSON)
	if err != nil {
		return
	}
	// over arching configuration file
	rawJSON := make(map[string]json.RawMessage)
	if err = json.Unmarshal(data, &rawJSON); err != nil {
		log.Fatalf("[ERROR] an error occurred unmarshalling the %s file: %v", pathToConfigJSON, err.Error())
	}
	// consensus specific configuration file
	config = new(typesCons.ConsensusConfig)
	err = json.Unmarshal(rawJSON[m.GetModuleName()], config)
	return
}

func (m *ConsensusModule) InitGenesis(pathToGenesisJSON string) (genesis modules.GenesisI, err error) {
	data, err := ioutil.ReadFile(pathToGenesisJSON)
	if err != nil {
		return
	}
	// over arching configuration file
	rawJSON := make(map[string]json.RawMessage)
	if err = json.Unmarshal(data, &rawJSON); err != nil {
		log.Fatalf("[ERROR] an error occurred unmarshalling the %s file: %v", pathToGenesisJSON, err.Error())
	}
	// consensus specific configuration file
	genesis = new(typesCons.ConsensusGenesisState)
	err = json.Unmarshal(rawJSON[m.GetModuleName()+GenesisStatePostfix], genesis)
	return
}

func (m *ConsensusModule) Start() error {
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

func (m *ConsensusModule) Stop() error {
	return nil
}

func (m *ConsensusModule) GetModuleName() string {
	return ConsensusModuleName
}

func (m *ConsensusModule) GetBus() modules.Bus {
	if m.bus == nil {
		log.Fatalf("PocketBus is not initialized")
	}
	return m.bus
}

func (m *ConsensusModule) SetBus(pocketBus modules.Bus) {
	m.bus = pocketBus
	m.paceMaker.SetBus(pocketBus)
	m.leaderElectionMod.SetBus(pocketBus)
}

func (m *ConsensusModule) loadPersistedState() error {
	persistenceContext, err := m.GetBus().GetPersistenceModule().NewReadContext(-1) // Unknown height
	if err != nil {
		return nil
	}
	defer persistenceContext.Close()

	latestHeight, err := persistenceContext.GetLatestBlockHeight()
	if err != nil || latestHeight == 0 {
		m.nodeLog("TODO: State sync not implement")
		return nil
	}

	appHash, err := persistenceContext.GetBlockHash(int64(latestHeight))
	if err != nil {
		return fmt.Errorf("error getting block hash for height %d even though it's in the database: %s", latestHeight, err)
	}

	// TODO: Populate the rest of the state from the persistence module: validator set, quorum cert, last block hash, etc...
	m.Height = uint64(latestHeight) + 1 // +1 because the height of the consensus module is where it is actively participating in consensus
	m.appHash = string(appHash)

	m.nodeLog(fmt.Sprintf("Starting node at height %d", latestHeight))
	return nil
}

// TODO(discuss): Low priority design: think of a way to make `hotstuff_*` files be a sub-package under consensus.
// This is currently not possible because functions tied to the `ConsensusModule`
// struct (implementing the ConsensusModule module), which spans multiple files.
/*
TODO(discuss): The reason we do not assign both the leader and the replica handlers
to the leader (which should also act as a replica when it is a leader) is because it
can create a weird inconsistent state (e.g. both the replica and leader try to restart
the Pacemaker timeout). This requires additional "replica-like" logic in the leader
handler which has both pros and cons:
	Pros:
		* The leader can short-circuit and optimize replica related logic
		* Avoids additional code flowing through the P2P pipeline
		* Allows for micro-optimizations
	Cons:
		* The leader's "replica related logic" requires an additional code path
		* Code is less "generalizable" and therefore potentially more error prone
*/

// TODO(olshansky): Should we just make these singletons or embed them directly in the ConsensusModule?
type HotstuffMessageHandler interface {
	HandleNewRoundMessage(*ConsensusModule, *typesCons.HotstuffMessage)
	HandlePrepareMessage(*ConsensusModule, *typesCons.HotstuffMessage)
	HandlePrecommitMessage(*ConsensusModule, *typesCons.HotstuffMessage)
	HandleCommitMessage(*ConsensusModule, *typesCons.HotstuffMessage)
	HandleDecideMessage(*ConsensusModule, *typesCons.HotstuffMessage)
}

func (m *ConsensusModule) HandleMessage(message *anypb.Any) error {
	switch message.MessageName() {
	case HotstuffMessage:
		var hotstuffMessage typesCons.HotstuffMessage
		err := anypb.UnmarshalTo(message, &hotstuffMessage, proto.UnmarshalOptions{})
		if err != nil {
			return err
		}
		m.handleHotstuffMessage(&hotstuffMessage)
	case UtilityMessage:
		panic("[WARN] UtilityMessage handling is not implemented by consensus yet...")
	default:
		return typesCons.ErrUnknownConsensusMessageType(message.MessageName())
	}

	return nil
}

func (m *ConsensusModule) handleHotstuffMessage(msg *typesCons.HotstuffMessage) {
	m.nodeLog(typesCons.DebugHandlingHotstuffMessage(msg))

	// Liveness & safety checks
	if err := m.paceMaker.ValidateMessage(msg); err != nil {
		// If a replica is not a leader for this round, but has already determined a leader,
		// and continues to receive NewRound messages, we avoid logging the "message discard"
		// because it creates unnecessary spam.
		if !(m.LeaderId != nil && !m.isLeader() && msg.Step == NewRound) {
			m.nodeLog(typesCons.WarnDiscardHotstuffMessage(msg, err.Error()))
		}
		return
	}

	// Need to execute leader election if there is no leader and we are in a new round.
	if m.Step == NewRound && m.LeaderId == nil {
		m.electNextLeader(msg)
	}

	if m.isReplica() {
		replicaHandlers[msg.Step](m, msg)
		return
	}

	// Note that the leader also acts as a replica, but this logic is implemented in the underlying code.
	leaderHandlers[msg.Step](m, msg)
}

func (m *ConsensusModule) AppHash() string {
	return m.appHash
}

func (m *ConsensusModule) CurrentHeight() uint64 {
	return m.Height
}

func (m *ConsensusModule) ValidatorMap() modules.ValidatorMap {
	return typesCons.ValidatorMapToModulesValidatorMap(m.validatorMap)
}
