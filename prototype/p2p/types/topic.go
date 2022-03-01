package types

import "strings"

var (
	Topics = struct {
		Consensus string
		Ping      string
		Pong      string
		TxMsg     string
	}{
		Consensus: "consensus",
		Ping:      "ping",
		Pong:      "pong",
		TxMsg:     "transaction",
	}
)

func Topic(topic string) PocketTopic {
	switch strings.ToLower(topic) {
	case Topics.Consensus:
		return PocketTopic_CONSENSUS
	case Topics.Ping:
		return PocketTopic_P2P_PING
	case Topics.Pong:
		return PocketTopic_P2P_PING
	case Topics.TxMsg:
		return PocketTopic_TRANSACTION
	}
	return PocketTopic_UNDEFINED
}
