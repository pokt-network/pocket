package identity

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPeerMultiAddrFromServiceURL_success(t *testing.T) {
	testCases := []struct {
		name                string
		serviceURL          string
		expetedMultiaddrStr string
	}{
		{
			"FQDN",
			"node1.consensus:8080",
			"/dns/node1.consensus/tcp/8080",
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
			actualMultiaddr, err := PeerMultiAddrFromServiceURL(testCase.serviceURL)
			require.NoError(t, err)
			require.NotNil(t, actualMultiaddr)
			require.Equal(t, testCase.expetedMultiaddrStr, actualMultiaddr.String())
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
		FQDN: "node1.consensus",
		IP4:  "142.250.181.196",
		IP6:  "2a00:1450:4005:802::2004",
	}

	testCases := []struct {
		name             string
		serviceURLFormat string
		// TODO: assert specific errors?
		// expectedError string
	}{
		// NB: usage of scheme is invalid.
		{
			"FQDN with scheme",
			"tcp://%s:8080",
		},

		// NB: port **number** is required
		{
			"invalid port number",
			"%s:abc",
		},
		{
			"missing port number",
			"%s:",
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
				serviceURL := fmt.Sprintf(testCase.serviceURLFormat, hostname)
				actualMultiaddr, err := PeerMultiAddrFromServiceURL(serviceURL)
				// TODO: assert specific errors?
				if !assert.NoError(t, err) {
					t.Fatalf("actualMultiaddr: %s", actualMultiaddr)
				}
			})
		}
	}
}
