package consensus

import (
	"encoding/gob"
	"log"

	"pocket/consensus/pkg/consensus/dkg"
	"pocket/consensus/pkg/consensus/leader_election"
	"pocket/consensus/pkg/consensus/statesync"
	consensus_types "pocket/consensus/pkg/consensus/types"
	"pocket/consensus/pkg/shared/context"
	"pocket/consensus/pkg/shared/modules"
	"pocket/consensus/pkg/types"
	"pocket/consensus/pkg/types/typespb"
)

const (
	DefaultLogPrefix string = "NODE" // Just a default that'll be replaced during consensus operations.
)

type consensusModule struct {
	*modules.BasePocketModule
	modules.ConsensusModule

	// Hotstuff
	Height BlockHeight
	Round  Round
	Step   Step
	Block  *typespb.Block // The current block being proposed.

	HighPrepareQC *QuorumCertificate // highest QC for which replica voted PRECOMMIT.
	LockedQC      *QuorumCertificate // highest QC for which replica voted COMMIT.

	// Leader Election
	NodeId   types.NodeId
	LeaderId *types.NodeId

	// Crypto
	PrivateKey types.PrivateKey // should we generalize to crypto.PrivateKey?

	// Module Dependencies
	paceMaker         PaceMaker
	stateSyncMod      statesync.StateSyncModule
	dkgMod            dkg.DKGModule
	leaderElectionMod leader_election.LeaderElectionModule

	// TODO: Remove later to config/context/injected-logger
	logPrefix string // TODO: Move over to config or context

	// TODO: Move this over to the persistance module or elsewhere?
	// Open questions for mempool:
	// - Should this be a map keyed by (height, round, step)?
	// - Should this be handeled by the persistance m?
	// - How is the mempool management handled between all 4 ms?
	// - When do we clear/remove messages from the mempool?
	MessagePool map[Step][]HotstuffMessage
	// MessagePool map[Height][Step][Roudn][]CandidateLeaderMessage
	// MessagePool map[Height][Step][Roudn][]HotstuffMessage
}

func Create(
	ctx *context.PocketContext,
	base *modules.BasePocketModule,
) (modules.ConsensusModule, error) {
	cfg := base.GetConfig()

	stateSyncMod, err := statesync.Create(ctx, base)
	if err != nil {
		return nil, err
	}

	leaderElectionMod, err := leader_election.Create(ctx, base)
	if err != nil {
		return nil, err
	}

	// TODO: Not used until we moved to threshold signatures.
	dkgMod, err := dkg.Create(ctx, base)
	if err != nil {
		return nil, err
	}

	m := &consensusModule{
		BasePocketModule: base,

		Height: 0,
		Round:  0,
		Step:   NewRound,

		HighPrepareQC: nil,
		LockedQC:      nil,

		NodeId:   cfg.Consensus.NodeId,
		LeaderId: nil,

		PrivateKey: cfg.PrivateKey,

		logPrefix: DefaultLogPrefix,

		paceMaker:         nil, // Updated below because of the 2 way pointer design.
		stateSyncMod:      stateSyncMod,
		dkgMod:            dkgMod,
		leaderElectionMod: leaderElectionMod,

		MessagePool: make(map[Step][]HotstuffMessage),
	}

	paceMaker, err := CreatePaceMaker(ctx, base, m)
	if err != nil {
		return nil, err
	}
	m.paceMaker = paceMaker

	return m, nil
}

func (m *consensusModule) Start(ctx *context.PocketContext) error {
	log.Println("Starting consensus module")

	gob.Register(&DebugMessage{})
	gob.Register(&HotstuffMessage{})
	gob.Register(&statesync.StateSyncMessage{})
	gob.Register(&dkg.DKGMessage{})
	gob.Register(&leader_election.LeaderElectionMessage{})

	if err := m.stateSyncMod.Start(ctx); err != nil {
		return err
	}

	if err := m.paceMaker.Start(ctx); err != nil {
		return err
	}

	if err := m.dkgMod.Start(ctx); err != nil {
		return err
	}

	if err := m.leaderElectionMod.Start(ctx); err != nil {
		return err
	}

	return nil
}

func (m *consensusModule) Stop(ctx *context.PocketContext) error {
	log.Println("Stopping consensus m")
	return nil
}

func (m *consensusModule) GetPocketBusMod() modules.PocketBusModule {
	return m.BasePocketModule.GetPocketBusMod()
}

func (m *consensusModule) SetPocketBusMod(bus modules.PocketBusModule) {
	m.BasePocketModule.SetPocketBusMod(bus)
}

func (m *consensusModule) HandleMessage(ctx *context.PocketContext, message *consensus_types.ConsensusMessage) {
	switch message.Message.GetType() {
	case consensus_types.HotstuffConsensusMessage:
		m.handleHotstuffMessage(message.Message.(*HotstuffMessage))
	case consensus_types.DebugConsensusMessage:
		m.handleDebugMessage(message.Message.(*DebugMessage))
	case consensus_types.DKGConsensusMessage:
		m.dkgMod.HandleMessage(ctx, message.Message.(*dkg.DKGMessage))
	case consensus_types.LeaderElectionMessage:
		m.leaderElectionMod.HandleMessage(ctx, message.Message.(*leader_election.LeaderElectionMessage))
	case consensus_types.StateSyncConsensusMessage:
		log.Println("[TODO] Not implementing StateSyncConsensusMessage")
	}
}

func (m *consensusModule) HandleTransaction(ctx *context.PocketContext, data []byte) {
	// TODO: decode data, basic validation, send to utility module.
	m.GetPocketBusMod().GetUtilityModule().HandleTransaction(ctx, &typespb.Transaction{})
}

func (m *consensusModule) HandleEvidence(ctx *context.PocketContext, data []byte) {
	// TODO: decode data, basic validation, send to utility module.
	m.GetPocketBusMod().GetUtilityModule().HandleEvidence(ctx, &typespb.Evidence{})
}
