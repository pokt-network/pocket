package modules

import (
	"google.golang.org/protobuf/types/known/anypb"
	"net"
)

type ConsensusModule interface {
	Module
	HandleMessage(*anypb.Any)
	// HandleMessage(*context.PocketContext, *consensus_types.ConsensusMessage)
	HandleTransaction(*anypb.Any)
	HandleEvidence([]byte)

	// Debugging & Telemetry
	HandleTelemetryMessage(net.Conn)
}
