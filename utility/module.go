package utility

import (
	"log"

	"github.com/pokt-network/pocket/utility/types"

	"github.com/pokt-network/pocket/shared/modules"
)

var _ modules.UtilityModule = &UtilityModule{}
var _ modules.UtilityConfig = &types.UtilityConfig{}

type UtilityModule struct {
	bus     modules.Bus
	Mempool types.Mempool
}

const (
	UtilityModuleName = "utility"
)

func Create(configPath, genesisPath string) (modules.UtilityModule, error) {
	return &UtilityModule{
		// TODO: Add `maxTransactionBytes` and `maxTransactions` to cfg.Utility
		Mempool: types.NewMempool(1000, 1000),
	}, nil
}

func (u *UtilityModule) InitConfig(pathToConfigJSON string) (config modules.ConfigI, err error) {
	return // No-op
}

func (u *UtilityModule) InitGenesis(pathToGenesisJSON string) (genesis modules.GenesisI, err error) {
	return // No-op
}

func (u *UtilityModule) Start() error {
	return nil
}

func (u *UtilityModule) Stop() error {
	return nil
}

func (u *UtilityModule) GetModuleName() string {
	return UtilityModuleName
}

func (u *UtilityModule) SetBus(bus modules.Bus) {
	u.bus = bus
}

func (u *UtilityModule) GetBus() modules.Bus {
	if u.bus == nil {
		log.Fatalf("Bus is not initialized")
	}
	return u.bus
}
