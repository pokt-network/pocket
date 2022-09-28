package utility

import (
	"fmt"
	"log"

	"github.com/pokt-network/pocket/utility/types"

	"github.com/pokt-network/pocket/shared/modules"
)

var _ modules.UtilityModule = &utilityModule{}
var _ modules.UtilityConfig = &types.UtilityConfig{}
var _ modules.Module = &utilityModule{}

type utilityModule struct {
	bus    modules.Bus
	config *types.UtilityConfig

	Mempool types.Mempool
}

const (
	UtilityModuleName = "utility"
)

func Create(runtime modules.Runtime) (modules.Module, error) {
	return new(utilityModule).Create(runtime)
}

func (*utilityModule) Create(runtime modules.Runtime) (modules.Module, error) {
	var m *utilityModule

	cfg := runtime.GetConfig()
	if err := m.ValidateConfig(cfg); err != nil {
		log.Fatalf("config validation failed: %v", err)
	}
	utilityCfg := cfg.Utility.(*types.UtilityConfig)

	return &utilityModule{
		config: utilityCfg,
		// TODO: Add `maxTransactionBytes` and `maxTransactions` to cfg.Utility
		Mempool: types.NewMempool(1000, 1000),
	}, nil
}

func (u *utilityModule) Start() error {
	return nil
}

func (u *utilityModule) Stop() error {
	return nil
}

func (u *utilityModule) GetModuleName() string {
	return UtilityModuleName
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

func (*utilityModule) ValidateConfig(cfg modules.Config) error {
	if _, ok := cfg.Utility.(*types.UtilityConfig); !ok {
		return fmt.Errorf("cannot cast to UtilityConfig")
	}
	return nil
}
