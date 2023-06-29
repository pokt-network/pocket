package host

import (
	"fmt"
	"strings"
)

////////////////////////////////////////////////////////////////////////////////
// ICS04
// The following paths are the keys to the store as defined in:
// https://github.com/cosmos/ibc/tree/master/spec/core/ics-004-channel-and-packet-semantics#store-paths
////////////////////////////////////////////////////////////////////////////////

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
