package leader_election

import (
	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
)

type LeaderElectionModule interface {
	modules.Module
	ElectNextLeader(*typesCons.HotstuffMessage) (typesCons.NodeId, error)
}

var _ LeaderElectionModule = &leaderElectionModule{}

type leaderElectionModule struct {
	base_modules.IntegratableModule
	base_modules.InterruptableModule
}

func Create(bus modules.Bus) (modules.Module, error) {
	return new(leaderElectionModule).Create(bus)
}

func (*leaderElectionModule) Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	m := &leaderElectionModule{}

	for _, option := range options {
		option(m)
	}

	bus.RegisterModule(m)
	return m, nil
}

func (m *leaderElectionModule) GetModuleName() string {
	return modules.LeaderElectionModuleName
}

func (m *leaderElectionModule) ElectNextLeader(msg *typesCons.HotstuffMessage) (typesCons.NodeId, error) {
	nodeId, err := m.electNextLeaderDeterministicRoundRobin(msg)
	if err != nil {
		return typesCons.NodeId(0), err
	}
	return nodeId, nil
}

func (m *leaderElectionModule) electNextLeaderDeterministicRoundRobin(msg *typesCons.HotstuffMessage) (typesCons.NodeId, error) {
	height := int64(msg.Height)
	readCtx, err := m.GetBus().GetPersistenceModule().NewReadContext(height)
	if err != nil {
		return typesCons.NodeId(0), err
	}
	defer readCtx.Release()

	vals, err := readCtx.GetAllValidators(height)
	if err != nil {
		return typesCons.NodeId(0), err
	}

	value := int64(msg.Height) + int64(msg.Round) - 1
	numVals := int64(len(vals))

	return typesCons.NodeId(value%numVals + 1), nil
}
