package utility

import (
	"fmt"
	"log"

	"github.com/pokt-network/pocket/utility/types"

	"github.com/pokt-network/pocket/shared/modules"
)

var _ modules.UtilityModule = &UtilityModule{}
var _ modules.UtilityConfig = &types.UtilityConfig{}
var _ modules.Module = &UtilityModule{}

type UtilityModule struct {
	bus    modules.Bus
	config *types.UtilityConfig

	Mempool types.Mempool
}

const (
	UtilityModuleName = "utility"
)

func Create(runtime modules.Runtime) (modules.Module, error) {
	var m UtilityModule
	return m.Create(runtime)
}

func (*UtilityModule) Create(runtime modules.Runtime) (modules.Module, error) {
	var m *UtilityModule

	cfg := runtime.GetConfig()
	if err := m.ValidateConfig(cfg); err != nil {
		log.Fatalf("config validation failed: %v", err)
	}
	utilityCfg := cfg.Utility.(*types.UtilityConfig)

	return &UtilityModule{
		config: utilityCfg,
		// TODO: Add `maxTransactionBytes` and `maxTransactions` to cfg.Utility
		Mempool: types.NewMempool(1000, 1000),
	}, nil
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

func (*UtilityModule) ValidateConfig(cfg modules.Config) error {
	if _, ok := cfg.Utility.(*types.UtilityConfig); !ok {
		return fmt.Errorf("cannot cast to UtilityConfig")
	}
	return nil
}
