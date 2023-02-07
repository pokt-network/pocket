package p2p

import (
	"testing"

	typesP2P "github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/stretchr/testify/require"
)

func Test_getAddrBookDelta(t *testing.T) {
	type args struct {
		before typesP2P.AddrBook
		after  typesP2P.AddrBook
	}
	tests := []struct {
		name        string
		args        args
		wantAdded   []*typesP2P.NetworkPeer
		wantRemoved []*typesP2P.NetworkPeer
	}{
		{
			name: "empty slices should return empty slices",
			args: args{
				before: []*typesP2P.NetworkPeer{},
				after:  []*typesP2P.NetworkPeer{},
			},
			wantAdded:   []*typesP2P.NetworkPeer{},
			wantRemoved: []*typesP2P.NetworkPeer{},
		},
		{
			name: "when a peer is added, it should be in the added slice",
			args: args{
				before: []*typesP2P.NetworkPeer{},
				after:  []*typesP2P.NetworkPeer{{Address: crypto.AddressFromString("000000000000000000000000000000000001")}},
			},
			wantAdded:   []*typesP2P.NetworkPeer{{Address: crypto.AddressFromString("000000000000000000000000000000000001")}},
			wantRemoved: []*typesP2P.NetworkPeer{},
		},
		{
			name: "when a peer is removed, it should be in the removed slice",
			args: args{
				before: []*typesP2P.NetworkPeer{{Address: crypto.AddressFromString("000000000000000000000000000000000001")}},
				after:  []*typesP2P.NetworkPeer{},
			},
			wantAdded:   []*typesP2P.NetworkPeer{},
			wantRemoved: []*typesP2P.NetworkPeer{{Address: crypto.AddressFromString("000000000000000000000000000000000001")}},
		},
		{
			name: "when no peers are added or removed, both slices should be empty",
			args: args{
				before: []*typesP2P.NetworkPeer{{Address: crypto.AddressFromString("000000000000000000000000000000000001")}, {Address: crypto.AddressFromString("000000000000000000000000000000000002")}},
				after:  []*typesP2P.NetworkPeer{{Address: crypto.AddressFromString("000000000000000000000000000000000001")}, {Address: crypto.AddressFromString("000000000000000000000000000000000002")}},
			},
			wantAdded:   []*typesP2P.NetworkPeer{},
			wantRemoved: []*typesP2P.NetworkPeer{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAdded, gotRemoved := getAddrBookDelta(tt.args.before, tt.args.after)
			require.ElementsMatch(t, gotAdded, tt.wantAdded)
			require.ElementsMatch(t, gotRemoved, tt.wantRemoved)
		})
	}
}
