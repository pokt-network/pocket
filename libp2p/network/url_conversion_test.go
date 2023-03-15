package network

import (
	"fmt"
	"net"
	"testing"

	"github.com/foxcpp/go-mockdns"

	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/require"
)

func TestPeerMultiAddrFromServiceURL_Success(t *testing.T) {
	zones := map[string]mockdns.Zone{
		"www.google.com.": {
			A: []string{"142.250.181.196"},
		},
	}
	closeDNSMock := prepareDNSResolverMock(t, zones)

	testCases := []struct {
		name              string
		serviceURL        string
		expectedMultiaddr string
	}{
		{
			"fqdn",
			"www.google.com:8080",
			"/ip4/142.250.181.196/tcp/8080",
		},
		{
			"IPv4",
			"142.250.181.196:8080",
			"/ip4/142.250.181.196/tcp/8080",
		},
		{
			"IPv6",
			"2a00:1450:4005:802::2004:8080",
			"/ip6/2a00:1450:4005:802::2004/tcp/8080",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actualMultiaddr, err := Libp2pMultiaddrFromServiceURL(testCase.serviceURL)
			require.NoError(t, err)
			require.NotNil(t, actualMultiaddr)
			require.Equal(t, testCase.expectedMultiaddr, actualMultiaddr.String())
		})
	}
	closeDNSMock()
}

const (
	// (see: https://en.wikipedia.org/wiki/Fully_qualified_domain_name)
	fqdn = "fqdn"
	ip4  = "ip4"
	ip6  = "ip6"
)

func TestPeerMultiAddrFromServiceURL_Error(t *testing.T) {
	hostnames := map[string]string{
		fqdn: "www.google.com",
		ip4:  "142.250.181.196",
		ip6:  "2a00:1450:4005:802::2004",
	}

	testCases := []struct {
		name             string
		serviceURLFormat string
		// TECHDEBT: assert specific errors.
		expectedErrContains string
	}{
		// Usage of scheme is invalid.
		{
			"fully qualified domain name with scheme",
			"tcp://%s:8080",
			"resolving peer IP for hostname",
		},

		// Port **number** is required
		{
			"invalid port number",
			"%s:abc",
			"invalid port",
		},
		{
			"missing port number",
			"%s:",
			"unexpected end of multiaddr",
		},
		// TODO: this case is tricky to detect as IPv6 addresses
		// can omit multiple "hextet" delimiters and still be valid.
		// (see: https://en.wikipedia.org/wiki/IPv6#Address_representation)
		// {
		// 	"missing port number and delimiter",
		// 	"%s",
		// },
	}

	for _, testCase := range testCases {
		for hostType, hostname := range hostnames {
			testName := fmt.Sprintf("%s/%s", testCase.name, hostType)
			t.Run(testName, func(t *testing.T) {
				serviceURL := fmt.Sprintf(testCase.serviceURLFormat, hostname)
				actualMultiaddr, err := Libp2pMultiaddrFromServiceURL(serviceURL)
				// TECHDEBT: assert specific errors
				// Print resulting multiaddr to understand why no error.
				require.ErrorContainsf(t, err, testCase.expectedErrContains, fmt.Sprintf("actualMultiaddr: %s", actualMultiaddr))
			})
		}
	}
}

func TestServiceURLFromLibp2pMultiaddr_Success(t *testing.T) {
	testCases := []struct {
		name         string
		multiaddrStr string
		expectedUrl  string
	}{
		{
			"IPv4",
			"/ip4/142.250.181.196/tcp/8080",
			"142.250.181.196:8080",
		},
		{
			"IPv6 full",
			"/ip6/2a00:1450:4005:0802:0000:0000:0000:2004/tcp/8080",
			"2a00:1450:4005:802::2004:8080",
		},
		{
			"IPv6 short",
			"/ip6/2a00:1450:4005:802::2004/tcp/8080",
			"2a00:1450:4005:802::2004:8080",
		},
		{
			"IPv6 shorter",
			// NB: this address is not equivalent to those above.
			"/ip6/2a00::2004/tcp/8080",
			"2a00::2004:8080",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			addr, err := multiaddr.NewMultiaddr(testCase.multiaddrStr)
			require.NoError(t, err)

			actualUrl, err := ServiceURLFromLibp2pMultiaddr(addr)
			require.NoError(t, err)
			require.NotNil(t, actualUrl)
			require.Equal(t, testCase.expectedUrl, actualUrl)
		})
	}
}

func TestServiceURLFromLibp2pMultiaddr_Error(t *testing.T) {
	testCases := []struct {
		name                string
		multiaddrStr        string
		expectedErrContains string
	}{
		{
			"`dns` network protocol",
			"/dns/www.google.com/tcp/8080",
			"unsupported network protocol",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			addr, err := multiaddr.NewMultiaddr(testCase.multiaddrStr)
			require.NoError(t, err)

			_, err = ServiceURLFromLibp2pMultiaddr(addr)
			// DISCUSS: improved error assertion methodology
			require.ErrorContains(t, err, testCase.expectedErrContains)
		})
	}
}

// TECHDEBT: add helpers for crating and/or using a "test resolver" which can
// be switched with `net.DefaultResolver` and can return mocked responses.
func TestGetPeerIP_Success(t *testing.T) {
	t.Skip("TODO: replace `net.DefaultResolver` with one which has a `Dial` function that returns a mocked `net.Conn` (see: https://pkg.go.dev/net#Resolver)")

	//nolint:gocritic // commentedOutCode - Outlines the minimum requirements for disproving regression.
	// testCases := []struct {
	// 	name       string
	// 	hostname   string
	// 	// TECHDEBT: seed math/rand for predictable selection within mocked response.
	// 	expectedIP net.IP
	// }{
	// 	{
	// 		"single A record",
	// 		"single.A.example",
	// 	},
	// 	{
	// 		"single AAAA record",
	// 		"single.AAAA.example",
	// 	},
	// 	{
	// 		"multiple A records",
	// 		"multi.A.example",
	// 	},
	// 	{
	// 		"multiple AAAA records",
	// 		"multi.AAAA.example",
	// 	},
	// }
}

func TestGetPeerIP_Error(t *testing.T) {
	// `example` top-level domains should not resolve by default and therefore
	// should reliably fail to resolve under normal, real-world conditions.
	// (see: https://en.wikipedia.org/wiki/.example)

	hostname := "nonexistent.example"
	_, err := getPeerIP(hostname)
	require.ErrorContains(t, err, errResolvePeerIPMsg)
	require.ErrorContains(t, err, hostname)
}

func prepareDNSResolverMock(t *testing.T, zones map[string]mockdns.Zone) (done func()) {
	srv, err := mockdns.NewServerWithLogger(zones, noopLogger{}, false)
	require.NoError(t, err)

	srv.PatchNet(net.DefaultResolver)
	return func() {
		_ = srv.Close()
		mockdns.UnpatchNet(net.DefaultResolver)
	}
}

// noopLogger implements go-mockdns's `mockdns.Logger` interface.
// The default logging behavior in mockdns is too noisy.
type noopLogger struct{}

func (nl noopLogger) Printf(format string, args ...interface{}) {
	// noop
}
