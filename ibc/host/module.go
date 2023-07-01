package host

import (
	"errors"
	"fmt"
	"time"

	"github.com/pokt-network/pocket/ibc/store"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
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

	privateKey string
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

// WithPrivateKey assigns the IBC host's private key
func WithPrivateKey(pkHex string) modules.IBCHostOption {
	return func(m modules.IBCHostModule) {
		if mod, ok := m.(*ibcHost); ok {
			mod.privateKey = pkHex
		}
	}
}

func (*ibcHost) Create(bus modules.Bus, options ...modules.IBCHostOption) (modules.IBCHostModule, error) {
	h := &ibcHost{}
	for _, option := range options {
		option(h)
	}
	h.logger.Info().Msg("🛰️ creating IBC host 🛰️")
	bus.RegisterModule(h)
	if h.storesDir == "" {
		return nil, fmt.Errorf("stores directory not set")
	}
	sm := store.NewStoreManager(h.bus, h.storesDir, h.privateKey)
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

// GetProvableStore returns an instance of a provable store with the given name with the
// CommitmentPrefix set to []byte(name). The store is created if it does not exist. Any changes
// made using the store are handled locally and propagated through the bus, added to all nodes'
// mempools ready for inclusion in the next block to transition the IBC store state tree.
// Any operations will ensure the CommitmentPrefix is prepended to the key if not present already.
func (h *ibcHost) GetProvableStore(name string) (modules.ProvableStore, error) {
	if err := h.storeManager.AddStore(name); err != nil && !errors.Is(err, coreTypes.ErrIBCStoreAlreadyExists(name)) {
		return nil, err
	}
	provableStore, err := h.storeManager.GetStore(name)
	if err != nil {
		return nil, err
	}
	return provableStore, nil
}
