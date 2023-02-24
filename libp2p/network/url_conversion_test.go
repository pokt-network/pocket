package network

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/require"
)

func TestPeerMultiAddrFromServiceURL_Success(t *testing.T) {
	testCases := []struct {
		name                  string
		serviceUrl            string
		expetedMultiaddrRegex string
	}{
		{
			"fqdn",
			"www.google.com:8080",
			`/ip4/(\d{1,3}\.){3}\d{1,3}/tcp/8080`,
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
			actualMultiaddr, err := Libp2pMultiaddrFromServiceUrl(testCase.serviceUrl)
			require.NoError(t, err)
			require.NotNil(t, actualMultiaddr)
			require.Regexp(t, regexp.MustCompile(testCase.expetedMultiaddrRegex), actualMultiaddr.String())
		})
	}
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
		serviceUrlFormat string
		// TECHDEBT: assert specific errors.
		expectedErrContains string
	}{
		// Usage of scheme is invalid.
		{
			"fully qualified domain name with scheme",
			"%s:8080",
			"no such host",
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
		// can omit hextet delimiters and still be valid.
		// (see: Address_representation)
		// {
		// 	"missing port number and delimiter",
		// 	"%s",
		// },
	}

	for _, testCase := range testCases {
		for hostType, hostname := range hostnames {
			testName := fmt.Sprintf("%s/%s", testCase.name, hostType)
			t.Run(testName, func(t *testing.T) {
				serviceURL := fmt.Sprintf(testCase.serviceUrlFormat, hostname)
				actualMultiaddr, err := Libp2pMultiaddrFromServiceUrl(serviceURL)
				// TECHDEBT: assert specific errors
				// Print resulting multiaddr to understand why no error.
				require.ErrorContainsf(t, err, testCase.expectedErrContains, fmt.Sprintf("actualMultiaddr: %s", actualMultiaddr))
			})
		}
	}
}

func TestServiceUrlFromLibp2pMultiaddr_Success(t *testing.T) {
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
			"IPv6",
			"/ip6/2a00:1450:4005:802::2004/tcp/8080",
			"2a00:1450:4005:802::2004:8080",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			addr, err := multiaddr.NewMultiaddr(testCase.multiaddrStr)
			require.NoError(t, err)

			actualUrl, err := ServiceUrlFromLibp2pMultiaddr(addr)
			require.NoError(t, err)
			require.NotNil(t, actualUrl)
			require.Equal(t, testCase.expectedUrl, actualUrl)
		})
	}
}

func TestServiceUrlFromLibp2pMultiaddr_Error(t *testing.T) {
	testCases := []struct {
		name                string
		multiaddrStr        string
		expectedErrContains string
	}{
		{
			"fqdn",
			"/dns/www.google.com/tcp/8080",
			"unsupported network protocol",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			addr, err := multiaddr.NewMultiaddr(testCase.multiaddrStr)
			require.NoError(t, err)

			_, err = ServiceUrlFromLibp2pMultiaddr(addr)
			// DISCUSS: asserting specific errors instead.
			require.ErrorContains(t, err, testCase.expectedErrContains)
		})
	}
}
