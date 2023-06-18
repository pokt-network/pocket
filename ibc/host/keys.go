package host

import (
	"fmt"
	"strings"
)

// Store key prefixes for IBC
const (
	KeyClientStorePrefix       = "clients"
	KeyClientState             = "clientState"
	KeyConsensusStatePrefix    = "consensusStates"
	KeyConnectionPrefix        = "connections"
	KeyChannelEndPrefix        = "channelEnds"
	KeyChannelPrefix           = "channels"
	KeyPortPrefix              = "ports"
	KeySequencePrefix          = "sequences"
	KeyChannelCapabilityPrefix = "capabilities"
	KeyNextSeqSendPrefix       = "nextSequenceSend"
	KeyNextSeqRecvPrefix       = "nextSequenceRecv"
	KeyNextSeqAckPrefix        = "nextSequenceAck"
	KeyPacketCommitmentPrefix  = "commitments"
	KeyPacketAckPrefix         = "acks"
	KeyPacketReceiptPrefix     = "receipts"
)

// DISCUSSION: Do we need both paths and keys with the ApplyPrefix and RemovePrefix functions?
// These seem to be redundant and could be removed, but are included in the cosmos/ibc-go repo

// fullClientPath returns the full path of a specific client path in the format:
// "clients/{clientID}/{key}" as a string.
func fullClientPath(clientID, key string) string {
	return fmt.Sprintf("%s/%s/%s", KeyClientStorePrefix, clientID, key)
}

// clientPath returns the path of a specific client within the client store in the format:
// "{clientID}/{key}" as a string
func clientPath(clientID, key string) string {
	return strings.TrimPrefix(fullClientPath(clientID, key), KeyClientStorePrefix+"/")
}

// FullClientKey returns the full path of specific client path in the format:
// "clients/{clientID}/{key}" as a byte array.
func fullClientKey(clientID, key string) []byte {
	return []byte(fullClientPath(clientID, key))
}

// ICS02
// The following paths are the keys to the store as defined in
// https://github.com/cosmos/ibc/tree/master/spec/core/ics-002-client-semantics#path-space

// FullClientStateKey takes a client identifier and returns a Key under which to store a
// particular client state.
func FullClientStateKey(clientID string) []byte {
	return fullClientKey(clientID, KeyClientState)
}

// ClientStatePath takes a client identifier and returns a Path string where it can be accessed
// within the client store
func ClientStatePath(clientID string) string {
	return clientPath(clientID, KeyClientState)
}

// consensusStatePath returns the suffix store key for the consensus state at a
// particular height stored in a client prefixed store.
func consensusStatePath(height uint64) string {
	return fmt.Sprintf("%s/%d", KeyConsensusStatePrefix, height)
}

// fullConsensusStatePath takes a client identifier and returns a Path under which to
// store the consensus state of a client.
func fullConsensusStatePath(clientID string, height uint64) string {
	return fullClientPath(clientID, consensusStatePath(height))
}

// FullConsensusStateKey returns the store key for the consensus state of a particular client.
func FullConsensusStateKey(clientID string, height uint64) []byte {
	return []byte(fullConsensusStatePath(clientID, height))
}

// ConsensusStatePath takes a client identifier and height and returns the Path where the consensus
// state can be accessed in the client store
func ConsensusStatePath(clientID string, height uint64) string {
	return clientPath(clientID, consensusStatePath(height))
}

// ICS03
// The following paths are the keys to the store as defined in:
// https://github.com/cosmos/ibc/blob/master/spec/core/ics-003-connection-semantics#store-paths

// clientConnectionsPath defines a reverse mapping from clients to a set of connections
func clientConnectionsPath(clientID string) string {
	return fullClientPath(clientID, KeyConnectionPrefix)
}

// ClientConnectionsKey returns the store key for the connections of a given client
func ClientConnectionsKey(clientID string) []byte {
	return []byte(clientConnectionsPath(clientID))
}

// ClientConnectionPath defines the path under which the connections are stored in the client store
func ClientConnectionPath(clientID string) string {
	return clientPath(clientID, KeyConnectionPrefix)
}

