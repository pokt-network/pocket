package consensus

import (
	"log"

	"github.com/pokt-network/pocket/shared/types"

	"github.com/pokt-network/pocket/consensus/leader_election"
	typesCons "github.com/pokt-network/pocket/consensus/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types/nodestate"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/pokt-network/pocket/shared/config"
	"github.com/pokt-network/pocket/shared/modules"
)

const (
	DefaultLogPrefix string = "NODE" // Just a default that'll be replaced during consensus operations.
)

var _ modules.ConsensusModule = &consensusModule{}

// TODO(olshansky): Any reason to make all of these attributes local only (i.e. not exposed outside the struct)?
type consensusModule struct {
	bus        modules.Bus
	privateKey cryptoPocket.Ed25519PrivateKey
	consCfg    *config.ConsensusConfig

	// Hotstuff
	Height uint64
	Round  uint64
	Step   typesCons.HotstuffStep
	Block  *types.Block // The current block being voted on prior to committing to finality

	HighPrepareQC *typesCons.QuorumCertificate // Highest QC for which replica voted PRECOMMIT
	LockedQC      *typesCons.QuorumCertificate // Highest QC for which replica voted COMMIT

	// Leader Election
	LeaderId       *typesCons.NodeId
	NodeId         typesCons.NodeId
	ValAddrToIdMap typesCons.ValAddrToIdMap // TODO(design): This needs to be updated every time the ValMap is modified
	IdToValAddrMap typesCons.IdToValAddrMap // TODO(design): This needs to be updated every time the ValMap is modified

	// Module Dependencies
	utilityContext    modules.UtilityContext
	paceMaker         Pacemaker
	leaderElectionMod leader_election.LeaderElectionModule

	logPrefix   string                                                  // TODO(design): Remove later when we build a shared/proper/injected logger
	MessagePool map[typesCons.HotstuffStep][]*typesCons.HotstuffMessage // TODO(design): Move this over to the persistence module or elsewhere?
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
	valIdMap, idValMap := typesCons.GetValAddrToIdMap(nodestate.GetNodeState(nil).ValidatorMap)

	m := &consensusModule{
		bus:        nil,
		privateKey: cfg.PrivateKey,
		consCfg:    cfg.Consensus,

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
		MessagePool: make(map[typesCons.HotstuffStep][]*typesCons.HotstuffMessage),
	}

	// TODO(olshansky): Look for a way to avoid doing this.
	paceMaker.SetConsensusModule(m)

	return m, nil
}

func (m *consensusModule) Start() error {
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

// TODO(discuss): Low priority design: think of a way to make `hotstuff_*` files be a sub-package under consensus.
// This is currently not possible because functions tied to the `consensusModule`
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

// TODO(olshansky): Should we just make these singletons or embed them directly in the consensusModule?
type HotstuffMessageHandler interface {
	HandleNewRoundMessage(*consensusModule, *typesCons.HotstuffMessage)
	HandlePrepareMessage(*consensusModule, *typesCons.HotstuffMessage)
	HandlePrecommitMessage(*consensusModule, *typesCons.HotstuffMessage)
	HandleCommitMessage(*consensusModule, *typesCons.HotstuffMessage)
	HandleDecideMessage(*consensusModule, *typesCons.HotstuffMessage)
}

func (m *consensusModule) HandleMessage(message *anypb.Any) error {
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

func (m *consensusModule) handleHotstuffMessage(msg *typesCons.HotstuffMessage) {
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
