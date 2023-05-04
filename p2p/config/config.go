package config

import (
	"fmt"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/pokt-network/pocket/p2p/providers"
	"github.com/pokt-network/pocket/shared/crypto"
	"go.uber.org/multierr"
)

// RainTreeConfig implements `RouterConfig` for use with `RainTreeRouter`
type RainTreeConfig struct {
	Addr                  crypto.Address
	CurrentHeightProvider providers.CurrentHeightProvider
	Host                  host.Host
	Hostname              string
	MaxNonces             uint64
	PeerstoreProvider     providers.PeerstoreProvider
}

func (cfg *RainTreeConfig) IsValid() (err error) {
	// TECHDEBT: can `Hostname` or `MaxNonces` be invalid?

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
	return err
}
