package identity

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"

	"github.com/pokt-network/pocket/p2p/transport"
	"github.com/pokt-network/pocket/p2p/types"
	poktCrypto "github.com/pokt-network/pocket/shared/crypto"
)

var (
	ErrIdentity = types.NewErrFactory("")
)

// PoktPeerFromStream builds a pokt peer from a libp2p stream.
// (NOTE: excludes `ServiceURL`)
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
		Address:    poktPubKey.Address(),
		ServiceUrl: stream.Conn().RemoteMultiaddr().String(),
	}, nil
}

// PubKeyFromPoktPeer retrieves the libp2p compatible public key from a pocket peer.
func PubKeyFromPoktPeer(poktPeer *types.NetworkPeer) (crypto.PubKey, error) {
	pubKey, err := crypto.UnmarshalEd25519PublicKey(poktPeer.PublicKey.Bytes())
	if err != nil {
		return nil, ErrIdentity("unable to unmarshal peer ed25519 public key", err)
	}

	return pubKey, nil
}

// PeerAddrInfoFromPoktPeer builds a libp2p AddrInfo which maps to the passed pocket peer.
func PeerAddrInfoFromPoktPeer(poktPeer *types.NetworkPeer) (peer.AddrInfo, error) {
	pubKey, err := PubKeyFromPoktPeer(poktPeer)
	if err != nil {
		return peer.AddrInfo{}, err
	}

	peerID, err := peer.IDFromPublicKey(pubKey)
	if err != nil {
		return peer.AddrInfo{}, ErrIdentity("unable to retrieve ID from peer public key", err)
	}

	peerMultiaddr, err := multiaddr.NewMultiaddr(poktPeer.ServiceUrl)
	// NB: early return if we already have a multiaddr.
	if err == nil {
		return peer.AddrInfo{
			ID: peerID,
			Addrs: []multiaddr.Multiaddr{
				peerMultiaddr,
			},
		}, nil
	}

	peerMultiaddr, err = PeerMultiAddrFromServiceURL(poktPeer.ServiceUrl)
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

func PeerMultiAddrFromServiceURL(serviceURL string) (multiaddr.Multiaddr, error) {
	// TODO: handle case where a scheme is present in the ServiceURL.
	// NB: hard-code a scheme for URL parsing to work.
	peerUrl, err := url.Parse("scheme://" + serviceURL)
	if err != nil {
		return nil, ErrIdentity(fmt.Sprintf(
			"unable to parse peer service URL: %s", serviceURL,
		), err)
	}

	// TODO: parameterize transport.
	var (
		peerTransportStr = "tcp"
		peerHostnameStr  = peerUrl.Hostname()
		// TODO: is there a way for us to effectively prefer IPv6 responses?
		// NB: default to assuming an FQDN-based ServiceURL.
		networkStr = "dns"
	)

	// NB: if ServiceURL is IP address (see: https://pkg.go.dev/net#ParseIP)
	if peerIP := net.ParseIP(peerHostnameStr); peerIP != nil {
		peerHostnameStr = peerIP.String()
		networkStr = "ip4"
		// TODO: there's probably a more conventional way to do this.
		// NB: check if we're dealing with IPv4 or IPv6
		if strings.Count(peerHostnameStr, ":") > 0 {
			networkStr = "ip6"
		}
	}

	peerMultiAddrStr := fmt.Sprintf(
		"/%s/%s/%s/%s",
		networkStr,
		peerHostnameStr,
		peerTransportStr,
		peerUrl.Port(),
	)
	return multiaddr.NewMultiaddr(peerMultiAddrStr)
}
