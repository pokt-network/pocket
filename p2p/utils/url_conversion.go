package utils

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"net"
	"net/url"
	"strconv"
	"strings"

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
	anyScheme                     = "scheme://"
	errResolvePeerIPMsg           = "resolving peer IP for hostname"
	errMissingPortAndDelimiterMsg = "missing port number and delimiter in service URL"
	errParsingServiceURLMsg       = "parsing peer service URL"
	errInvalidPortMsg             = "invalid port"
	errInvalidSchemeUsageMsg      = "usage of scheme is invalid in service URL"
)

// Libp2pMultiaddrFromServiceURL transforms a URL into its libp2p multiaddr equivalent. The URL must contain a port number.
// The URL may contain a hostname or IP address. If a hostname is provided, it will be resolved to an IP address.
// The URL may not contain a scheme. The URL may contain an IPv6 address, but it must be enclosed in square brackets.
// The URL may contain an IPv4 address, but it must not be enclosed in square brackets.
// (see: https://www.rfc-editor.org/rfc/rfc3986#section-3.2.2)
// (see: https://github.com/libp2p/specs/blob/master/addressing/README.md#multiaddr-basics)
func Libp2pMultiaddrFromServiceURL(serviceURL string) (multiaddr.Multiaddr, error) {
	var (
		// TECHDEBT: assuming TCP; remote peer's transport type must be knowable!
		// (ubiquitously switching to multiaddr, instead of a URL, would resolve this)
		peerTransportStr = transportTypeTCP
		// Default to IPv4; updated via subsequent checks.
		peerIPVersionStr = ipVersion4
	)

	// Conditionally add the anyScheme prefix if no scheme is present.
	if strings.Contains(serviceURL, "://") {
		return nil, fmt.Errorf("%s: %s", errInvalidSchemeUsageMsg, serviceURL)
	} else {
		serviceURL = anyScheme + serviceURL
	}

	peerUrl, err := url.Parse(serviceURL)
	if err != nil {
		return nil, fmt.Errorf("%s: %s: %w", errParsingServiceURLMsg, serviceURL, err)
	}

	// Check if service URL contains a port number
	if _, port, err := net.SplitHostPort(peerUrl.Host); err != nil || port == "" {
		return nil, fmt.Errorf("%s: %s: %w", errMissingPortAndDelimiterMsg, serviceURL, err)
	}

	peerIP, err := getPeerIP(peerUrl.Hostname())
	if err != nil {
		return nil, fmt.Errorf("%s: %s: %w", errResolvePeerIPMsg, peerUrl.Hostname(), err)
	}

	// Check IP version.
	if peerIP.To4() == nil {
		peerIPVersionStr = ipVersion6
	}

	// Check if port is valid.
	if _, err := strconv.Atoi(peerUrl.Port()); err != nil {
		return nil, fmt.Errorf("%s: %s", errInvalidPortMsg, peerUrl.Port())
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

// ServiceURLFromLibp2pMultiaddr converts a multiaddr into a URL string.
func ServiceURLFromLibp2pMultiaddr(addr multiaddr.Multiaddr) (string, error) {
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
	case multiaddr.P_IP4:
		return fmt.Sprintf("%s:%s", networkValue, transportValue), nil
	case multiaddr.P_IP6:
		return fmt.Sprintf("[%s]:%s", networkValue, transportValue), nil
	}

	return "", fmt.Errorf(
		"unsupported network protocol, %s",
		networkProtocol.Name,
	)
}

func newResolvePeerIPErr(hostname string, err error) error {
	return fmt.Errorf("%s: %s, %w", errResolvePeerIPMsg, hostname, err)
}

func getPeerIP(hostname string) (net.IP, error) {
	// Attempt to parse peer hostname as an IP address.
	// (see: https://pkg.go.dev/net#ParseIP)
	peerIP := net.ParseIP(hostname)
	if peerIP != nil {
		return peerIP, nil
	}

	// CONSIDERATION: using a `/dns<4 or 6>/<hostname>` multiaddr instead of resolving here.
	// I attempted using `/dns4/.../tcp/...` and go this error:
	// > failed to listen on any addresses: [can only dial TCP over IPv4 or IPv6]
	// TECHDEBT(#595): receive `ctx` from caller.
	addrs, err := net.DefaultResolver.LookupHost(context.TODO(), hostname)
	if err != nil {
		return nil, newResolvePeerIPErr(hostname, err)
	}

	// CONSIDERATION: preferring IPv6 responses when resolving DNS.
	// Return first address which is a parsable IP address.
	var (
		validIPs    []net.IP
		randomIndex int
	)
	for _, addr := range addrs {
		ip := net.ParseIP(addr)
		if ip == nil {
			continue
		}
		validIPs = append(validIPs, ip)
	}

	switch len(validIPs) {
	case 0:
		return nil, newResolvePeerIPErr(hostname, err)
	case 1:
		return validIPs[0], nil
	}

	bigRandIndex, randErr := rand.Int(rand.Reader, big.NewInt(int64(len(validIPs))))
	if randErr == nil {
		randomIndex = int(bigRandIndex.Int64())
	}

	// Select a pseudorandom, valid IP address.
	// Because `randomIndex` defaults to 0, selection will fall back to first
	// valid IP if `randErr != nil`.
	peerIP = validIPs[randomIndex]

	// TECHDEBT(#557): remove this log line once direction is clearer
	// on supporting multiple network addresses per peer.
	logger.Global.Warn().Msg("resolved multiple addresses but only using one. See ticket #557 for more details")
	logger.Global.Warn().
		Str("hostname", hostname).
		Array("resolved", stringLogArrayMarshaler{strs: addrs}).
		IPAddr("using", peerIP)

	return peerIP, nil
}

// stringLogArrayMarshaler implements the `zerolog.LogArrayMarshaler` interface
// to marshal an array of strings for use with zerolog.
// TECHDEBT(#609): move & de-duplicate
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
