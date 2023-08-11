package types

import (
	"encoding/hex"
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"

	poktCrypto "github.com/pokt-network/pocket/shared/crypto"
)

func TestPeerAddrMap_AddPeer(t *testing.T) {
	tests := []struct {
		name     string
		peerArg  Peer
		wantErr  bool
		wantSize int
	}{
		{
			name:     "adds peer",
			peerArg:  newTestPeer(999),
			wantErr:  false,
			wantSize: 5,
		},
		{
			name:     "returns error if peer exists",
			peerArg:  newTestPeer(1),
			wantErr:  true,
			wantSize: 4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testPStore := newTestPeers(4)

			err := testPStore.AddPeer(tt.peerArg)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.wantSize, testPStore.Size())
		})
	}
}

func TestPeerAddrMap_GetPeer(t *testing.T) {
	tests := []struct {
		name     string
		wantPeer Peer
		addrArg  poktCrypto.Address
	}{
		{
			name:     "returns peer",
			wantPeer: newTestPeer(1),
			addrArg:  getTestAddr(1),
		},
		{
			name:     "returns nil if peer not found",
			addrArg:  getTestAddr(999),
			wantPeer: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testPStore := newTestPeers(4)

			gotPeer := testPStore.GetPeer(tt.addrArg)
			require.EqualValues(t, tt.wantPeer, gotPeer)
		})
	}
}

func newTestPeers(count int) Peerstore {
	pstore := make(PeerAddrMap)

	for i := 0; i < count; i++ {
		peer := newTestPeer(i + 1)
		// NB: intentionally not using `PeerAddrMap#AddPeer()` to isolate the test.
		pstore[peer.GetAddress().String()] = peer
	}
	return pstore
}

func newTestPeer(idx int) Peer {
	return &NetworkPeer{
		Address: getTestAddr(idx),
	}
}

func getTestAddr(idx int) poktCrypto.Address {
	return poktCrypto.AddressFromString(hex.EncodeToString(
		[]byte(fmt.Sprintf("pokt%.3d", idx)),
	))
}
