package testutil

import (
	"fmt"
	"net"
	"net/url"

	"github.com/foxcpp/go-mockdns"
)

func PrepareDNSMockFromServiceURLs(serviceURLs []string) (done func(), err error) {
	zones := make(map[string]mockdns.Zone)
	for i, u := range serviceURLs {
		// Perpend `scheme://` as serviceURLs are currently scheme-less.
		// Required for parsing to produce useful results.
		// (see: https://pkg.go.dev/net/url@go1.20.2#URL)
		serviceURL, err := url.Parse(fmt.Sprintf("scheme://%s", u))
		if err != nil {
			return nil, err
		}

		ipStr := fmt.Sprintf("10.0.0.%d", i+1)

		if i >= 254 {
			panic(fmt.Sprintf("would generate invalid IPv4 address: %s", ipStr))
		}

		zones[fmt.Sprintf("%s.", serviceURL.Hostname())] = mockdns.Zone{
			A: []string{ipStr},
		}
	}

	return PrepareDNSMock(zones), nil
}

func PrepareDNSMock(zones map[string]mockdns.Zone) (done func()) {
	srv, _ := mockdns.NewServerWithLogger(zones, noopLogger{}, false)
	srv.PatchNet(net.DefaultResolver)
	return func() {
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
