package utility

import (
	"encoding/json"
	"github.com/pokt-network/pocket/utility/types"
	"log"

	"github.com/pokt-network/pocket/shared/modules"
)

var _ modules.UtilityModule = &UtilityModule{}
var _ modules.UtilityConfig = &types.UtilityConfig{}

type UtilityModule struct {
	bus     modules.Bus
	Mempool types.Mempool
}

func Create(config, genesis json.RawMessage) (modules.UtilityModule, error) {
	return &UtilityModule{
		// TODO: Add `maxTransactionBytes` and `maxTransactions` to cfg.Utility
		Mempool: types.NewMempool(1000, 1000),
	}, nil
}

func InitGenesis(data json.RawMessage) {
	// TODO (Team) add genesis state if necessary
}

func InitConfig(data json.RawMessage) (config *types.UtilityConfig, err error) {
	// TODO (Team) add config if necessary
	config = new(types.UtilityConfig)
	err = json.Unmarshal(data, config)
	return
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
