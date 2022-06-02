package raintree

import (
	"fmt"
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

	testCases := []struct {
		numNodes          int
		numExpectedLevels int
	}{
		// 0 levels
		{1, 0}, // Just self
		// 1 level
		{2, 1},
		{3, 1},
		// 2 levels
		{4, 2},
		{9, 2},
		// 3 levels
		{10, 3},
		{27, 3},
		// 4 levels
		{28, 4},
		{81, 4},
		// 5 levels
		{82, 5},
		// 10 levels
		{59049, 10},
		// 11 levels
		{59050, 11},
	}

	for _, testCase := range testCases {
		n := testCase.numNodes
		t.Run(fmt.Sprintf("n=%d", n), func(t *testing.T) {
			addrBook := getAddrBook(t, n-1)
			addrBook = append(addrBook, &types.NetworkPeer{Address: addr})
			network := NewRainTreeNetwork(addr, addrBook, cfg).(*rainTreeNetwork)

			err = network.handleAddrBookUpdates()
			require.NoError(t, err)

			require.Equal(t, len(network.addrList), n)
			require.Equal(t, len(network.addrBookMap), n)
			require.Equal(t, int(network.maxNumLevels), testCase.numExpectedLevels)
		})
	}
}

func BenchmarkAddrBookUpdates(b *testing.B) {
	cfg := &config.Config{}

	addr, err := cryptoPocket.GenerateAddress()
	require.NoError(b, err)

	testCases := []struct {
		numNodes          int
		numExpectedLevels int
	}{
		// Small
		{9, 2},
		// Large
		{59049, 10},
	}

	for _, testCase := range testCases {
		n := testCase.numNodes
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			addrBook := getAddrBook(nil, n-1)
			addrBook = append(addrBook, &types.NetworkPeer{Address: addr})
			network := NewRainTreeNetwork(addr, addrBook, cfg).(*rainTreeNetwork)

			err = network.handleAddrBookUpdates()
			require.NoError(b, err)

			require.Equal(b, len(network.addrList), n)
			require.Equal(b, len(network.addrBookMap), n)
			require.Equal(b, int(network.maxNumLevels), testCase.numExpectedLevels)
		})
	}
}

func getAddrBook(t *testing.T, n int) (addrBook types.AddrBook) {
	addrBook = make([]*types.NetworkPeer, 0)
	for i := 0; i < n; i++ {
		addr, err := cryptoPocket.GenerateAddress()
		if t != nil {
			require.NoError(t, err)
		}
		addrBook = append(addrBook, &types.NetworkPeer{Address: addr})
	}
	return
}
