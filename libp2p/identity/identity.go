// CONSIDERATION: Consider moving this into `shared` if the libp2p identity
// ends up consolidating with the node's identity.
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
	ErrIdentity = types.NewErrFactory("identity error")
)

// PoktPeerFromStream builds a network peer using peer info available
// from the given libp2p stream. **The returned `ServiceUrl` is a libp2p
// multiaddr string as opposed to a proper URL.**
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

	// TECHDEBT: currently returning libp2p multiaddr version of ServiceUrl.
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
		return nil, ErrIdentity(fmt.Sprintf(
			"unmarshalling peer ed25519 public key, pokt address: %s", poktPeer.Address,
		), err)
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
		return peer.AddrInfo{}, ErrIdentity(fmt.Sprintf(
			"retrieving ID from peer public key, pokt address: %s", poktPeer.Address,
		), err)
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

// Libp2pMultiaddrFromServiceUrl transforms a URL into its libp2p muleiaddr equivalent.
// (see: https://github.com/libp2p/specs/blob/master/addressing/README.md#multiaddr-basics)
// TECHDEBT: this probably belongs somewhere else, it's more of a networking helper.
func PeerMultiAddrFromServiceURL(serviceURL string) (multiaddr.Multiaddr, error) {
	// NB: hard-code a scheme for URL parsing to work.
	peerUrl, err := url.Parse("scheme://" + serviceURL)
	if err != nil {
		return nil, ErrIdentity(fmt.Sprintf(
			"parsing peer service URL: %s", serviceURL,
		), err)
	}

	var (
		// TECHDEBT: assuming TCP; remote peer's transport type must be knowable!
		// (ubiquitously switching to multiaddr, instead of a URL, would resolve this)
		peerTransportStr = "tcp"
		peerHostnameStr  = peerUrl.Hostname()
		// TECHDEBT: is there a way for us to effectively prefer IPv6 responses?
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
