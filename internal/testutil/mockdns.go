package testutil

import (
	"fmt"
	"net"

	"github.com/foxcpp/go-mockdns"
	"github.com/regen-network/gocuke"
	"github.com/stretchr/testify/require"
)

func DNSMockFromServiceURLs(t gocuke.TestingT, serviceURLs []string) (srv *mockdns.Server, done func()) {
	t.Helper()

	srv, done = MinimalDNSMock(t)
	for _, serviceURL := range serviceURLs {
		AddServiceURLZone(t, srv, serviceURL)
	}
	return srv, done
}

func AddServiceURLZone(t gocuke.TestingT, srv *mockdns.Server, serviceURL string) {
	t.Helper()

	// TODO_THIS_COMMIT: move & de-dup
	hostname, _, err := net.SplitHostPort(serviceURL)
	require.NoError(t, err)

	zone := mockdns.Zone{
		A: []string{"10.0.0.1"},
	}

	err = srv.AddZone(fmt.Sprintf("%s.", hostname), zone)
	require.NoError(t, err)
}

func MinimalDNSMock(t gocuke.TestingT) (srv *mockdns.Server, done func()) {
	t.Helper()

	return BaseDNSMock(t, nil)
}

func BaseDNSMock(t gocuke.TestingT, zones map[string]mockdns.Zone) (srv *mockdns.Server, done func()) {
	t.Helper()

	if zones == nil {
		zones = make(map[string]mockdns.Zone)
	}

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
