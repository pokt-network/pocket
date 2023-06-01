package config

import (
	"fmt"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/pokt-network/pocket/p2p/providers"
	"github.com/pokt-network/pocket/shared/crypto"
	"go.uber.org/multierr"
)

// baseConfig implements `RouterConfig` using the given libp2p host and current
// height and peerstore providers. Intended for internal use by other `RouterConfig`
// implementations with common config parameters.
//
// NB: intentionally *not* embedding `baseConfig` to improve readability of usages
// of would-be embedders (e.g. `BackgroundConfig`).
type baseConfig struct {
	Host                  host.Host
	Addr                  crypto.Address
	CurrentHeightProvider providers.CurrentHeightProvider
	PeerstoreProvider     providers.PeerstoreProvider
	Handler               func(data []byte) error
}

// BackgroundConfig implements `RouterConfig` for use with `BackgroundRouter`.
type BackgroundConfig struct {
	Host                  host.Host
	Addr                  crypto.Address
	CurrentHeightProvider providers.CurrentHeightProvider
	PeerstoreProvider     providers.PeerstoreProvider
	Handler               func(data []byte) error
}

// RainTreeConfig implements `RouterConfig` for use with `RainTreeRouter`.
type RainTreeConfig struct {
	Host                  host.Host
	Addr                  crypto.Address
	CurrentHeightProvider providers.CurrentHeightProvider
	PeerstoreProvider     providers.PeerstoreProvider
	Handler               func(data []byte) error
}

// IsValid implements the respective member of the `RouterConfig` interface.
func (cfg *baseConfig) IsValid() (err error) {
	if cfg.Addr == nil {
		err = multierr.Append(err, fmt.Errorf("pokt address not configured"))
	}

	if cfg.CurrentHeightProvider == nil {
		err = multierr.Append(err, fmt.Errorf("current height provider not configured"))
	}

	if cfg.Host == nil {
		err = multierr.Append(err, fmt.Errorf("host not configured"))
	}

	if cfg.PeerstoreProvider == nil {
		err = multierr.Append(err, fmt.Errorf("peerstore provider not configured"))
	}

	if cfg.Handler == nil {
		err = multierr.Append(err, fmt.Errorf("handler not configured"))
	}
	return err
}

// IsValid implements the respective member of the `RouterConfig` interface.
func (cfg *BackgroundConfig) IsValid() (err error) {
	baseCfg := baseConfig{
		Host:                  cfg.Host,
		Addr:                  cfg.Addr,
		CurrentHeightProvider: cfg.CurrentHeightProvider,
		PeerstoreProvider:     cfg.PeerstoreProvider,
		Handler:               cfg.Handler,
	}
	return multierr.Append(err, baseCfg.IsValid())
}

// IsValid implements the respective member of the `RouterConfig` interface.
func (cfg *RainTreeConfig) IsValid() (err error) {
	baseCfg := baseConfig{
		Host:                  cfg.Host,
		Addr:                  cfg.Addr,
		CurrentHeightProvider: cfg.CurrentHeightProvider,
		PeerstoreProvider:     cfg.PeerstoreProvider,
		Handler:               cfg.Handler,
	}
	return multierr.Append(err, baseCfg.IsValid())
}
