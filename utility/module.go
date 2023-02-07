package utility

import (
	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/shared/mempool"
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

	logger  modules.Logger
	mempool mempool.TXMempool
}

func Create(bus modules.Bus) (modules.Module, error) {
	return new(utilityModule).Create(bus)
}

func (*utilityModule) Create(bus modules.Bus) (modules.Module, error) {
	m := &utilityModule{}
	if err := bus.RegisterModule(m); err != nil {
		return nil, err
	}

	runtimeMgr := bus.GetRuntimeMgr()

	cfg := runtimeMgr.GetConfig()
	utilityCfg := cfg.Utility

	m.config = utilityCfg
	m.mempool = types.NewTxFIFOMempool(utilityCfg.MaxMempoolTransactionBytes, utilityCfg.MaxMempoolTransactions)

	return m, nil
}

func (u *utilityModule) Start() error {
	u.logger = logger.Global.CreateLoggerForModule(u.GetModuleName())
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
		u.logger.Fatal().Msg("Bus is not initialized")
	}
	return u.bus
}