// connectionPath defines the path under which connection paths are stored
func connectionPath(connectionID string) string {
	return fmt.Sprintf("%s/%s", KeyConnectionPrefix, connectionID)
}

// ConnectionKey returns the store key for a particular connection
func ConnectionKey(connectionID string) []byte {
	return []byte(connectionPath(connectionID))
}

// ConnectionPath returns the path for a particular connection in the connections store
func ConnectionPath(connectionID string) string {
	return connectionID
}

// ICS04
// The following paths are the keys to the store as defined in:
// https://github.com/cosmos/ibc/tree/master/spec/core/ics-004-channel-and-packet-semantics#store-paths

func channelPath(portID, channelID string) string {
	return fmt.Sprintf("%s/%s/%s/%s", KeyPortPrefix, portID, KeyChannelPrefix, channelID)
}

// fullChannelPath defines the path under which channels are stored
func fullChannelPath(portID, channelID string) string {
	return fmt.Sprintf("%s/%s", KeyChannelEndPrefix, channelPath(portID, channelID))
}

// ChannelKey returns the store key for a particular channel
func ChannelKey(portID, channelID string) []byte {
	return []byte(fullChannelPath(portID, channelID))
}

// ChannelPath returns the path under which a particular channel is stored in the ChannelEnd store
func ChannelPath(portID, channelID string) string {
	return channelPath(portID, channelID)
}

// channelCapabilityPath defines the path under which capability keys associated
// with a channel are stored
func channelCapabilityPath(portID, channelID string) string {
	return fmt.Sprintf("%s/%s", KeyChannelCapabilityPrefix, channelPath(portID, channelID))
}

// ChannelCapabilityKey returns the store key for the capability associated with a channel
func ChannelCapabilityKey(portID, channelID string) []byte {
	return []byte(channelCapabilityPath(portID, channelID))
}

// ChannelCapabilityPath returns the path under which a particular channel capability is stored
// in the channel capability store
func ChannelCapabilityPath(portID, channelID string) string {
	return strings.TrimPrefix(channelCapabilityPath(portID, channelID), KeyChannelCapabilityPrefix+"/")
}

// nextSequenceSendPath defines the next send sequence counter store path
func nextSequenceSendPath(portID, channelID string) string {
	return fmt.Sprintf("%s/%s", KeyNextSeqSendPrefix, channelPath(portID, channelID))
}

// NextSequenceSendKey returns the store key for the send sequence of a particular
// channel binded to a specific port.
func NextSequenceSendKey(portID, channelID string) []byte {
	return []byte(nextSequenceSendPath(portID, channelID))
}

// NextSequenceSendPath returns the path under which the NextSequenceSend is stored in the NextSequenceSend store
func NextSequenceSendPath(portID, channelID string) string {
	return channelPath(portID, channelID)
}

// nextSequenceRecvPath defines the next receive sequence counter store path.
func nextSequenceRecvPath(portID, channelID string) string {
	return fmt.Sprintf("%s/%s", KeyNextSeqRecvPrefix, channelPath(portID, channelID))
}

// NextSequenceRecvKey returns the store key for the receive sequence of a particular
// channel binded to a specific port
func NextSequenceRecvKey(portID, channelID string) []byte {
	return []byte(nextSequenceRecvPath(portID, channelID))
}

// NextSequenceRecvPath returns the path under which the NextSequenceRecv is stored in the NextSequenceRecv store
func NextSequenceRecvPath(portID, channelID string) string {
	return channelPath(portID, channelID)
}

// nextSequenceAckPath defines the next acknowledgement sequence counter store path
func nextSequenceAckPath(portID, channelID string) string {
	return fmt.Sprintf("%s/%s", KeyNextSeqAckPrefix, channelPath(portID, channelID))
}

// NextSequenceAckKey returns the store key for the acknowledgement sequence of
// a particular channel binded to a specific port.
func NextSequenceAckKey(portID, channelID string) []byte {
	return []byte(nextSequenceAckPath(portID, channelID))
}

