package protocol

import "github.com/libp2p/go-libp2p/core/protocol"

const (
	// RaintreeProtocolID is the libp2p protocol ID used in the Raintree router
	// when opening a new stream to a remote peer and setting the stream handler
	// for the local peer. Libp2p APIs use this to distinguish which multiplexed
	// protocols/streams to consider.
	RaintreeProtocolID = protocol.ID("pokt/raintree/v1.0.0")
	// BackgroundProtocolID is the libp2p protocol ID used in the Background router
	// when opening a new stream to a remote peer and setting the stream handler
	// for the local peer. Libp2p APIs use this to distinguish which multiplexed
	// protocols/streams to consider.
	BackgroundProtocolID = protocol.ID("pokt/background/v1.0.0")
	// BackgroundTopicStr is a "default" pubsub topic string used when
	// subscribing and broadcasting.
	BackgroundTopicStr = "pokt/background"
	// PeerDiscoveryNamespace used by both advertiser and discoverer to rendezvous
	// during peer discovery. Advertiser(s) and discoverer(s) MUST have matching
	// discovery namespaces to find one another.
	PeerDiscoveryNamespace = "pokt/peer_discovery"
)
