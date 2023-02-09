package identity

import (
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"

	"github.com/pokt-network/pocket/p2p/common"
	"github.com/pokt-network/pocket/p2p/transport"
	"github.com/pokt-network/pocket/p2p/types"
	poktCrypto "github.com/pokt-network/pocket/shared/crypto"
)

var (
	ErrIdentity = common.NewErrFactory("")
)

func PoktPeerFromStream(stream network.Stream) (*types.NetworkPeer, error) {
	remotePubKeyBytes, err := stream.Conn().RemotePublicKey().Raw()
	// NB: abort handling this stream.
	if err != nil {
		return nil, err
	}
	poktPubKey, err := poktCrypto.NewPublicKeyFromBytes(remotePubKeyBytes)
	if err != nil {
		return nil, err
	}

	return &types.NetworkPeer{
		Dialer:    transport.NewLibP2PTransport(stream),
		PublicKey: poktPubKey,
		// NB: pokt analogue of libp2p peer.ID
		Address: poktPubKey.Address(),
	}, nil
}

func PeerAddrInfoFromPoktPeer(poktPeer *types.NetworkPeer) (peer.AddrInfo, error) {
	pubKey, err := crypto.UnmarshalEd25519PublicKey(poktPeer.PublicKey.Bytes())
	if err != nil {
		return peer.AddrInfo{}, ErrIdentity("unable to unmarshal peer ed25519 public key", err)
	}

	peerID, err := peer.IDFromPublicKey(pubKey)
	if err != nil {
		return peer.AddrInfo{}, ErrIdentity("unable to retrieve ID from peer public key", err)
	}

	peerMultiaddr, err := multiaddr.NewMultiaddr(poktPeer.ServiceUrl)
	if err != nil {
		return peer.AddrInfo{}, err
	}

	return peer.AddrInfo{
		ID: peerID,
		Addrs: []multiaddr.Multiaddr{
			peerMultiaddr,
		},
	}, nil
}
