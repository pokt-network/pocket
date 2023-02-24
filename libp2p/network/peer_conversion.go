package network

import (
	"fmt"

	libp2pCrypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/network"
	libp2pPeer "github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"

	"github.com/pokt-network/pocket/libp2p/transport"
	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/shared/crypto"
)

// PeerFromLibp2pStream builds a network peer using peer info available
// from the given libp2p stream.
func PeerFromLibp2pStream(stream network.Stream) (*types.NetworkPeer, error) {
	publicKeyBz, err := stream.Conn().RemotePublicKey().Raw()
	if err != nil {
		return nil, err
	}
	publicKey, err := crypto.NewPublicKeyFromBytes(publicKeyBz)
	if err != nil {
		return nil, err
	}

	peerMultiaddr := stream.Conn().RemoteMultiaddr()
	peerServiceUrl, err := ServiceUrlFromLibp2pMultiaddr(peerMultiaddr)
	if err != nil {
		return nil, fmt.Errorf("converting multiaddr to service URL: %w", err)
	}

	return &types.NetworkPeer{
		Dialer:     transport.NewLibP2PTransport(stream),
		PublicKey:  publicKey,
		Address:    publicKey.Address(),
		Multiaddr:  peerMultiaddr,
		ServiceUrl: peerServiceUrl,
	}, nil
}

// Libp2pPublicKeyFromPeer retrieves the libp2p compatible public key from a pocket peer.
func Libp2pPublicKeyFromPeer(peer *types.NetworkPeer) (libp2pCrypto.PubKey, error) {
	publicKey, err := libp2pCrypto.UnmarshalEd25519PublicKey(peer.PublicKey.Bytes())
	if err != nil {
		return nil, fmt.Errorf(
			"unmarshalling peer ed25519 public key, pokt address: %s: %w",
			peer.Address,
			err,
		)
	}
	return publicKey, nil
}

// Libp2pAddrInfoFromPeer builds a libp2p AddrInfo which maps to the passed pocket peer.
func Libp2pAddrInfoFromPeer(peer *types.NetworkPeer) (libp2pPeer.AddrInfo, error) {
	publicKey, err := Libp2pPublicKeyFromPeer(peer)
	if err != nil {
		return libp2pPeer.AddrInfo{}, err
	}

	peerID, err := libp2pPeer.IDFromPublicKey(publicKey)
	if err != nil {
		return libp2pPeer.AddrInfo{}, fmt.Errorf(
			"retrieving ID from peer public key, pokt address: %s: %w",
			peer.Address,
			err,
		)
	}

	logger.Global.Info().Str("serviceUrl", peer.ServiceUrl).Msg("requesting peerMultiaddr in peer_conversion.go")
	peerMultiaddr, err := Libp2pMultiaddrFromServiceUrl(peer.ServiceUrl)
	if err != nil {
		return libp2pPeer.AddrInfo{}, err
	}

	return libp2pPeer.AddrInfo{
		ID: peerID,
		Addrs: []multiaddr.Multiaddr{
			peerMultiaddr,
		},
	}, nil
}
