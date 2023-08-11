package utils

import (
	"fmt"

	libp2pCrypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/network"
	libp2pPeer "github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/pokt-network/pocket/p2p/types"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/shared/crypto"
)

// PeerFromLibp2pStream builds a network peer using peer info available
// from the given libp2p stream.
func PeerFromLibp2pStream(stream network.Stream) (typesP2P.Peer, error) {
	addrInfo := libp2pPeer.AddrInfo{
		ID:    stream.Conn().RemotePeer(),
		Addrs: []multiaddr.Multiaddr{stream.Conn().RemoteMultiaddr()},
	}
	return PeerFromLibp2pAddrInfo(&addrInfo)
}

func PeerFromLibp2pAddrInfo(addrInfo *libp2pPeer.AddrInfo) (typesP2P.Peer, error) {
	publicKeyErr := func(err error) error {
		return fmt.Errorf(
			"error converting public key from libp2p peer ID %q: %w",
			addrInfo.ID, err,
		)
	}

	libP2PPublicKey, err := addrInfo.ID.ExtractPublicKey()
	if err != nil {
		return nil, publicKeyErr(err)
	}

	publicKeyBz, err := libP2PPublicKey.Raw()
	if err != nil {
		return nil, publicKeyErr(err)
	}

	publicKey, err := crypto.NewPublicKeyFromBytes(publicKeyBz)
	if err != nil {
		return nil, publicKeyErr(err)
	}

	var (
		peerServiceURL string
		peerMultiaddr  multiaddr.Multiaddr
	)
	if len(addrInfo.Addrs) > 0 {
		// NOTE: only using the first multiaddr.
		peerMultiaddr = addrInfo.Addrs[0]
		peerServiceURL, err = ServiceURLFromLibp2pMultiaddr(addrInfo.Addrs[0])
		if err != nil {
			return nil, fmt.Errorf("error converting multiaddr to service URL: %w", err)
		}
	}

	// NOTE: only using the first multiaddr.
	return &types.NetworkPeer{
		PublicKey:  publicKey,
		Address:    publicKey.Address(),
		Multiaddr:  peerMultiaddr,
		ServiceURL: peerServiceURL,
	}, nil
}

// Libp2pPublicKeyFromPeer retrieves the libp2p compatible public key from a pocket peer.
func Libp2pPublicKeyFromPeer(peer typesP2P.Peer) (libp2pCrypto.PubKey, error) {
	publicKey, err := libp2pCrypto.UnmarshalEd25519PublicKey(peer.GetPublicKey().Bytes())
	if err != nil {
		return nil, fmt.Errorf(
			"unmarshalling peer ed25519 public key, pokt address: %s: %w",
			peer.GetAddress(),
			err,
		)
	}
	return publicKey, nil
}

// Libp2pAddrInfoFromPeer builds a libp2p AddrInfo which maps to the passed pocket peer.
func Libp2pAddrInfoFromPeer(peer typesP2P.Peer) (libp2pPeer.AddrInfo, error) {
	publicKey, err := Libp2pPublicKeyFromPeer(peer)
	if err != nil {
		return libp2pPeer.AddrInfo{}, err
	}

	peerID, err := libp2pPeer.IDFromPublicKey(publicKey)
	if err != nil {
		return libp2pPeer.AddrInfo{}, fmt.Errorf(
			"retrieving ID from peer public key, pokt address: %s: %w",
			peer.GetAddress(),
			err,
		)
	}

	peerMultiaddr, err := Libp2pMultiaddrFromServiceURL(peer.GetServiceURL())
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
