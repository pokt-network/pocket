package leader_election

import (
	"github.com/pokt-network/pocket/shared/types/genesis"
	"log"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/modules"
)

type LeaderElectionModule interface {
	modules.Module
	ElectNextLeader(*typesCons.HotstuffMessage) (typesCons.NodeId, error)
}

var _ leaderElectionModule = leaderElectionModule{}

type leaderElectionModule struct {
	bus modules.Bus
}

func Create(
	_ *genesis.Config,
) (LeaderElectionModule, error) {
	return &leaderElectionModule{}, nil
}

func (m *leaderElectionModule) Start() error {
	// TODO(olshansky): Use persistence to create leader election module.
	return nil
}

func (m *leaderElectionModule) Stop() error {
	return nil
}

func (m *leaderElectionModule) SetBus(pocketBus modules.Bus) {
	m.bus = pocketBus
}

func (m *leaderElectionModule) GetBus() modules.Bus {
	if m.bus == nil {
		log.Fatalf("PocketBus is not initialized")
	}
	return m.bus
}

func (m *leaderElectionModule) ElectNextLeader(message *typesCons.HotstuffMessage) (typesCons.NodeId, error) {
	return m.electNextLeaderDeterministicRoundRobin(message), nil
}

func (m *leaderElectionModule) electNextLeaderDeterministicRoundRobin(message *typesCons.HotstuffMessage) typesCons.NodeId {
	value := int64(message.Height) + int64(message.Round) + int64(message.Step) - 1
	return typesCons.NodeId(value%int64(len(m.GetBus().GetConsensusModule().ValidatorMap())) + 1)
}
