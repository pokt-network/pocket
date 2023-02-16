// CONSIDERATION: Consider moving this into `shared` if the libp2p identity
// ends up consolidating with the node's identity.
package identity

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	libp2pCrypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/network"
	libp2pPeer "github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"

	"github.com/pokt-network/pocket/p2p/transport"
	"github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/shared/crypto"
)

var (
	ErrIdentity = types.NewErrFactory("identity error")
)

// PeerFromLibp2pStream builds a network peer using peer info available
// from the given libp2p stream. **The returned `ServiceUrl` is a libp2p
// multiaddr string as opposed to a proper URL.**
func PeerFromLibp2pStream(stream network.Stream) (*types.NetworkPeer, error) {
	publicKeyBz, err := stream.Conn().RemotePublicKey().Raw()
	// NB: abort handling this stream.
	if err != nil {
		return nil, err
	}
	publicKey, err := crypto.NewPublicKeyFromBytes(publicKeyBz)
	if err != nil {
		return nil, err
	}

	// TECHDEBT: currently returning libp2p multiaddr version of ServiceUrl.
	return &types.NetworkPeer{
		Dialer:    transport.NewLibP2PTransport(stream),
		PublicKey: publicKey,
		// NB: pokt analogue of libp2p peer.ID
		Address:    publicKey.Address(),
		ServiceUrl: stream.Conn().RemoteMultiaddr().String(),
	}, nil
}

// Libp2pPublicKeyFromPeer retrieves the libp2p compatible public key from a pocket peer.
func Libp2pPublicKeyFromPeer(peer *types.NetworkPeer) (libp2pCrypto.PubKey, error) {
	publicKey, err := libp2pCrypto.UnmarshalEd25519PublicKey(peer.PublicKey.Bytes())
	if err != nil {
		return nil, ErrIdentity(fmt.Sprintf(
			"unmarshalling peer ed25519 public key, pokt address: %s", peer.Address,
		), err)
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
		return libp2pPeer.AddrInfo{}, ErrIdentity(fmt.Sprintf(
			"retrieving ID from peer public key, pokt address: %s", peer.Address,
		), err)
	}

	peerMultiaddr, err := multiaddr.NewMultiaddr(peer.ServiceUrl)
	// NB: early return if we already have a multiaddr.
	if err == nil {
		return libp2pPeer.AddrInfo{
			ID: peerID,
			Addrs: []multiaddr.Multiaddr{
				peerMultiaddr,
			},
		}, nil
	}

	peerMultiaddr, err = Libp2pMultiaddrFromServiceUrl(peer.ServiceUrl)
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

// Libp2pMultiaddrFromServiceUrl transforms a URL into its libp2p muleiaddr equivalent.
// (see: https://github.com/libp2p/specs/blob/master/addressing/README.md#multiaddr-basics)
// TECHDEBT: this probably belongs somewhere else, it's more of a networking helper.
func Libp2pMultiaddrFromServiceUrl(serviceUrl string) (multiaddr.Multiaddr, error) {
	// NB: hard-code a scheme for URL parsing to work.
	peerUrl, err := url.Parse("scheme://" + serviceUrl)
	if err != nil {
		return nil, ErrIdentity(fmt.Sprintf(
			"parsing peer service URL: %s", serviceUrl,
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
