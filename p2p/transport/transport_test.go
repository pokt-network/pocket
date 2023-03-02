package transport

import (
	"testing"

	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/runtime/configs/types"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/stretchr/testify/require"
)

const (
	// localhostName represents an IPv4 address on the loopback interface
	localhostName = "127.0.0.1"
	randPort      = 42069
)

func TestTcpConn_ReadAll(t *testing.T) {
	expectedData := []byte("testing 123")
	receiver := newTestReceiver(t, localhostName, randPort)
	sender := newTestSender(t, receiver.address.String())

	// Send via `Write`
	numBzWritten, err := sender.Write(expectedData)
	require.NoError(t, err)
	require.Equal(t, len(expectedData), numBzWritten)

	// Receive via `ReadAll`
	actualData, err := receiver.ReadAll()
	require.NoError(t, err)
	require.Equal(t, expectedData, actualData)
}

func newTestReceiver(t *testing.T, hostname string, port int) *tcpConn {
	pk, err := crypto.GeneratePrivateKey()
	require.NoError(t, err)

	receiver, err := createTCPListener(&configs.P2PConfig{
		PrivateKey:     pk.String(),
		Hostname:       hostname,
		Port:           uint32(port),
		UseRainTree:    false,
		ConnectionType: types.ConnectionType_TCPConnection,
	})
	require.NoError(t, err)
	return receiver
}

func newTestSender(t *testing.T, receiverURL string) *tcpConn {
	pk, err := crypto.GeneratePrivateKey()
	require.NoError(t, err)

	sender, err := createTCPDialer(&configs.P2PConfig{
		PrivateKey:     pk.String(),
		UseRainTree:    false,
		IsClientOnly:   true,
		ConnectionType: types.ConnectionType_TCPConnection,
	}, receiverURL)
	require.NoError(t, err)
	return sender
}
