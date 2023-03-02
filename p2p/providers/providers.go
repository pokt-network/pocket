package providers

import (
	"github.com/pokt-network/pocket/p2p/providers/current_height_provider"
	"github.com/pokt-network/pocket/p2p/providers/peerstore_provider"
)

type PeerstoreProvider = peerstore_provider.PeerstoreProvider
type CurrentHeightProvider = current_height_provider.CurrentHeightProvider
