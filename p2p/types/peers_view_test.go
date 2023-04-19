package types

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
)

func TestSortedPeersView_Add_Remove(t *testing.T) {
	testCases := []struct {
		name          string
		selfAddr      string
		addAddrs      string
		removeAddrs   string
		expectedAddrs string
	}{
		{
			name:          "lowest self address",
			selfAddr:      "A",
			addAddrs:      "BC",
			expectedAddrs: "ABC",
		},
		{
			name:          "highest self address",
			selfAddr:      "C",
			addAddrs:      "AB",
			expectedAddrs: "CAB",
		},
		{
			name:          "penultimate self address",
			selfAddr:      "W",
			addAddrs:      "DYZEHGI",
			removeAddrs:   "YE",
			expectedAddrs: "WZDGHI",
		},
		{
			name:          "middle self address",
			selfAddr:      "S",
			addAddrs:      "DTUEVGH",
			removeAddrs:   "E",
			expectedAddrs: "STUVDGH",
		},
		{
			name:          "discontiguous resulting addresses",
			selfAddr:      "O",
			addAddrs:      "DTURSEVGH",
			removeAddrs:   "EDU",
			expectedAddrs: "ORSTVGH",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			selfAddr := cryptoPocket.Address(testCase.selfAddr)
			selfPeer := &NetworkPeer{Address: selfAddr}

			pstore := make(PeerAddrMap)
			err := pstore.AddPeer(selfPeer)
			require.NoError(t, err)

			view := NewSortedPeersView(selfAddr, pstore)
			initialAddrs := []string{selfAddr.String()}
			initialPeers := []Peer{selfPeer}

			// assert initial state
			require.ElementsMatchf(t, initialAddrs, view.sortedAddrs, "initial addresses don't match")
			require.ElementsMatchf(t, initialPeers, view.sortedPeers, "initial peers don't match")

			addrsToAdd := strings.Split(testCase.addAddrs, "")
			addrsToRemove := strings.Split(testCase.removeAddrs, "")
			expectedAddrs := fromCharAddrs(testCase.expectedAddrs)

			// add peers
			for _, addr := range addrsToAdd {
				peer := &NetworkPeer{Address: []byte(addr)}
				view.Add(peer)
				t.Logf("sortedAddrs: %s", toCharAddrs(view.sortedAddrs))
			}

			// remove peers
			for _, addr := range addrsToRemove {
				view.Remove([]byte(addr))
				t.Logf("sortedAddrs: %s", toCharAddrs(view.sortedAddrs))
			}

			// assert resulting state
			var expectedPeers []Peer
			for _, addr := range expectedAddrs {
				expectedPeers = append(expectedPeers, &NetworkPeer{Address: cryptoPocket.AddressFromString(addr)})
			}

			actualAddrsStr := toCharAddrs(view.sortedAddrs)
			require.Equal(t, testCase.expectedAddrs, actualAddrsStr, "resulting addresses don't match")
			require.ElementsMatchf(t, expectedPeers, view.sortedPeers, "resulting peers don't match")
		})
	}
}

func TestSortedPeersView_Remove(t *testing.T) {
	t.Skip("TECHDEBT(#554): test that this method works as expected when target peer/addr is not in the list!")
}

func TestSortedPeersView(t *testing.T) {
	testCases := []struct {
		name              string
		selfAddr          string
		initialAddrs      string
		expectedSortOrder string
	}{
		{
			name:              "lowest self address",
			selfAddr:          "A",
			initialAddrs:      "BDCEA",
			expectedSortOrder: "ABCDE",
		},
		{
			name:              "highest self address",
			selfAddr:          "E",
			initialAddrs:      "BDCEA",
			expectedSortOrder: "EABCD",
		},
		{
			name:              "middle self address",
			selfAddr:          "C",
			initialAddrs:      "BDCEA",
			expectedSortOrder: "CDEAB",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			selfAddr := cryptoPocket.Address(testCase.selfAddr)

			pstore := make(PeerAddrMap)
			var initialPeers []Peer
			initialAddrs := fromCharAddrs(testCase.initialAddrs)
			for _, addr := range initialAddrs {
				peer := &NetworkPeer{Address: cryptoPocket.AddressFromString(addr)}
				initialPeers = append(initialPeers, peer)
				err := pstore.AddPeer(peer)
				require.NoError(t, err)
			}

			view := NewSortedPeersView(selfAddr, pstore)
			require.ElementsMatchf(t, initialAddrs, view.sortedAddrs, "initial addresses don't match")
			require.ElementsMatchf(t, initialPeers, view.sortedPeers, "initial peers don't match")

			var actualPeersStr string
			for _, peer := range view.sortedPeers {
				actualPeersStr += string(peer.GetAddress().Bytes())
			}

			actualAddrsStr := toCharAddrs(view.sortedAddrs)
			require.Equal(t, testCase.expectedSortOrder, actualAddrsStr, "resulting addresses don't match")
			require.Equal(t, testCase.expectedSortOrder, actualPeersStr, "resulting addresses don't match")
		})
	}
}

// fromCharAddrs converts each char in charAddrs into a serialized pokt address
func fromCharAddrs(charAddrs string) (addrs []string) {
	for _, ch := range strings.Split(charAddrs, "") {
		addrs = append(addrs, cryptoPocket.Address(ch).String())
	}
	return addrs
}

// toCharAddrs converts each string in addrStrs to a raw pokt address binary
// string and concatenates them for return
func toCharAddrs(addrStrs []string) (charAddrs string) {
	for _, addrStr := range addrStrs {
		charAddrs += string(cryptoPocket.AddressFromString(addrStr).Bytes())
	}
	return charAddrs
}
