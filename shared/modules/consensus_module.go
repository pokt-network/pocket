package modules

import (
	"net"

	consensus_types "pocket/consensus/pkg/consensus/types"
	"pocket/shared/context"

	"google.golang.org/protobuf/types/known/anypb"
)

type ConsensusModule interface {
	PocketModule

	HandleMessage(*context.PocketContext, *anypb.Any)
	// HandleMessage(*context.PocketContext, *consensus_types.ConsensusMessage)
	HandleTransaction(*context.PocketContext, *anypb.Any)
	HandleEvidence(*context.PocketContext, []byte)

	// Debugging & Telemetry
	HandleTelemetryMessage(*context.PocketContext, net.Conn)
	GetNodeState() consensus_types.ConsensusNodeState
}
