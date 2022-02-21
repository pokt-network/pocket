package types

var (
	Topics = struct {
		Consensus string
		Ping      string
		Pong      string
	}{
		Consensus: "consensus",
		Ping:      "ping",
		Pong:      "pong",
	}
)

func Topic(topic string) PocketTopic {
	switch topic {
	case Topics.Consensus:
		return PocketTopic_CONSENSUS
	case Topics.Ping:
		return PocketTopic_P2P_PING
	case Topics.Pong:
		return PocketTopic_P2P_PING
	}
	return PocketTopic_UNDEFINED
}
