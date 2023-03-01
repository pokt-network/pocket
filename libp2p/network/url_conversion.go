package network

import (
	"fmt"
	"math/rand"
	"net"
	"net/url"

	"github.com/multiformats/go-multiaddr"
	"github.com/pokt-network/pocket/logger"
	"github.com/rs/zerolog"
)

const (
	transportTypeTCP = "tcp"
	ipVersion4       = "ip4"
	ipVersion6       = "ip6"

	// anyScheme is a string which fills the "scheme" degree of freedom for
	// a URL. Used to pad URLs without schemes such that they can be parsed
	// via stdlib URL parser.
	anyScheme           = "scheme://"
	errResolvePeerIPMsg = "resolving peer IP for hostname: %s: %w"
)

// Libp2pMultiaddrFromServiceUrl transforms a URL into its libp2p multiaddr equivalent.
// (see: https://github.com/libp2p/specs/blob/master/addressing/README.md#multiaddr-basics)
func Libp2pMultiaddrFromServiceUrl(serviceUrl string) (multiaddr.Multiaddr, error) {
	var (
		// TECHDEBT: assuming TCP; remote peer's transport type must be knowable!
		// (ubiquitously switching to multiaddr, instead of a URL, would resolve this)
		peerTransportStr = transportTypeTCP
		// Default to IPv4; updated via subsequent checks.
		peerIPVersionStr = ipVersion4
	)

	peerUrl, err := url.Parse(anyScheme + serviceUrl)
	if err != nil {
		return nil, fmt.Errorf(
			"parsing peer service URL: %s: %w",
			serviceUrl,
			err,
		)
	}

	peerIP, err := getPeerIP(peerUrl.Hostname())
	if err != nil {
		return nil, err
	}

	// Check IP version.
	if peerIP.To4() == nil {
		peerIPVersionStr = ipVersion6
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
	protocols := addr.Protocols()
	if len(protocols) < 2 {
		return "", fmt.Errorf(
			"unsupported multiaddr: %s; expected at least 2 protocols, got: %d",
			addr, len(protocols),
		)
	}

	networkProtocol := protocols[0]
	// e.g. IP address: "/ip4/10.0.0.1" --> "10.0.0.1"
	networkValue, err := addr.ValueForProtocol(networkProtocol.Code)
	if err != nil {
		return "", err
	}

	transportProtocol := protocols[1]
	// e.g. Port: "/tcp/42069" --> "42069"
	transportValue, err := addr.ValueForProtocol(transportProtocol.Code)
	if err != nil {
		return "", err
	}

	// Top level protocol must be a network protocol (e.g. ip4, ip6).
	switch networkProtocol.Code {
	case multiaddr.P_IP4, multiaddr.P_IP6:
		return fmt.Sprintf("%s:%s", networkValue, transportValue), nil
	}

	return "", fmt.Errorf(
		"unsupported network protocol, %s",
		networkProtocol.Name,
	)
}

func newResolvePeerIPErr(hostname string, err error) error {
	return fmt.Errorf(errResolvePeerIPMsg, hostname, err)
}

func getPeerIP(hostname string) (net.IP, error) {
	// Attempt to parse peer hostname as an IP address.
	// (see: https://pkg.go.dev/net#ParseIP)
	peerIP := net.ParseIP(hostname)
	if peerIP != nil {
		return peerIP, nil
	}

	// CONSIDER: using a `/dns<4 or 6>/<hostname>` multiaddr instead of resolving here.
	// I attempted using `/dns4/.../tcp/...` and go this error:
	// > failed to listen on any addresses: [can only dial TCP over IPv4 or IPv6]
	addrs, err := net.LookupHost(hostname)
	if err != nil {
		return nil, newResolvePeerIPErr(hostname, err)
	}

	// CONSIDER: which address(es) should we use when multiple
	// are provided in a DNS response?
	// CONSIDER: preferring IPv6 responses when resolving DNS.
	// Return first address which is a parsable IP address.
	for _, addr := range addrs {
		peerIP = net.ParseIP(addr)
		if peerIP == nil {
			continue
		}
		// TECHDEBT: remove this log line once direction is clearer
		// on supporting multiple network addresses per peer.
		if len(addrs) > 1 {
			logger.Global.Warn().
				Array("resolved", newStringLogArrayMarshaler(addrs)).
				IPAddr("using", peerIP)
		}
		return peerIP, nil
	}
	return nil, newResolvePeerIPErr(hostname, err)
}

// stringLogArrayMarshaler implements the `zerolog.LogArrayMarshaler` interface
// to marshal an array of strings for use with zerolog.
type stringLogArrayMarshaler struct {
	strs []string
}

// MarshalZerologArray implements the respective `zerolog.LogArrayMarshaler`
// interface member.
func (marshaler stringLogArrayMarshaler) MarshalZerologArray(arr *zerolog.Array) {
	for _, str := range marshaler.strs {
		arr.Str(str)
	}
}

func newStringLogArrayMarshaler(strs []string) zerolog.LogArrayMarshaler {
	return stringLogArrayMarshaler{
		strs: strs,
	}
}
