package consensus

import (
	"log"

	"github.com/pokt-network/pocket/consensus/leader_election"
	types_consensus "github.com/pokt-network/pocket/consensus/types"
	pcrypto "github.com/pokt-network/pocket/shared/crypto"

	"github.com/pokt-network/pocket/shared/config"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/types"
)

const (
	DefaultLogPrefix string = "NODE" // Just a default that'll be replaced during consensus operations.
)

var _ modules.ConsensusModule = &consensusModule{}

// TODO(olshansky): Any reason to make all of these attributes local only (i.e. not exposed outside the struct)?
type consensusModule struct {
	bus        modules.Bus
	privateKey pcrypto.Ed25519PrivateKey

	// Hotstuff
	Height uint64
	Round  uint64
	Step   types_consensus.HotstuffStep
	Block  *types_consensus.BlockConsensusTemp // The current block being voted on prior to committing to finality

	HighPrepareQC *types_consensus.QuorumCertificate // Highest QC for which replica voted PRECOMMIT
	LockedQC      *types_consensus.QuorumCertificate // Highest QC for which replica voted COMMIT

	// Leader Election
	LeaderId   *types_consensus.NodeId
	NodeId     types_consensus.NodeId
	ValToIdMap types_consensus.ValToIdMap // TODO(design): This needs to be updated every time the ValMap is modified
	IdToValMap types_consensus.IdToValMap // TODO(design): This needs to be updated every time the ValMap is modified

	// Module Dependencies
	utilityContext    modules.UtilityContext
	paceMaker         PaceMaker
	leaderElectionMod leader_election.LeaderElectionModule

	logPrefix   string                                                              // TODO(design): Remove later when we build a shared/proper/injected logger
	MessagePool map[types_consensus.HotstuffStep][]*types_consensus.HotstuffMessage // TODO(design): Move this over to the persistence module or elsewhere?
}

func Create(cfg *config.Config) (modules.ConsensusModule, error) {
	leaderElectionMod, err := leader_election.Create(cfg)
	if err != nil {
		return nil, err
	}

	// TODO(olshansky): Can we make this a submodule?
	paceMaker, err := CreatePaceMaker(cfg)
	if err != nil {
		return nil, err
	}

	address := cfg.PrivateKey.Address().String()
	valIdMap, idValMap := types_consensus.GetValToIdMap(types.GetTestState(nil).ValidatorMap)

	m := &consensusModule{
		bus:        nil,
		privateKey: cfg.PrivateKey,

		Height: 0,
		Round:  0,
		Step:   NewRound,
		Block:  nil,

		HighPrepareQC: nil,
		LockedQC:      nil,

		NodeId:     valIdMap[address],
		LeaderId:   nil,
		ValToIdMap: valIdMap,
		IdToValMap: idValMap,

		utilityContext:    nil,
		paceMaker:         paceMaker,
		leaderElectionMod: leaderElectionMod,

		logPrefix:   DefaultLogPrefix,
		MessagePool: make(map[types_consensus.HotstuffStep][]*types_consensus.HotstuffMessage),
	}

	// TODO(olshansky): Look for a way to avoid doing this.
	paceMaker.SetConsensusModule(m)

	return m, nil
}

func (m *consensusModule) Start() error {
	log.Println("Starting consensus module")

	if err := m.paceMaker.Start(); err != nil {
		return err
	}

	if err := m.leaderElectionMod.Start(); err != nil {
		return err
	}

	return nil
}

func (m *consensusModule) Stop() error {
	log.Println("Stopping consensus module")
	return nil
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
