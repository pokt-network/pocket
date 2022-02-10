package leader_election

import (
	"log"

	"pocket/consensus/pkg/consensus/leader_election/vrf"
	consensus_types "pocket/consensus/pkg/consensus/types"
	"pocket/consensus/pkg/p2p"
	"pocket/consensus/pkg/p2p/p2p_types"
	"pocket/consensus/pkg/shared"
	"pocket/consensus/pkg/shared/context"
	"pocket/consensus/pkg/shared/events"
	"pocket/consensus/pkg/shared/modules"
	"pocket/consensus/pkg/types"
)

// HotPocket will only have one leader per round but we set this value to 3
// to increase the likelihood of someone getting elected and avoid needing to
// default to the round robin method.
const NumCandidatesLeadersPerRound float64 = 3

type LeaderElectionModule interface {
	modules.PocketModule

	HandleMessage(*context.PocketContext, *LeaderElectionMessage)
	RegenerateVRFKeys(*context.PocketContext, consensus_types.BlockHeight, consensus_types.Round)            // Needs to be triggered every N blocks depending on some parameter.
	BroadcastVRFProofIfCandidate(*context.PocketContext, consensus_types.BlockHeight, consensus_types.Round) // Needs to be executed on every DECIDE phase.

	GetLeaderId(consensus_types.BlockHeight, consensus_types.Round) (types.NodeId, error)
}

type leaderElectionModule struct {
	*modules.BasePocketModule

	// Module metadata
	nodeId         types.NodeId
	previousLeader *types.NodeId

	// VRF
	vrfSecretKey       *vrf.SecretKey
	vrfVerificationKey *vrf.VerificationKey
}

type LeaderElectionMethod uint8

const (
	RoundRobin LeaderElectionMethod = iota
	VRFWithCDF
)

func Create(
	ctx *context.PocketContext,
	base *modules.BasePocketModule,
) (LeaderElectionModule, error) {
	state := shared.GetPocketState()

	return &leaderElectionModule{
		BasePocketModule: base,

		nodeId:         state.Config.Consensus.NodeId,
		previousLeader: nil,

		vrfSecretKey:       nil,
		vrfVerificationKey: nil,
	}, nil
}

func (m *leaderElectionModule) Start(*context.PocketContext) error {
	log.Println("[TODO] Use persistance to create leader election module.")

	return nil
}

func (m *leaderElectionModule) Stop(*context.PocketContext) error {
	log.Println("Stopping leader election module")
	return nil
}

func (m *leaderElectionModule) HandleMessage(ctx *context.PocketContext, message *LeaderElectionMessage) {
	switch message.Type {
	case VRFKeyBroadcast:
		log.Println("[TODO] Handle VRF key broadcast")
	case VRFProofBroadcast:
		log.Println("[TODO] Handle VRF proof broadcast")
	default:
		log.Println("[ERROR] Unknown message type:", message.Type)
	}
}

func (m *leaderElectionModule) RegenerateVRFKeys(ctx *context.PocketContext, height consensus_types.BlockHeight, round consensus_types.Round) {
	sk, vk, err := vrf.GenerateVRFKeys(nil)
	if err != nil {
		log.Println("[ERROR] Failed to generate VRF keys:", err)
		return
	}
	m.vrfSecretKey = sk
	m.vrfVerificationKey = vk

	message := &LeaderElectionMessage{
		Height: height,
		Round:  round,

		Type: VRFKeyBroadcast,
		KeyMsg: &LeaderElectionKeyBroadcastMessage{
			VerificationKey: *vk,
			VKStartHeight:   height,
			VKEndHeight:     height + 1,
		},

		Sender: m.nodeId,
	}
	err = m.publishLeaderElectionMessage(message)
	if err != nil {
		log.Println("[ERROR] Failed to broadcast VRF proof:", err)
	}
}

func (m *leaderElectionModule) GetLeaderId(height consensus_types.BlockHeight, round consensus_types.Round) (types.NodeId, error) {
	// Run SelectLead
	return 0, nil
}

func (m *leaderElectionModule) BroadcastVRFProofIfCandidate(ctx *context.PocketContext, height consensus_types.BlockHeight, round consensus_types.Round) {
	state := shared.GetPocketState()

	validator, ok := state.ValidatorMap[m.nodeId]
	if !ok {
		log.Printf("[ERROR] Cannot broadcast VRF Proof because validator not foudn in Pocket State: %d", m.nodeId)
		return
	}

	prevBlockHash, err := m.GetPocketBusMod().GetPersistanceModule().GetBlockHash(uint64(height) - 1)
	if err != nil {
		log.Printf("[ERROR] Cannot determine the block hash for height: %d", height-1)
		return
	}

	if m.vrfSecretKey == nil {
		log.Printf("[ERROR] Cannot broadcast VRF proof for leader candidate if the secret key is nil")
		return
	}

	leaderCandidate, err := IsLeaderCandidate(
		validator,
		height,
		round,
		string(prevBlockHash),
		float64(validator.UPokt),        // TODO: Guarantee this value is up to date.
		float64(state.TotalVotingPower), // TODO: Guarantee this value is up to date.
		NumCandidatesLeadersPerRound,
		m.vrfSecretKey,
	)
	if err != nil {
		log.Printf("[ERROR] Cannot determine if validator %d is a candidate leader: %s", m.nodeId, err)
		return
	}

	if leaderCandidate == nil {
		log.Printf("[INFO] %d is not a candidate leader for height %d round %d\n", m.nodeId, height, round)
		return
	}

	message := &LeaderElectionMessage{
		Height: height,
		Round:  round,

		Type: VRFProofBroadcast,
		ProofMsg: &LeaderElectionProofBroadcastMessage{
			VRFOut:          leaderCandidate.vrfOut,
			VRFProof:        leaderCandidate.vrfProof,
			SortitionResult: leaderCandidate.sortitionResult,
		},

		Sender: m.nodeId,
	}
	err = m.publishLeaderElectionMessage(message)
	if err != nil {
		log.Println("[ERROR] Failed to broadcast VRF proof:", err)
	}
}

func (m *leaderElectionModule) publishLeaderElectionMessage(message *LeaderElectionMessage) error {
	consensusMessage := &consensus_types.ConsensusMessage{
		Message: message,
		Sender:  m.nodeId,
	}
	data, err := consensus_types.EncodeConsensusMessage(consensusMessage)
	if err != nil {
		return err
	}

	networkMsg := &p2p_types.NetworkMessage{
		Topic: events.CONSENSUS_MESSAGE,
		Data:  data,
	}
	networkMsgEncoded, err := p2p.EncodeNetworkMessage(networkMsg)
	if err != nil {
		return err
	}

	e := &events.PocketEvent{
		SourceModule: events.LEADER_ELECTION,
		PocketTopic:  events.P2P_BROADCAST_MESSAGE,
		MessageData:  networkMsgEncoded,
	}
	m.GetPocketBusMod().PublishEventToBus(e)
	return nil
}
