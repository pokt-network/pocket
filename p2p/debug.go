//go:build debug

package p2p

import (
	"fmt"

	libp2pHost "github.com/libp2p/go-libp2p/core/host"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/shared/messaging"
)

type P2PModule = p2pModule

func (m *p2pModule) handleDebugMessage(msg *messaging.DebugMessage) error {
	switch msg.Action {
	case messaging.DebugMessageAction_DEBUG_P2P_PRINT_PEER_LIST:
		if !m.cfg.EnablePeerDiscoveryDebugRpc {
			return typesP2P.ErrPeerDiscoveryDebugRPCDisabled
		}
		routerType := RouterType(msg.Message.Value)
		return PrintPeerList(m.GetBus(), routerType)
	default:
		return fmt.Errorf("unsupported P2P debug message action: %s", msg.Action)
	}
}

func (m *p2pModule) GetLibp2pHost() libp2pHost.Host {
	return m.host
}
