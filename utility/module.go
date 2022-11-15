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
	config modules.UtilityConfig

	Mempool types.Mempool
}

const (
	utilityModuleName = "utility"
)

func Create(runtime modules.RuntimeMgr) (modules.Module, error) {
	return new(utilityModule).Create(runtime)
}

func (*utilityModule) Create(runtime modules.RuntimeMgr) (modules.Module, error) {
	var m *utilityModule

	cfg := runtime.GetConfig()
	if err := m.ValidateConfig(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}
	utilityCfg := cfg.GetUtilityConfig()

	return &utilityModule{
		config:  utilityCfg,
		Mempool: types.NewMempool(utilityCfg.GetMaxMempoolTransactionBytes(), utilityCfg.GetMaxMempoolTransactions()),
	}, nil
}

func (u *utilityModule) Start() error {
	return nil
}

func (u *utilityModule) Stop() error {
	return nil
}

func (u *utilityModule) GetModuleName() string {
	return utilityModuleName
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
	// TODO (#334): implement this
	return nil
}
