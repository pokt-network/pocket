package modules

import (
	"net"

	consensus_types "pocket/consensus/pkg/consensus/types"
	"pocket/consensus/pkg/shared/context"
)

type ConsensusModule interface {
	PocketModule

	HandleMessage(*context.PocketContext, *consensus_types.ConsensusMessage)
	HandleTransaction(*context.PocketContext, []byte)
	HandleEvidence(*context.PocketContext, []byte)

	// Debugging & Telemetry
	HandleTelemetryMessage(*context.PocketContext, net.Conn)
	GetNodeState() consensus_types.ConsensusNodeState
}
