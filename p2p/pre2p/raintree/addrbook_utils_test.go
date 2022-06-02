package raintree

import (
	"testing"

	"github.com/pokt-network/pocket/p2p/pre2p/types"
	"github.com/pokt-network/pocket/shared/config"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/stretchr/testify/require"
)

func TestRainTreeAddrBookUtilsHandleUpdate(t *testing.T) {
	cfg := &config.Config{}

	addr, err := cryptoPocket.GenerateAddress()
	require.NoError(t, err)

	t.Run("n=8", func(t *testing.T) {
		addrBook := getAddrBook(t, 8)
		addrBook = append(addrBook, &types.NetworkPeer{Address: addr})
		n := NewRainTreeNetwork(addr, addrBook, cfg).(*rainTreeNetwork)
		err = n.handleAddrBookUpdates()
		require.NoError(t, err)
	})
}

func getAddrBook(t *testing.T, n uint8) (addrBook types.AddrBook) {
	addrBook = make([]*types.NetworkPeer, 0)
	for i := uint8(0); i < n; i++ {
		addr, err := cryptoPocket.GenerateAddress()
		require.NoError(t, err)
		addrBook = append(addrBook, &types.NetworkPeer{Address: addr})
	}
	return
}
