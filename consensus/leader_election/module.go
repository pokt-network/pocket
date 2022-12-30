package leader_election

import (
	"log"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/modules"
)

type LeaderElectionModule interface {
	modules.Module
	ElectNextLeader(*typesCons.HotstuffMessage) (typesCons.NodeId, error)
}

var _ LeaderElectionModule = &leaderElectionModule{}

type leaderElectionModule struct {
	bus modules.Bus
}

func Create(bus modules.Bus) (modules.Module, error) {
	return new(leaderElectionModule).Create(bus)
}

func (*leaderElectionModule) Create(bus modules.Bus) (modules.Module, error) {
	m := &leaderElectionModule{}
	bus.RegisterModule(m)
	return m, nil
}

func (m *leaderElectionModule) Start() error {
	// TODO(olshansky): Use persistence to create leader election module.
	return nil
}

func (m *leaderElectionModule) Stop() error {
	return nil
}

func (m *leaderElectionModule) GetModuleName() string {
	return modules.LeaderElectionModuleName
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
	nodeId, err := m.electNextLeaderDeterministicRoundRobin(message)
	if err != nil {
		return typesCons.NodeId(0), err
	}
	return nodeId, nil
}

func (m *leaderElectionModule) electNextLeaderDeterministicRoundRobin(message *typesCons.HotstuffMessage) (typesCons.NodeId, error) {
	value := int64(message.Height) + int64(message.Round) + int64(message.Step) - 1

	uCtx, err := m.GetBus().GetUtilityModule().NewContext(int64(message.Height))
	if err != nil {
		return typesCons.NodeId(0), err
	}
	vals, err := uCtx.GetPersistenceContext().GetAllValidators(int64(message.Height))
	if err != nil {
		return typesCons.NodeId(0), err
	}
	return typesCons.NodeId(value%int64(len(vals)) + 1), nil
}
