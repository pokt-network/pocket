package host

import (
	"errors"
	"time"

	"github.com/pokt-network/pocket/ibc/store"
	"github.com/pokt-network/pocket/runtime/configs"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
)

var _ modules.IBCHostSubmodule = &ibcHost{}

type ibcHost struct {
	base_modules.IntegrableModule

	cfg    *configs.IBCHostConfig
	logger *modules.Logger
}

func Create(bus modules.Bus, config *configs.IBCHostConfig, options ...modules.IBCHostOption) (modules.IBCHostSubmodule, error) {
	return new(ibcHost).Create(bus, config, options...)
}

// WithLogger assigns a logger for the IBC host
func WithLogger(logger *modules.Logger) modules.IBCHostOption {
	return func(m modules.IBCHostSubmodule) {
		if mod, ok := m.(*ibcHost); ok {
			mod.logger = logger
		}
	}
}

func (*ibcHost) Create(bus modules.Bus, config *configs.IBCHostConfig, options ...modules.IBCHostOption) (modules.IBCHostSubmodule, error) {
	h := &ibcHost{
		cfg: config,
	}
	for _, option := range options {
		option(h)
	}
	h.logger.Info().Msg("🛰️ Creating IBC host 🛰️")
	bus.RegisterModule(h)
	_, err := store.Create(h.GetBus(),
		h.cfg.BulkStoreCacher,
		store.WithLogger(h.logger),
		store.WithStoresDir(h.cfg.StoresDir),
		store.WithPrivateKey(h.cfg.PrivateKey),
	)
	if err != nil {
		return nil, err
	}
	return h, nil
}

func (h *ibcHost) GetModuleName() string { return modules.IBCHostSubmoduleName }

// GetTimestamp returns the current unix timestamp
func (h *ibcHost) GetTimestamp() uint64 {
	return uint64(time.Now().Unix())
}

// GetProvableStore returns an instance of a provable store with the given name with the
// CommitmentPrefix set to []byte(name). The store is created if it does not exist. Any changes
// made using the store are handled locally and propagated through the bus, added to all nodes'
// mempools ready for inclusion in the next block to transition the IBC store state tree.
// Any operations will ensure the CommitmentPrefix is prepended to the key if not present already.
func (h *ibcHost) GetProvableStore(name string) (modules.ProvableStore, error) {
	if err := h.GetBus().GetBulkStoreCacher().AddStore(name); err != nil && !errors.Is(err, coreTypes.ErrIBCStoreAlreadyExists(name)) {
		return nil, err
	}
	provableStore, err := h.GetBus().GetBulkStoreCacher().GetStore(name)
	if err != nil {
		return nil, err
	}
	return provableStore, nil
}
