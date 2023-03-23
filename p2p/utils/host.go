package utils

import (
	"context"
	"fmt"
	"time"

	libp2pHost "github.com/libp2p/go-libp2p/core/host"
	"go.uber.org/multierr"

	"github.com/pokt-network/pocket/p2p/protocol"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
)

const (
	week = time.Hour * 24 * 7
	// TECHDEBT: consider more carefully and parameterize.
	// TECHDEBT: unexport after consolidation of P2P modules
	DefaultPeerTTL = 2 * week
)

// PopulateLibp2pHost iterates through peers in given `pstore`, converting peer
// info for use with libp2p and adding it to the underlying libp2p host's peerstore.
// (see: https://pkg.go.dev/github.com/libp2p/go-libp2p@v0.26.2/core/host#Host)
// (see: https://pkg.go.dev/github.com/libp2p/go-libp2p@v0.26.2/core/peerstore#Peerstore)
func PopulateLibp2pHost(host libp2pHost.Host, pstore typesP2P.Peerstore) error {
	for _, peer := range pstore.GetPeerList() {
		pubKey, err := Libp2pPublicKeyFromPeer(peer)
		if err != nil {
			return fmt.Errorf(
				"converting peer public key, pokt address: %s: %w",
				peer.GetAddress(),
				err,
			)
		}
		libp2pPeer, err := Libp2pAddrInfoFromPeer(peer)
		if err != nil {
			return fmt.Errorf(
				"converting peer info, pokt address: %s: %w",
				peer.GetAddress(),
				err,
			)
		}

		host.Peerstore().AddAddrs(libp2pPeer.ID, libp2pPeer.Addrs, DefaultPeerTTL)
		if err := host.Peerstore().AddPubKey(libp2pPeer.ID, pubKey); err != nil {
			return fmt.Errorf(
				"adding peer public key, pokt address: %s: %w",
				peer.GetAddress(),
				err,
			)
		}
	}
	return nil
}

func Libp2pSendToPeer(host libp2pHost.Host, data []byte, peer typesP2P.Peer) error {
	// TECHDEBT(#595): add ctx to interface methods and propagate down.
	ctx := context.Background()

	peerInfo, err := Libp2pAddrInfoFromPeer(peer)
	if err != nil {
		return err
	}

	stream, err := host.NewStream(ctx, peerInfo.ID, protocol.PoktProtocolID)
	if err != nil {
		return fmt.Errorf("opening stream: %w", err)
	}

	if _, err = stream.Write(data); err != nil {
		return multierr.Append(
			fmt.Errorf("writing to stream: %w", err),
			stream.Reset(),
		)
	}

	return stream.CloseWrite()
}
