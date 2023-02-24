package network

import (
	"fmt"
	"net"
	"net/url"

	"github.com/multiformats/go-multiaddr"
)

const (
	// anyScheme is a string which fills the "scheme" degree of freedom for
	// a URL. Used to pad URLs without schemes such that they can be parsed
	// via stdlib URL parser.
	anyScheme = "scheme://"
)

var (
	// TECHDEBT: assuming TCP; remote peer's transport type must be knowable!
	// (ubiquitously switching to multiaddr, instead of a URL, would resolve this)
	peerTransportStr = "tcp"
	peerIPVersionStr = "ip4"
)

// Libp2pMultiaddrFromServiceUrl transforms a URL into its libp2p multiaddr equivalent.
// (see: https://github.com/libp2p/specs/blob/master/addressing/README.md#multiaddr-basics)
func Libp2pMultiaddrFromServiceUrl(serviceUrl string) (multiaddr.Multiaddr, error) {
	peerUrl, err := url.Parse(anyScheme + serviceUrl)
	if err != nil {
		return nil, fmt.Errorf(
			"parsing peer service URL: %s: %w",
			serviceUrl,
			err,
		)
	}

	/* CONSIDER: using a `/dns<4 or 6>/<hostname>` multiaddr instead of resolving here.
	 * I attempted using `/dns4/.../tcp/...` and go this error:
	 * > failed to listen on any addresses: [can only dial TCP over IPv4 or IPv6]
	 *
	 * TECHDEBT: is there a way for us to effectively prefer IPv6 responses when resolving DNS?
	 * TECHDEBT: resolving DNS this way has limitations:
	 * > The address parameter can use a host name, but this is not recommended,
	 * > because it will return at most one of the host name's IP addresses.
	 * (see: https://pkg.go.dev/net#ResolveIPAddr)
	 */
	peerIP, err := net.ResolveIPAddr(peerIPVersionStr, peerUrl.Hostname())
	if err != nil {
		return nil, fmt.Errorf(
			"resolving peer IP for hostname: %s: %w",
			peerUrl.Hostname(),
			err,
		)
	}

	// Test for IP version.
	// (see: https://pkg.go.dev/net#IP.To4)
	if peerIP.IP.To4() == nil {
		peerIPVersionStr = "ip6"
	}

	peerMultiAddrStr := fmt.Sprintf(
		"/%s/%s/%s/%s",
		peerIPVersionStr,
		peerIP,
		peerTransportStr,
		peerUrl.Port(),
	)
	return multiaddr.NewMultiaddr(peerMultiAddrStr)
}

// ServiceUrlFromLibp2pMultiaddr converts a multiaddr into a URL string.
func ServiceUrlFromLibp2pMultiaddr(addr multiaddr.Multiaddr) (string, error) {
	protos := addr.Protocols()
	if len(protos) < 2 {
		return "", fmt.Errorf(
			"unsupported multiaddr: %s; expected at least 2 protocols, got: %d",
			addr, len(protos),
		)
	}

	networkProto := protos[0]
	// e.g. IP address: "/ip4/10.0.0.1" --> "10.0.0.1"
	networkValue, err := addr.ValueForProtocol(networkProto.Code)
	if err != nil {
		return "", err
	}

	transportProto := protos[1]
	// e.g. Port: "/tcp/42069" --> "42069"
	transportValue, err := addr.ValueForProtocol(transportProto.Code)
	if err != nil {
		return "", err
	}

	// Top level proto must be a network protocol (e.g. ip4, ip6).
	switch networkProto.Code {
	case multiaddr.P_IP4, multiaddr.P_IP6:
		return fmt.Sprintf("%s:%s", networkValue, transportValue), nil
	}

	return "", fmt.Errorf(
		"unsupported network protocol, %s",
		networkProto.Name,
	)
}
