package utility

import (
	"log"

	"github.com/pokt-network/pocket/shared/config"
	"github.com/pokt-network/pocket/shared/logging"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/types"
)

var _ modules.UtilityModule = &UtilityModule{}

type UtilityModule struct {
	bus     modules.Bus
	Mempool types.Mempool
}

func Create(_ *config.Config) (modules.UtilityModule, error) {
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

func (u *UtilityModule) SetBus(bus modules.Bus) {
	u.bus = bus
}

func (u *UtilityModule) GetBus() modules.Bus {
	if u.bus == nil {
		log.Fatalf("Bus is not initialized")
	}
	return u.bus
}

func (m *UtilityModule) Logger() logging.Logger {
	return m.GetBus().GetTelemetryModule().Logger()
}
