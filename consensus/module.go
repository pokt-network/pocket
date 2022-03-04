package consensus

import (
	"fmt"
	"log"

	"github.com/pokt-network/pocket/consensus/leader_election"
	types_consensus "github.com/pokt-network/pocket/consensus/types"

	"github.com/pokt-network/pocket/shared/config"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/types"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	DefaultLogPrefix string = "NODE" // Just a default that'll be replaced during consensus operations.
)

var _ modules.ConsensusModule = &consensusModule{}

// TODO(olshansky): Any reason to make all of these attributes local only (i.e. not exposed outside the struct)?
type consensusModule struct {
	bus modules.Bus

	// Hotstuff
	Height uint64
	Round  uint64
	Step   types_consensus.HotstuffStep
	// TODO(olshansky): Merge with block from utility
	Block *types_consensus.BlockConsensusTemp // The current block being voted on priot to committing to finality

	HighPrepareQC *types_consensus.QuorumCertificate // Highest QC for which replica voted PRECOMMIT
	LockedQC      *types_consensus.QuorumCertificate // Highest QC for which replica voted COMMIT

	// Leader Election
	NodeId    types_consensus.NodeId
	LeaderId  *types_consensus.NodeId
	NodeIdMap types_consensus.ValToIdMap // TODO(design): This needs to be updated every time the ValMap is modified

	// Module Dependencies
	utilityContext    modules.UtilityContext
	paceMaker         PaceMaker
	leaderElectionMod leader_election.LeaderElectionModule

	// TODO(design): Remove later when we build a shared/proper/injected logger
	logPrefix string
	// TODO(design): Move this over to the persistence module or elsewhere?
	MessagePool map[types_consensus.HotstuffStep][]types_consensus.HotstuffMessage
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

	state := types.GetTestState(nil)
	nodeIdMap := types_consensus.GetValToIdMap(types.GetTestState(nil).ValidatorMap)

	m := &consensusModule{
		bus: nil,

		Height: 0,
		Round:  0,
		Step:   NewRound,
		Block:  nil,

		HighPrepareQC: nil,
		LockedQC:      nil,

		NodeId:    nodeIdMap[state.PrivateKey.Address().String()],
		LeaderId:  nil,
		NodeIdMap: nodeIdMap,

		utilityContext:    nil,
		paceMaker:         paceMaker,
		leaderElectionMod: leaderElectionMod,

		logPrefix:   DefaultLogPrefix,
		MessagePool: make(map[types_consensus.HotstuffStep][]types_consensus.HotstuffMessage),
	}

	// TODO(olshansky): Can I avoid doing this?
	paceMaker.SetConsensusMod(m)

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
	case types_consensus.ConsensusMessageType_CONSENSUS_HOTSTUFF_MESSAGE:
		var hotstuffMessage types_consensus.HotstuffMessage
		err := anypb.UnmarshalTo(consensusMessage.Message, &hotstuffMessage, proto.UnmarshalOptions{})
		if err != nil {
			return err
		}
		m.handleHotstuffMessage(&hotstuffMessage)
	default:
		return fmt.Errorf("Unknown consensus message type: %v", consensusMessage.Type)
	}
	return nil
}

func (m *consensusModule) HandleDebugMessage(debugMessage *types.DebugMessage) error {
	switch debugMessage.Action {
	case types.DebugMessageAction_DEBUG_CONSENSUS_RESET_TO_GENESIS:
		log.Println("TEMP")
	case types.DebugMessageAction_DEBUG_CONSENSUS_PRINT_NODE_STATE:
		log.Println("TEMP")
	case types.DebugMessageAction_DEBUG_CONSENSUS_TRIGGER_NEXT_VIEW:
		log.Println("TEMP")
	case types.DebugMessageAction_DEBUG_CONSENSUS_TOGGLE_PACE_MAKER_MODE:
		log.Println("TEMP")
	default:
		log.Printf("Debug message: %s \n", debugMessage.Message)
	}
	return nil
}

// func (m *consensusModule) handleTransaction(anyMessage *anypb.Any) {
// 	messageProto := &types_consensus.Message{}

// 	if err := anyMessage.UnmarshalTo(messageProto); err != nil {
// 		m.nodeLogError("[HandleTransaction] Error unmarshalling message: %v" + err.Error())
// 		return
// 	}

// 	// TODO: decode data, basic validation, send to utility module.
// 	module := m.GetBus().GetUtilityModule()
// 	m.utilityContext, _ = module.NewContext(int64(m.Height))
// 	if err := m.utilityContext.CheckTransaction(messageProto.Data); err != nil {
// 		m.nodeLogError(err.Error())
// 	}
// 	fmt.Println("TRANSACTION IS CHECKED")
// 	m.utilityContext.ReleaseContext()
// }
