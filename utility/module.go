package utility

import (
	"log"

	"github.com/pokt-network/pocket/utility/types"

	"github.com/pokt-network/pocket/shared/modules"
)

var _ modules.UtilityModule = &UtilityModule{}
var _ modules.UtilityConfig = &types.UtilityConfig{}
var _ modules.Module = &UtilityModule{}

type UtilityModule struct {
	bus     modules.Bus
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
	return &UtilityModule{
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
