package network

import (
	"fmt"
	"net"
	"net/url"

	"github.com/multiformats/go-multiaddr"
	"github.com/pokt-network/pocket/logger"
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
		logger.Global.Error().Err(err).Str("serviceUrl", serviceUrl).Msg("tried to parse peer service URL")

		return nil, fmt.Errorf(
			"parsing peer service URL: %s: %w",
			serviceUrl,
			err,
		)
	}

	logger.Global.Info().Str("peer_url", peerUrl.String()).Msg("parsed peer url")

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
		logger.Global.Error().Str("peer_url", peerUrl.String()).Str("peerIPVersionStr", peerIPVersionStr).Str("hostname", peerUrl.Hostname()).Err(err).Msg("tried to resolve IP addr")
		return nil, fmt.Errorf(
			"resolving peer IP for hostname: %s: %w",
			peerUrl.Hostname(),
			err,
		)
	}

	logger.Global.Info().Str("peer_ip", peerIP.String()).Msg("resolved IP address")

	// Test for IP version.
	// (see: https://pkg.go.dev/net#IP.To4)
	if peerIP.IP.To4() == nil {
		peerIPVersionStr = "ip6"
	}

	logger.Global.Info().Str("peerIPVersionStr", peerIPVersionStr).Msg("checked IP version")

	peerMultiAddrStr := fmt.Sprintf(
		"/%s/%s/%s/%s",
		peerIPVersionStr,
		peerIP,
		peerTransportStr,
		peerUrl.Port(),
	)

	logger.Global.Info().Str("peerMultiAddrStr", peerMultiAddrStr).Msg("composed peerMultiAddrStr")

	ma, err := multiaddr.NewMultiaddr(peerMultiAddrStr)
	if err != nil {
		logger.Global.Error().Err(err).Msg("tried to create NewMultiaddr")
		return nil, err
	}

	logger.Global.Info().Str("Multiaddr", ma.String()).Msg("created new Multiaddr")
	return ma, err
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