// NextSequenceAckPath returns the path under which the NextSequenceAck is stored in the NextSequenceAck store
func NextSequenceAckPath(portID, channelID string) string {
	return channelPath(portID, channelID)
}

// packetCommitmentPrefixPath defines the prefix for commitments to packet data fields store path.
func packetCommitmentPrefixPath(portID, channelID string) string {
	return fmt.Sprintf("%s/%s/%s", KeyPacketCommitmentPrefix, channelPath(portID, channelID), KeySequencePrefix)
}

// packetCommitmentPath defines the commitments to packet data fields store path
func packetCommitmentPath(portID, channelID string, sequence uint64) string {
	return fmt.Sprintf("%s/%d", packetCommitmentPrefixPath(portID, channelID), sequence)
}

// PacketCommitmentKey returns the store key of under which a packet commitment is stored
func PacketCommitmentKey(portID, channelID string, sequence uint64) []byte {
	return []byte(packetCommitmentPath(portID, channelID, sequence))
}

// PacketCommitmentPath returns the path under which the PacketCommitment is stored in the PacketCommitment store
func PacketCommitmentPath(portID, channelID string, sequence uint64) string {
	return strings.TrimPrefix(packetCommitmentPath(portID, channelID, sequence), KeyPacketCommitmentPrefix+"/")
}

// packetAcknowledgementPrefixPath defines the prefix for commitments to packet data fields store path.
func packetAcknowledgementPrefixPath(portID, channelID string) string {
	return fmt.Sprintf("%s/%s/%s", KeyPacketAckPrefix, channelPath(portID, channelID), KeySequencePrefix)
}

// packetAcknowledgementPath defines the packet acknowledgement store path
func packetAcknowledgementPath(portID, channelID string, sequence uint64) string {
	return fmt.Sprintf("%s/%d", packetAcknowledgementPrefixPath(portID, channelID), sequence)
}

// PacketAcknowledgementKey returns the store key of under which a packet acknowledgement is stored
func PacketAcknowledgementKey(portID, channelID string, sequence uint64) []byte {
	return []byte(packetAcknowledgementPath(portID, channelID, sequence))
}

// PacketAcknowledgementPath returns the path under which the PacketAcknowledgement is stored in the PacketAcknowledgement store
func PacketAcknowledgementPath(portID, channelID string, sequence uint64) string {
	return strings.TrimPrefix(packetAcknowledgementPath(portID, channelID, sequence), KeyPacketAckPrefix+"/")
}

func sequencePath(sequence uint64) string {
	return fmt.Sprintf("%s/%d", KeySequencePrefix, sequence)
}

// packetReceiptPath defines the packet receipt store path
func packetReceiptPath(portID, channelID string, sequence uint64) string {
	return fmt.Sprintf("%s/%s/%s", KeyPacketReceiptPrefix, channelPath(portID, channelID), sequencePath(sequence))
}

// PacketReceiptKey returns the store key of under which a packet receipt is stored
func PacketReceiptKey(portID, channelID string, sequence uint64) []byte {
	return []byte(packetReceiptPath(portID, channelID, sequence))
}

// PacketReceiptPath returns the path under which the PacketReceipt is stored in the PacketReceipt store
func PacketReceiptPath(portID, channelID string, sequence uint64) string {
	return strings.TrimPrefix(packetReceiptPath(portID, channelID, sequence), KeyPacketReceiptPrefix+"/")
}

// ICS05
// The following paths are the keys to the store as defined in
// https://github.com/cosmos/ibc/tree/master/spec/core/ics-005-port-allocation#store-paths

// portPath defines the path under which ports paths are stored on the capability module
func portPath(portID string) string {
	return fmt.Sprintf("%s/%s", KeyPortPrefix, portID)
}

// PortKey returns the store key for a port in the capability module
func PortKey(portID string) []byte {
	return []byte(portPath(portID))
}

// PortPath returns the path under which the Port is stored in the Port store
func PortPath(portID string) string {
	return portID
}
