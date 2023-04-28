package utils

import (
	"fmt"

	"github.com/libp2p/go-libp2p/core/host"
	"go.uber.org/multierr"

	"github.com/pokt-network/pocket/p2p/providers"
	"github.com/pokt-network/pocket/shared/crypto"
)

// RouterConfig is used to configure `Router` implementations using the given
// libp2p host and current height and peerstore providers.
// TECHDEBT: I would prefer for this to be in p2p/types/router.go but this causes
// an import cycle between `typesP2P` and `providers`.
type RouterConfig struct {
	Addr                  crypto.Address
	CurrentHeightProvider providers.CurrentHeightProvider
	Host                  host.Host
	Hostname              string
	MaxMempoolCount       uint64
	PeerstoreProvider     providers.PeerstoreProvider
}

func (cfg *RouterConfig) IsValid() (err error) {
	// TECHDEBT: can `Hostname` or `MaxMempoolCount` be invalid?

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
