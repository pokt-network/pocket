package consensus

import (
	"fmt"
	"log"

	"github.com/pokt-network/pocket/consensus/leader_election"
	types_consensus "github.com/pokt-network/pocket/consensus/types"
	pcrypto "github.com/pokt-network/pocket/shared/crypto"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

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
	LeaderId       *types_consensus.NodeId
	NodeId         types_consensus.NodeId
	ValAddrToIdMap types_consensus.ValAddrToIdMap // TODO(design): This needs to be updated every time the ValMap is modified
	IdToValAddrMap types_consensus.IdToValAddrMap // TODO(design): This needs to be updated every time the ValMap is modified

	// Module Dependencies
	utilityContext    modules.UtilityContext
	paceMaker         Pacemaker
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
	paceMaker, err := CreatePacemaker(cfg)
	if err != nil {
		return nil, err
	}

	address := cfg.PrivateKey.Address().String()
	valIdMap, idValMap := types_consensus.GetValAddrToIdMap(types.GetTestState(nil).ValidatorMap)

	m := &consensusModule{
		bus:        nil,
		privateKey: cfg.PrivateKey,

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

func (m *consensusModule) HandleMessage(message *anypb.Any) error {
	var consensusMessage types_consensus.ConsensusMessage
	err := anypb.UnmarshalTo(message, &consensusMessage, proto.UnmarshalOptions{})
	if err != nil {
		return err
	}

	switch consensusMessage.Type {
	case HotstuffMessage:
		var hotstuffMessage types_consensus.HotstuffMessage
		err := anypb.UnmarshalTo(consensusMessage.Message, &hotstuffMessage, proto.UnmarshalOptions{})
		if err != nil {
			return err
		}
		m.handleHotstuffMessage(&hotstuffMessage)
	case UtilityMessage:
		m.nodeLog("[WARN] UtilityMessage handling is not implemented by consensus yet...")
	default:
		return fmt.Errorf("unknown consensus message type: %v", consensusMessage.Type)
	}

	return nil
}
