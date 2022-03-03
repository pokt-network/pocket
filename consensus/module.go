package consensus

import (
	"encoding/gob"
	"fmt"
	"log"

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

type consensusModule struct {
	bus modules.Bus

	// Hotstuff
	Height BlockHeight
	Round  Round
	Step   Step
	Block  *types_consensus.BlockConsensusTemp // The current block being voted on priot to committing to finality

	HighPrepareQC *QuorumCertificate // Highest QC for which replica voted PRECOMMIT
	LockedQC      *QuorumCertificate // Highest QC for which replica voted COMMIT

	// Leader Election
	NodeId   types_consensus.NodeId
	LeaderId *types_consensus.NodeId // Pointer because it is nullable

	// Module Dependencies
	UtilityContext modules.UtilityContext
	paceMaker      PaceMaker
	// leaderElectionMod leader_election.LeaderElectionModule

	// TODO(design): Remove later when we build a shared/proper/injected logger
	logPrefix string
	// TODO(design): Move this over to the persistence module or elsewhere?
	MessagePool map[Step][]HotstuffMessage
}

func Create(cfg *config.Config) (modules.ConsensusModule, error) {
	gob.Register(&DebugMessage{})
	gob.Register(&HotstuffMessage{})
	// gob.Register(&leader_election.LeaderElectionMessage{})
	// gob.Register(&TxWrapperMessage{})
	state := types.GetTestState()
	state.LoadStateFromConfig(cfg)

	// leaderElectionMod, err := leader_election.Create(cfg)
	// if err != nil {
	// 	return nil, err
	// }
	//
	m := &consensusModule{
		bus: nil,

		Height: 0,
		Round:  0,
		Step:   NewRound,
		Block:  nil,

		HighPrepareQC: nil,
		LockedQC:      nil,

		NodeId:   types_consensus.NodeId(cfg.Consensus.NodeId),
		LeaderId: nil,

		paceMaker: nil, // Updated below because of the 2 way pointer design.
		// leaderElectionMod: leaderElectionMod,

		logPrefix:   DefaultLogPrefix,
		MessagePool: make(map[Step][]HotstuffMessage),
	}

	paceMaker, err := CreatePaceMaker(cfg)
	if err != nil {
		return nil, err
	}
	m.paceMaker = paceMaker
	paceMaker.SetConsensusMod(m)

	return m, nil
}

func (m *consensusModule) Start() error {
	log.Println("Starting consensus module")

	if err := m.paceMaker.Start(); err != nil {
		return err
	}

	// if err := m.leaderElectionMod.Start(); err != nil {
	// 	return err
	// }

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
	// m.leaderElectionMod.SetBus(pocketBus)
}

func (m *consensusModule) Stop() error {
	log.Println("Stopping consensus m")
	return nil
}

func (m *consensusModule) HandleMessage(anyMessage *anypb.Any) {
	d, err := anypb.UnmarshalNew(anyMessage, proto.UnmarshalOptions{})
	if err != nil {
		panic(err.Error()) // TODO remove
	}
	messageProto, ok := d.(*types_consensus.Message)
	if !ok {
		panic("any isn't Message type")
	}
	message, err := types_consensus.DecodeConsensusMessage(messageProto.Data)
	if err != nil {
		m.nodeLogError("[HandleMessage] Error unmarshalling message: %v" + err.Error())
		return
	}

	switch message.Message.GetType() {
	case types_consensus.HotstuffConsensusMessage:
		m.handleHotstuffMessage(message.Message.(*HotstuffMessage))
	// case types_consensus.DebugConsensusMessage:
	// 	m.handleDebugMessage(message.Message.(*DebugMessage))
	// case types_consensus.LeaderElectionMessage:
	// 	m.leaderElectionMod.HandleMessage(message.Message.(*leader_election.LeaderElectionMessage))
	case types_consensus.StateSyncConsensusMessage:
		log.Println("[TODO] Not implementing StateSyncConsensusMessage")
	}
}

func (m *consensusModule) HandleTransaction(anyMessage *anypb.Any) {
	messageProto := &types_consensus.Message{}

	if err := anyMessage.UnmarshalTo(messageProto); err != nil {
		m.nodeLogError("[HandleTransaction] Error unmarshalling message: %v" + err.Error())
		return
	}

	// TODO: decode data, basic validation, send to utility module.
	module := m.GetBus().GetUtilityModule()
	m.UtilityContext, _ = module.NewContext(int64(m.Height))
	if err := m.UtilityContext.CheckTransaction(messageProto.Data); err != nil {
		m.nodeLogError(err.Error())
	}
	fmt.Println("TRANSACTION IS CHECKED")
	m.UtilityContext.ReleaseContext()
}
