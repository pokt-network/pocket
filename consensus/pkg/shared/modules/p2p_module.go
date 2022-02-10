package modules

import (
	"pocket/consensus/pkg/p2p/p2p_types"
	"pocket/consensus/pkg/shared/context"
	"pocket/consensus/pkg/types"
)

type NetworkModule interface {
	PocketModule

	Broadcast(*context.PocketContext, *p2p_types.NetworkMessage) error
	Send(*context.PocketContext, *p2p_types.NetworkMessage, types.NodeId) error
	GetNetwork() p2p_types.Network
}
