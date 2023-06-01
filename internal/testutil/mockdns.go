package testutil

import (
	"fmt"
	"github.com/foxcpp/go-mockdns"
	"github.com/regen-network/gocuke"
	"github.com/stretchr/testify/require"
	"net"
	"net/url"
)

func PrepareDNSMockFromServiceURLs(t gocuke.TestingT, serviceURLs []string) (srv *mockdns.Server, done func()) {
	t.Helper()

	zones := make(map[string]mockdns.Zone)
	for i, u := range serviceURLs {
		// Perpend `scheme://` as serviceURLs are currently scheme-less.
		// Required for parsing to produce useful results.
		// (see: https://pkg.go.dev/net/url@go1.20.2#URL)
		serviceURL, err := url.Parse(fmt.Sprintf("scheme://%s", u))
		require.NoError(t, err)

		ipStr := fmt.Sprintf("10.0.0.%d", i+1)

		if i >= 254 {
			panic(fmt.Sprintf("would generate invalid IPv4 address: %s", ipStr))
		}

		zones[fmt.Sprintf("%s.", serviceURL.Hostname())] = mockdns.Zone{
			A: []string{ipStr},
		}
	}

	return PrepareDNSMock(t, zones)
}

func PrepareDNSMock(t gocuke.TestingT, zones map[string]mockdns.Zone) (srv *mockdns.Server, done func()) {
	t.Helper()

	srv, _ = mockdns.NewServerWithLogger(zones, noopLogger{}, false)
	srv.PatchNet(net.DefaultResolver)
	return srv, func() {
		_ = srv.Close()
		mockdns.UnpatchNet(net.DefaultResolver)
	}
}

// NB: default logging behavior is too noisy.
// noopLogger implements go-mockdns's `mockdns.Logger` interface.
type noopLogger struct{}

func (nl noopLogger) Printf(format string, args ...interface{}) {
	// noop
}
