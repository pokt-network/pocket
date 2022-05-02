package consensus

import (
	"encoding/gob"
	"fmt"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"log"
	"pocket/consensus/dkg"
	"pocket/consensus/leader_election"
	"pocket/consensus/statesync"
	consensus_types "pocket/consensus/types"
	"pocket/shared/config"
	"pocket/shared/crypto"
	"pocket/shared/modules"
)

const (
	DefaultLogPrefix string = "NODE" // Just a default that'll be replaced during consensus operations.
)

type ConsensusModule struct {
	modules.ConsensusModule
	pocketBusMod modules.Bus

	// Hotstuff
	Height BlockHeight
	Round  Round
	Step   Step
	Block  *consensus_types.BlockConsTemp // The current block being proposed.

	HighPrepareQC *QuorumCertificate // highest QC for which replica voted PRECOMMIT.
	LockedQC      *QuorumCertificate // highest QC for which replica voted COMMIT.

	// Leader Election
	NodeId   consensus_types.NodeId
	LeaderId *consensus_types.NodeId

	// Crypto
	PrivateKey crypto.PrivateKey // should we generalize to crypto.PrivateKey?

	// Module Dependencies
	paceMaker         PaceMaker
	stateSyncMod      statesync.StateSyncModule
	dkgMod            dkg.DKGModule
	leaderElectionMod leader_election.LeaderElectionModule

	// TODO: Remove later to build/config/context/injected-logger
	logPrefix string // TODO: Move over to config or context

	// TODO: Move this over to the persistence module or elsewhere?
	// Open questions for mempool:
	// - Should this be a map keyed by (height, round, step)?
	// - Should this be handeled by the persistence m?
	// - How is the mempool management handled between all 4 ms?
	// - When do we clear/remove messages from the mempool?
	MessagePool map[Step][]HotstuffMessage
	// MessagePool map[Height][Step][Roudn][]CandidateLeaderMessage
	// MessagePool map[Height][Step][Roudn][]HotstuffMessage

	UtilityContext modules.UtilityContext
}

func Create(cfg *config.Config) (modules.ConsensusModule, error) {
	gob.Register(&DebugMessage{})
	gob.Register(&HotstuffMessage{})
	gob.Register(&statesync.StateSyncMessage{})
	gob.Register(&dkg.DKGMessage{})
	gob.Register(&leader_election.LeaderElectionMessage{})
	gob.Register(&TxWrapperMessage{})
	state := consensus_types.GetTestState()
	state.LoadStateFromConfig(cfg)

	//stateSyncMod, err := statesync.Create(ctx, base)
	//if err != nil {
	//	return nil, err
	//}
	//
	leaderElectionMod, err := leader_election.Create(cfg)
	if err != nil {
		return nil, err
	}
	//
	//// TODO: Not used until we moved to threshold signatures.
	//dkgMod, err := dkg.Create(ctx, base)
	//if err != nil {
	//	return nil, err
	//}
	pk, err := crypto.NewPrivateKey(cfg.PrivateKey)
	if err != nil {
		panic(err)
	}
	m := &ConsensusModule{
		Height: 0,
		Round:  0,
		Step:   NewRound,

		HighPrepareQC: nil,
		LockedQC:      nil,

		NodeId:   consensus_types.NodeId(cfg.Consensus.NodeId),
		LeaderId: nil,

		PrivateKey: pk,

		logPrefix: DefaultLogPrefix,

		paceMaker:         nil, // Updated below because of the 2 way pointer design.
		stateSyncMod:      nil,
		dkgMod:            nil,
		leaderElectionMod: leaderElectionMod,

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

func (m *ConsensusModule) Start() error {
	log.Println("Starting consensus module")

	//if err := m.dkgMod.Start(); err != nil {
	//	return err
	//}

	if err := m.paceMaker.Start(); err != nil {
		return err
	}

	if err := m.leaderElectionMod.Start(); err != nil {
		return err
	}

	//if err := m.stateSyncMod.Start(); err != nil {
	//	return err
	//}

	return nil
}

func (m *ConsensusModule) GetBus() modules.Bus {
	if m.pocketBusMod == nil {
		log.Fatalf("PocketBus is not initialized")
	}
	return m.pocketBusMod
}

func (m *ConsensusModule) SetBus(pocketBus modules.Bus) {
	m.pocketBusMod = pocketBus
	m.paceMaker.SetBus(pocketBus)
	m.leaderElectionMod.SetBus(pocketBus)
}

func (m *ConsensusModule) Stop() error {
	log.Println("Stopping consensus m")
	return nil
}

func (m *ConsensusModule) HandleMessage(anyMessage *anypb.Any) {
	d, err := anypb.UnmarshalNew(anyMessage, proto.UnmarshalOptions{})
	if err != nil {
		panic(err.Error()) // TODO remove
	}
	messageProto, ok := d.(*consensus_types.Message)
	if !ok {
		panic("any isn't Message type")
	}
	message, err := consensus_types.DecodeConsensusMessage(messageProto.Data)
	if err != nil {
		m.nodeLogError("[HandleMessage] Error unmarshalling message: %v" + err.Error())
		return
	}

	switch message.Message.GetType() {
	case consensus_types.HotstuffConsensusMessage:
		m.handleHotstuffMessage(message.Message.(*HotstuffMessage))
	case consensus_types.DebugConsensusMessage:
		m.handleDebugMessage(message.Message.(*DebugMessage))
	case consensus_types.DKGConsensusMessage:
		m.dkgMod.HandleMessage(message.Message.(*dkg.DKGMessage))
	case consensus_types.LeaderElectionMessage:
		m.leaderElectionMod.HandleMessage(message.Message.(*leader_election.LeaderElectionMessage))
	case consensus_types.StateSyncConsensusMessage:
		log.Println("[TODO] Not implementing StateSyncConsensusMessage")
	}
}

func (m *ConsensusModule) HandleTransaction(anyMessage *anypb.Any) {
	messageProto := &consensus_types.Message{}

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
