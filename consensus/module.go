package consensus

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/pokt-network/pocket/consensus/leader_election"
	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/codec"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/test_artifacts"
	"google.golang.org/protobuf/types/known/anypb"

	consensusTelemetry "github.com/pokt-network/pocket/consensus/telemetry"
	"github.com/pokt-network/pocket/shared/modules"
)

const (
	DefaultLogPrefix    = "NODE" // TODO(#164): Make implicit when logging is standardized
	ConsensusModuleName = "consensus"
)

var _ modules.ConsensusGenesisState = &typesCons.ConsensusGenesisState{}
var _ modules.PacemakerConfig = &typesCons.PacemakerConfig{}
var _ modules.ConsensusConfig = &typesCons.ConsensusConfig{}
var _ modules.ConsensusModule = &ConsensusModule{}

// TODO: Do not export the `ConsensusModule` struct or the fields inside of it.
type ConsensusModule struct {
	bus        modules.Bus
	privateKey cryptoPocket.Ed25519PrivateKey

	consCfg     *typesCons.ConsensusConfig
	consGenesis *typesCons.ConsensusGenesisState

	// Hotstuff
	Height uint64
	Round  uint64
	Step   typesCons.HotstuffStep
	Block  *typesCons.Block // The current block being proposed / voted on; it has not been committed to finality

	HighPrepareQC *typesCons.QuorumCertificate // Highest QC for which replica voted PRECOMMIT
	LockedQC      *typesCons.QuorumCertificate // Highest QC for which replica voted COMMIT

	// Leader Election
	LeaderId       *typesCons.NodeId
	NodeId         typesCons.NodeId
	ValAddrToIdMap typesCons.ValAddrToIdMap // TODO: This needs to be updated every time the ValMap is modified
	idToValAddrMap typesCons.IdToValAddrMap // TODO: This needs to be updated every time the ValMap is modified

	// Consensus State
	lastAppHash  string // TODO: Always retrieve this variable from the persistence module and simplify this struct
	validatorMap typesCons.ValidatorMap

	// Module Dependencies
	UtilityContext    modules.UtilityContext
	paceMaker         Pacemaker
	leaderElectionMod leader_election.LeaderElectionModule

	// DEPRECATE: Remove later when we build a shared/proper/injected logger
	logPrefix string

	// TECHDEBT: Move this over to use the txIndexer
	MessagePool map[typesCons.HotstuffStep][]*typesCons.HotstuffMessage
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

		privateKey:  privateKey.(cryptoPocket.Ed25519PrivateKey),
		consCfg:     cfg,
		consGenesis: genesis,

		Height: 0,
		Round:  0,
		Step:   NewRound,
		Block:  nil,

		HighPrepareQC: nil,
		LockedQC:      nil,

		NodeId:         valIdMap[address],
		LeaderId:       nil,
		ValAddrToIdMap: valIdMap,
		idToValAddrMap: idValMap,

		lastAppHash:  "",
		validatorMap: valMap,

		UtilityContext:    nil,
		paceMaker:         paceMaker,
		leaderElectionMod: leaderElectionMod,

		logPrefix:   DefaultLogPrefix,
		MessagePool: make(map[typesCons.HotstuffStep][]*typesCons.HotstuffMessage),
	}

	// TODO(olshansky): Look for a way to avoid doing this.
	paceMaker.SetConsensusModule(m)

	return m, nil
}

func (m *ConsensusModule) InitConfig(pathToConfigJSON string) (config modules.IConfig, err error) {
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

func (m *ConsensusModule) InitGenesis(pathToGenesisJSON string) (genesis modules.IGenesis, err error) {
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

	err = json.Unmarshal(rawJSON[test_artifacts.GetGenesisFileName(m.GetModuleName())], genesis)
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

func (m *ConsensusModule) HandleMessage(message *anypb.Any) error {
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

func (m *ConsensusModule) AppHash() string {
	return m.lastAppHash
}

func (m *ConsensusModule) CurrentHeight() uint64 {
	return m.Height
}

func (m *ConsensusModule) ValidatorMap() modules.ValidatorMap {
	return typesCons.ValidatorMapToModulesValidatorMap(m.validatorMap)
}

// TODO: Populate the entire state from the persistence module: validator set, quorum cert, last block hash, etc...
func (m *ConsensusModule) loadPersistedState() error {
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
