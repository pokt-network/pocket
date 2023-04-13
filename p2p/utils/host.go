package utils

import (
	"context"
	"fmt"
	"time"

	libp2pHost "github.com/libp2p/go-libp2p/core/host"
	"go.uber.org/multierr"

	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/p2p/protocol"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
)

const (
	week = time.Hour * 24 * 7
	// TECHDEBT(#629): consider more carefully and parameterize.
	defaultPeerTTL = 2 * week
)

// PopulateLibp2pHost iterates through peers in given `pstore`, converting peer
// info for use with libp2p and adding it to the underlying libp2p host's peerstore.
// (see: https://pkg.go.dev/github.com/libp2p/go-libp2p@v0.26.2/core/host#Host)
// (see: https://pkg.go.dev/github.com/libp2p/go-libp2p@v0.26.2/core/peerstore#Peerstore)
func PopulateLibp2pHost(host libp2pHost.Host, pstore typesP2P.Peerstore) (err error) {
	for _, peer := range pstore.GetPeerList() {
		if addErr := AddPeerToLibp2pHost(host, peer); addErr != nil {
			err = multierr.Append(err, addErr)
		}
	}
	return err
}

// AddPeerToLibp2pHost covnerts the given pocket peer for use with libp2p and adds
// it to the given libp2p host's underlying peerstore.
func AddPeerToLibp2pHost(host libp2pHost.Host, peer typesP2P.Peer) error {
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

	host.Peerstore().AddAddrs(libp2pPeer.ID, libp2pPeer.Addrs, defaultPeerTTL)
	if err := host.Peerstore().AddPubKey(libp2pPeer.ID, pubKey); err != nil {
		return fmt.Errorf(
			"adding peer public key, pokt address: %s: %w",
			peer.GetAddress(),
			err,
		)
	}
	return nil
}

// RemovePeerFromLibp2pHost removes the given peer's libp2p public keys and
// protocols from the libp2p host's underlying peerstore.
func RemovePeerFromLibp2pHost(host libp2pHost.Host, peer typesP2P.Peer) error {
	peerInfo, err := Libp2pAddrInfoFromPeer(peer)
	if err != nil {
		return err
	}

	host.Peerstore().RemovePeer(peerInfo.ID)
	return host.Peerstore().RemoveProtocols(peerInfo.ID)
}

// Libp2pSendToPeer sends data to the given pocket peer from the given libp2p host.
func Libp2pSendToPeer(host libp2pHost.Host, data []byte, peer typesP2P.Peer) error {
	// TECHDEBT(#595): add ctx to interface methods and propagate down.
	ctx := context.TODO()

	peerInfo, err := Libp2pAddrInfoFromPeer(peer)
	if err != nil {
		return err
	}

	// debug logging: network resource scope stats
	// (see: https://pkg.go.dev/github.com/libp2p/go-libp2p@v0.27.0/core/network#ResourceManager)
	// (see: https://pkg.go.dev/github.com/libp2p/go-libp2p@v0.27.0/core/network#ResourceScopeViewer)
	logScope := LogScopeStatFactory(&logger.Global.Logger, "host transient resource scope")
	err = host.Network().ResourceManager().ViewTransient(logScope)
	if err != nil {
		logger.Global.Debug().Err(err).Msg("logging resource scope stats")
	}

	stream, err := host.NewStream(ctx, peerInfo.ID, protocol.PoktProtocolID)
	if err != nil {
		return fmt.Errorf("opening stream: %w", err)
	}

	if n, err := stream.Write(data); err != nil {
		return multierr.Append(
			fmt.Errorf("writing to stream: %w", err),
			stream.Reset(),
		)
	} else {
		logger.Global.Debug().Int("bytes", n).Msg("written to peer stream")
	}

	// MUST USE `streamClose` NOT `stream.CloswWrite`; otherwise, outbound streams
	// will accumulate until resource limits are hit; e.g.:
	// > "opening stream: stream-3478: transient: cannot reserve outbound stream: resource limit exceeded"
	return stream.Close()
}
