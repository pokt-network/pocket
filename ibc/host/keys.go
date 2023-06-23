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
