package leader_election

import (
	"log"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/modules"
)

const (
	leaderElectionModuleName = "leader_election"
)

type LeaderElectionModule interface {
	modules.Module
	ElectNextLeader(*typesCons.HotstuffMessage) (typesCons.NodeId, error)
}

var _ LeaderElectionModule = &leaderElectionModule{}

type leaderElectionModule struct {
	bus modules.Bus
}

func Create(runtime modules.RuntimeMgr) (modules.Module, error) {
	return new(leaderElectionModule).Create(runtime)
}

func (*leaderElectionModule) Create(runtime modules.RuntimeMgr) (modules.Module, error) {
	return &leaderElectionModule{}, nil
}

func (m *leaderElectionModule) Start() error {
	// TODO(olshansky): Use persistence to create leader election module.
	return nil
}

func (m *leaderElectionModule) Stop() error {
	return nil
}

func (m *leaderElectionModule) GetModuleName() string {
	return leaderElectionModuleName
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

	ctx, err := m.GetBus().GetPersistenceModule().NewReadContext(int64(message.Height))
	if err != nil {
		log.Fatal(err)
	}
	vals, err := ctx.GetAllValidators(int64(message.Height))
	if err != nil {
		log.Fatal(err)
	}
	return typesCons.NodeId(value%int64(len(vals)) + 1)
}
