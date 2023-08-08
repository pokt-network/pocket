//go:build !debug

package p2p

import "github.com/pokt-network/pocket/shared/messaging"

func (m *p2pModule) handleDebugMessage(_ *messaging.DebugMessage) error {
	return nil
}
