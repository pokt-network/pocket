package protocol

import "github.com/libp2p/go-libp2p/core/protocol"

const (
	// PoktProtocolID is the libp2p protocol ID used when opening a new stream
	// to a remote peer and setting the stream handler for the local peer.
	// Libp2p APIs use this to distinguish which multiplexed protocols/streams to consider.
	PoktProtocolID = protocol.ID("pokt/v1.0.0")
	// DefaultTopicStr is a "default" pubsub topic string for use when subscribing.
	DefaultTopicStr = "pokt/default"
)
