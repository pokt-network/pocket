package p2p

import (
	"fmt"
	"testing"

	shared "github.com/pokt-network/pocket/shared/config"
	"github.com/stretchr/testify/assert"
)

func TestNetworkLifecycle_Configure(t *testing.T) {
	m := newP2PModule()

	{ // ProtocolField
		config := &shared.P2PConfig{}

		assert.EqualError(
			t,
			m.configure(config),
			ErrMissingOrEmptyConfigField("Protocol").Error(),
			"Configure: failed, expected configure to throw an error on missing fields, but it did not.",
		)
	}
	{ // AddressField
		config := &shared.P2PConfig{
			Protocol: "tcp",
		}

		assert.EqualError(
			t,
			m.configure(config),
			ErrMissingOrEmptyConfigField("Address").Error(),
			"Configure: failed, expected configure to throw an error on missing fields, but it did not.",
		)
	}
	{ // ExternalIpField
		config := &shared.P2PConfig{
			Protocol: "tcp",
			Address:  make([]byte, 0),
		}

		assert.EqualError(
			t,
			m.configure(config),
			ErrMissingOrEmptyConfigField("ExternalIp").Error(),
			"Configure: failed, expected configure to throw an error on missing fields, but it did not.",
		)
	}

	{ // PeersField
		config := &shared.P2PConfig{
			Protocol:   "tcp",
			Address:    make([]byte, 10),
			ExternalIp: "1.1.1.2:3333",
		}

		assert.EqualError(
			t,
			m.configure(config),
			ErrMissingOrEmptyConfigField("Peers").Error(),
			"Configure: failed, expected configure to throw an error on missing fields, but it did not.",
		)
	}

	{ // MaxInboundField
		config := &shared.P2PConfig{
			Protocol:   "tcp",
			Address:    make([]byte, 10),
			ExternalIp: "1.1.1.2:3333",
			Peers:      []string{"1:2222", "2:33333", "3:44444"},
		}

		assert.EqualError(
			t,
			m.configure(config),
			ErrMissingOrEmptyConfigField("MaxInbound").Error(),
			"Configure: failed, expected configure to throw an error on missing fields, but it did not.",
		)
	}

	{ // MaxOutbound
		config := &shared.P2PConfig{
			Protocol:   "tcp",
			Address:    make([]byte, 10),
			ExternalIp: "101010",
			Peers:      []string{"1", "2", "d"},
			MaxInbound: 10,
		}

		assert.EqualError(
			t,
			m.configure(config),
			ErrMissingOrEmptyConfigField("MaxOutbound").Error(),
			"Configure: failed, expected configure to throw an error on missing fields, but it did not.",
		)
	}

	{ // WireHeaderLength
		config := &shared.P2PConfig{
			Protocol:    "tcp",
			Address:     make([]byte, 10),
			ExternalIp:  "1.1.1.2:3333",
			Peers:       []string{"1:2222", "2:33333", "3:44444"},
			MaxInbound:  10,
			MaxOutbound: 10,
		}

		assert.EqualError(
			t,
			m.configure(config),
			ErrMissingOrEmptyConfigField("WireHeaderLength").Error(),
			"Configure: failed, expected configure to throw an error on missing fields, but it did not.",
		)
	}

	{ // BufferSize
		config := &shared.P2PConfig{
			Protocol:         "tcp",
			Address:          make([]byte, 10),
			ExternalIp:       "1.1.1.2:3333",
			Peers:            []string{"1:111", "2:2222", "3:33333"},
			MaxInbound:       10,
			MaxOutbound:      10,
			WireHeaderLength: 8,
		}

		assert.EqualError(
			t,
			m.configure(config),
			ErrMissingOrEmptyConfigField("BufferSize").Error(),
			"Configure: failed, expected configure to throw an error on missing fields, but it did not.",
		)
	}

	{ // TimeoutInMs
		config := &shared.P2PConfig{
			Protocol:         "tcp",
			Address:          make([]byte, 10),
			ExternalIp:       "1.1.1.2:3333",
			Peers:            []string{"1:2222", "2:33333", "3:44444"},
			MaxInbound:       10,
			MaxOutbound:      10,
			WireHeaderLength: 8,
			BufferSize:       100,
		}

		assert.EqualError(
			t,
			m.configure(config),
			ErrMissingOrEmptyConfigField("TimeoutInMs").Error(),
			"Configure: failed, expected configure to throw an error on missing fields, but it did not.",
		)
	}

	{ // No throwing
		config := &shared.P2PConfig{
			Protocol:         "tcp",
			Address:          make([]byte, 10),
			ExternalIp:       "1.1.1.2:3333",
			Peers:            []string{"1:2222", "2:33333", "3:44444"},
			MaxInbound:       10,
			MaxOutbound:      10,
			WireHeaderLength: 8,
			BufferSize:       100,
			TimeoutInMs:      10,
		}

		assert.ErrorIs(
			t,
			m.configure(config),
			nil,
			"Configure: failed, expected configure to not throw an error on no missing fields, but it did.",
		)
	}
}

func TestNetworkLifecycle_InitCodec(t *testing.T) {
	m := newP2PModule()

	err := m.initCodec()

	assert.ErrorIs(
		t,
		err,
		nil,
		fmt.Sprintf("initCodec: failed, expected to initialize successfully, got error %s", err),
	)

	assert.NotNil(
		t,
		m.c,
		"initCodec: failed, codec is still nil after initialization",
	)

	actual := m.c.Registered()
	expected := 1
	assert.Equal(
		t,
		actual,
		expected,
		fmt.Sprintf("initCodec: failed, expected to have registered %d messages, got %d", expected, actual),
	)
}

func TestNetworkLifecycle_InitSocketPools(t *testing.T) {
	m := newP2PModule()
	m.config = &shared.P2PConfig{
		Protocol:         "tcp",
		Address:          make([]byte, 10),
		ExternalIp:       "1.1.1.2:3333",
		Peers:            []string{"1:2222", "2:33333", "3:44444"},
		MaxInbound:       10,
		MaxOutbound:      10,
		WireHeaderLength: 8,
		BufferSize:       100,
		TimeoutInMs:      10,
	}

	m.initSocketPools()

	assert.Equal(
		t,
		m.inbound.Capacity(),
		m.config.MaxInbound,
		fmt.Sprintf("initSocketPools: failed, expected initialized inbound pool to have a capacitty of %d, got %d", m.config.MaxInbound, m.inbound.Capacity()),
	)

	assert.Equal(
		t,
		m.outbound.Capacity(),
		m.config.MaxOutbound,
		fmt.Sprintf("initSocketPools: failed, expected initialized outbound pool to have a capacitty of %d, got %d", m.config.MaxOutbound, m.outbound.Capacity()),
	)
}
