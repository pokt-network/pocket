package config

import (
	"fmt"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/protocol"
	"go.uber.org/multierr"

	typesP2P "github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
)

var (
	_ typesP2P.RouterConfig = &baseConfig{}
	_ typesP2P.RouterConfig = &UnicastRouterConfig{}
	_ typesP2P.RouterConfig = &BackgroundConfig{}
	_ typesP2P.RouterConfig = &RainTreeConfig{}
)

// baseConfig implements `RouterConfig` using the given libp2p host, pokt address
// and handler function. Intended for internal use by other `RouterConfig`
// implementations with common config parameters.
//
// NB: intentionally *not* embedding `baseConfig` to improve readability of usages
// of would-be embedders (e.g. `BackgroundConfig`).
type baseConfig struct {
	Host    host.Host
	Addr    crypto.Address
	Handler func(data []byte) error
}

type UnicastRouterConfig struct {
	Logger         *modules.Logger
	Host           host.Host
	ProtocolID     protocol.ID
	MessageHandler typesP2P.MessageHandler
	PeerHandler    func(peer typesP2P.Peer) error
}

// BackgroundConfig implements `RouterConfig` for use with `BackgroundRouter`.
type BackgroundConfig struct {
	Host    host.Host
	Addr    crypto.Address
	Handler func(data []byte) error
}

// RainTreeConfig implements `RouterConfig` for use with `RainTreeRouter`.
type RainTreeConfig struct {
	Host    host.Host
	Addr    crypto.Address
	Handler func(data []byte) error
}

// IsValid implements the respective member of the `RouterConfig` interface.
func (cfg *baseConfig) IsValid() (err error) {
	if cfg.Addr == nil {
		err = multierr.Append(err, fmt.Errorf("pokt address not configured"))
	}

	if cfg.Host == nil {
		err = multierr.Append(err, fmt.Errorf("host not configured"))
	}

	if cfg.Handler == nil {
		err = multierr.Append(err, fmt.Errorf("handler not configured"))
	}
	return err
}

// IsValid implements the respective member of the `RouterConfig` interface.
func (cfg *UnicastRouterConfig) IsValid() (err error) {
	if cfg.Logger == nil {
		err = multierr.Append(err, fmt.Errorf("logger not configured"))
	}

	if cfg.Host == nil {
		err = multierr.Append(err, fmt.Errorf("host not configured"))
	}

	if cfg.ProtocolID == "" {
		err = multierr.Append(err, fmt.Errorf("protocol id not configured"))
	}

	if cfg.MessageHandler == nil {
		err = multierr.Append(err, fmt.Errorf("message handler not configured"))
	}

	if cfg.PeerHandler == nil {
		err = multierr.Append(err, fmt.Errorf("peer handler not configured"))
	}
	return err
}

// IsValid implements the respective member of the `RouterConfig` interface.
func (cfg *BackgroundConfig) IsValid() error {
	baseCfg := baseConfig{
		Host:    cfg.Host,
		Addr:    cfg.Addr,
		Handler: cfg.Handler,
	}
	return baseCfg.IsValid()
}

// IsValid implements the respective member of the `RouterConfig` interface.
func (cfg *RainTreeConfig) IsValid() error {
	baseCfg := baseConfig{
		Host:    cfg.Host,
		Addr:    cfg.Addr,
		Handler: cfg.Handler,
	}
	return baseCfg.IsValid()
}
