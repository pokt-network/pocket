package protocol

import "github.com/libp2p/go-libp2p/core/protocol"

// PoktProtocolID is the libp2p protocol ID matching current version of the pokt protocol.
var PoktProtocolID = protocol.ID("pokt/v1.0.0")

// DefaultTopicStr is a "default" pubsub topic string for use when subscribing.
var DefaultTopicStr = "pokt/default"
