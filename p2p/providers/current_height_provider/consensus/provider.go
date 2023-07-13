package consensus

import (
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
)

var _ modules.CurrentHeightProvider = &consensusCurrentHeightProvider{}

type consensusCurrentHeightProvider struct {
	base_modules.IntegrableModule
}

func Create(bus modules.Bus) (modules.CurrentHeightProvider, error) {
	return new(consensusCurrentHeightProvider).Create(bus)
}

func (*consensusCurrentHeightProvider) Create(bus modules.Bus) (modules.CurrentHeightProvider, error) {
	consCHP := &consensusCurrentHeightProvider{}
	bus.RegisterModule(consCHP)
	return consCHP, nil
}

func (consCHP *consensusCurrentHeightProvider) GetModuleName() string {
	return modules.CurrentHeightProviderSubmoduleName
}

func (consCHP *consensusCurrentHeightProvider) CurrentHeight() uint64 {
	return consCHP.GetBus().GetConsensusModule().CurrentHeight()
}
