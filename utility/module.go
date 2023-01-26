package utility

import (
	"log"

	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/utility/types"
)

var (
	_ modules.UtilityModule = &utilityModule{}
	_ modules.Module        = &utilityModule{}
)

type utilityModule struct {
	bus    modules.Bus
	config *configs.UtilityConfig

	mempool types.Mempool
}

func Create(bus modules.Bus) (modules.Module, error) {
	return new(utilityModule).Create(bus)
}

func (*utilityModule) Create(bus modules.Bus) (modules.Module, error) {
	m := &utilityModule{}
	bus.RegisterModule(m)

	runtimeMgr := bus.GetRuntimeMgr()

	cfg := runtimeMgr.GetConfig()
	utilityCfg := cfg.Utility

	m.config = utilityCfg
	m.mempool = types.NewMempool(utilityCfg.MaxMempoolTransactionBytes, utilityCfg.MaxMempoolTransactions)

	return m, nil
}

func (u *utilityModule) Start() error {
	return nil
}

func (u *utilityModule) Stop() error {
	return nil
}

func (u *utilityModule) GetModuleName() string {
	return modules.UtilityModuleName
}

func (u *utilityModule) SetBus(bus modules.Bus) {
	u.bus = bus
}

func (u *utilityModule) GetBus() modules.Bus {
	if u.bus == nil {
		log.Fatalf("Bus is not initialized")
	}
	return u.bus
}
