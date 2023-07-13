package p2p

import (
	"fmt"

	"github.com/pokt-network/pocket/p2p/debug"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/shared/messaging"
)

func (m *p2pModule) handleDebugMessage(msg *messaging.DebugMessage) error {
	switch msg.Action {
	case messaging.DebugMessageAction_DEBUG_P2P_PEER_LIST:
		if !m.cfg.EnablePeerDiscoveryDebugRpc {
			return typesP2P.ErrPeerDiscoveryDebugRPCDisabled
		}
	default:
		// This debug message isn't intended for the P2P module, ignore it.
		return nil
	}

	switch msg.Action {
	case messaging.DebugMessageAction_DEBUG_P2P_PEER_LIST:
		routerType := debug.RouterType(msg.Message.Value)
		return debug.PrintPeerList(m.GetBus(), routerType)
	default:
		return fmt.Errorf("unsupported P2P debug message action: %s", msg.Action)
	}
}
