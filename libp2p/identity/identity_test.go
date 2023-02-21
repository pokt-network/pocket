package identity

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPeerMultiAddrFromServiceURL_success(t *testing.T) {
	testCases := []struct {
		name                  string
		serviceUrl            string
		expetedMultiaddrRegex string
	}{
		{
			"FQDN",
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
	FQDN = "fqdn"
	IP4  = "ip4"
	IP6  = "ip6"
)

func TestPeerMultiAddrFromServiceURL_error(t *testing.T) {
	hostnames := map[string]string{
		FQDN: "www.google.com",
		IP4:  "142.250.181.196",
		IP6:  "2a00:1450:4005:802::2004",
	}

	testCases := []struct {
		name             string
		serviceUrlFormat string
		// TECHDEBT: assert specific errors.
		expectedErrorStr string
	}{
		// NB: usage of scheme is invalid.
		{
			"FQDN with scheme",
			"tcp://%s:8080",
			"no such host",
		},

		// NB: port **number** is required
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
		//{
		//	"missing port number and delimiter",
		//	"%s",
		//},
	}

	for _, testCase := range testCases {
		for hostType, hostname := range hostnames {
			testName := fmt.Sprintf("%s/%s", testCase.name, hostType)
			t.Run(testName, func(t *testing.T) {
				serviceURL := fmt.Sprintf(testCase.serviceUrlFormat, hostname)
				actualMultiaddr, err := Libp2pMultiaddrFromServiceUrl(serviceURL)
				// TECHDEBT: assert specific errors
				// Print resulting multiaddr to understand why no error.
				require.ErrorContains(t, err, testCase.expectedErrorStr, fmt.Sprintf("actualMultiaddr: %s", actualMultiaddr))
			})
		}
	}
}
