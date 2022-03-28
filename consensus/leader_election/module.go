package leader_election

import (
	"log"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/config"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/types"
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
	config *config.Config,
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
	valMap := types.GetTestState(nil).ValidatorMap
	value := int64(message.Height) + int64(message.Round) + int64(message.Step) - 1
	return typesCons.NodeId(value%int64(len(valMap)) + 1)
}
