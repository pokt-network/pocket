package host

import (
	"fmt"
	"time"

	"github.com/pokt-network/pocket/ibc/store"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
)

var _ modules.IBCHostModule = &ibcHost{}

type ibcHost struct {
	base_modules.IntegrableModule

	bus modules.Bus

	logger       *modules.Logger
	storesDir    string
	storeManager modules.IBCStoreManager
}

func Create(bus modules.Bus, options ...modules.IBCHostOption) (modules.IBCHostModule, error) {
	return new(ibcHost).Create(bus, options...)
}

// WithLogger assigns a logger for the IBC host
func WithLogger(logger *modules.Logger) modules.IBCHostOption {
	return func(m modules.IBCHostModule) {
		if mod, ok := m.(*ibcHost); ok {
			mod.logger = logger
		}
	}
}

// WithStoresDir assigns the IBC host's stores directory
func WithStoresDir(dir string) modules.IBCHostOption {
	return func(m modules.IBCHostModule) {
		if mod, ok := m.(*ibcHost); ok {
			mod.storesDir = fmt.Sprintf("%s/host", dir)
		}
	}
}

func (*ibcHost) Create(bus modules.Bus, options ...modules.IBCHostOption) (modules.IBCHostModule, error) {
	h := &ibcHost{}
	for _, option := range options {
		option(h)
	}
	h.logger.Info().Msg("🛰️ creating IBC host 🛰️")
	if h.storesDir == "" {
		return nil, fmt.Errorf("stores directory not set")
	}
	sm := store.NewStoreManager(h.storesDir)
	h.storeManager = sm
	return h, nil
}

func (h *ibcHost) GetModuleName() string  { return modules.IBCHostModuleName }
func (h *ibcHost) GetBus() modules.Bus    { return h.bus }
func (h *ibcHost) SetBus(bus modules.Bus) { h.bus = bus }

// GetTimestamp returns the current unix timestamp
func (h *ibcHost) GetTimestamp() uint64 {
	return uint64(time.Now().Unix())
}
