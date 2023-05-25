package service

import (
	"fmt"
	"testing"

	"github.com/pokt-network/pocket/runtime/configs"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
)

func TestAdmitRelay(t *testing.T) {
	testCases := []struct {
		name          string
		config        configs.ServicerConfig
		relay         coreTypes.Relay
		currentHeight int64
		session       *coreTypes.Session
		errSession    error
		expected      error
	}{
		{
			name:     "Relay with empty Meta is rejected",
			expected: fmt.Errorf("Error admitting relay: relay metadata failed validation: empty relay metadata"),
		},
		{
			name:     "Relay with unspecified chain is rejected",
			relay:    coreTypes.Relay{Meta: &coreTypes.RelayMeta{}},
			expected: fmt.Errorf("Error admitting relay: relay metadata failed validation: relay chain unspecified"),
		},
		{
			name:     "Relay for unsupported chain is rejected",
			relay:    testRelay("0021", 0),
			expected: fmt.Errorf("Error admitting relay: relay metadata failed validation: relay chain not supported: %s", "0021"),
		},
		{
			name:     "Relay with height set in a past session is rejected",
			relay:    testRelay("0021", 5),
			config:   testServicerConfig("0021"),
			session:  testSession(sessionNumber(2), sessionBlocks(4), sessionHeight(2)),
			expected: fmt.Errorf("Error admitting relay: relay failed block height validation: relay block height 5 not within session ID session-1 starting block 8 and last block 10"),
		},
		{
			name:     "Relay with height set in a future session is rejected",
			relay:    testRelay("0021", 9999),
			config:   testServicerConfig("0021"),
			session:  testSession(sessionNumber(2), sessionBlocks(4), sessionHeight(2)),
			expected: fmt.Errorf("Error admitting relay: relay failed block height validation: relay block height 9999 not within session ID session-1 starting block 8 and last block 10"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			servicer := servicer{
				provider: mockProvider{
					Height:     tc.currentHeight,
					Session:    tc.session,
					ErrSession: tc.errSession,
				},
				config: tc.config,
			}

			err := servicer.admitRelay(tc.relay)
			switch {
			case err == nil && tc.expected != nil:
				t.Fatalf("Expected error %v, got: nil", tc.expected)
			case err != nil && tc.expected == nil:
				t.Fatalf("Unexpected error %v", err)
			case err != nil && err.Error() != tc.expected.Error():
				t.Fatalf("Expected error %v, got: %v", tc.expected, err)
			}
		})
	}
}

type mockProvider struct {
	Height     int64
	Session    *coreTypes.Session
	ErrSession error
}

func (m mockProvider) CurrentHeight() int64 {
	return m.Height
}

func (m mockProvider) GetSession(appAddr string, height int64, relayChain, geoZone string) (*coreTypes.Session, error) {
	return m.Session, m.ErrSession
}

func testRelay(chain string, height int64) coreTypes.Relay {
	return coreTypes.Relay{
		Meta: &coreTypes.RelayMeta{
			BlockHeight: height,
			RelayChain: &coreTypes.Identifiable{
				Id: chain,
			},
			GeoZone: &coreTypes.Identifiable{
				Id: "geozone",
			},
		},
	}
}

func testServicerConfig(chains ...string) configs.ServicerConfig {
	return configs.ServicerConfig{
		Chains: append([]string{}, chains...),
	}
}

type sessionModifier func(*coreTypes.Session)

func sessionNumber(number int64) func(*coreTypes.Session) {
	return func(session *coreTypes.Session) {
		session.SessionNumber = number
	}
}

func sessionBlocks(blocksPerSession int64) func(*coreTypes.Session) {
	return func(session *coreTypes.Session) {
		session.NumSessionBlocks = blocksPerSession
	}
}

func sessionHeight(height int64) func(*coreTypes.Session) {
	return func(session *coreTypes.Session) {
		session.SessionHeight = height
	}
}

func testSession(editors ...sessionModifier) *coreTypes.Session {
	session := coreTypes.Session{
		Id: "session-1",
	}
	for _, editor := range editors {
		editor(&session)
	}
	return &session
}
