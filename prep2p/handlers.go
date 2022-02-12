package prep2p

import (
	"io/ioutil"
	"log"
	"net"

	"pocket/shared/events"
	"pocket/shared/messages"

	"google.golang.org/protobuf/proto"
)

func (m *networkModule) handleNetworkMessage(conn net.Conn) {
	defer conn.Close()

	data, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Println("Error reading from conn: ", err)
		return
	}

	networkMessage := messages.NetworkMessage{}
	proto.Unmarshal(data, &networkMessage)
	// networkMessage, err := DecodeNetworkMessage(data)
	if err != nil {
		log.Println("Error decoding network message: ", err)
		return
	}

	// temporarily convert

	event := events.PocketEvent{
		SourceModule: events.P2P,
		PocketTopic:  networkMessage.Topic,
		MessageData:  networkMessage.Data,
	}

	m.GetPocketBusMod().PublishEventToBus(&event)
}

func (m *networkModule) respondToTelemetryMessage(conn net.Conn) {
	// TODO: quick hack. not running `defer conn.Close()` since the connection is passed
	// to Consensus node for debugging purposes.
	log.Println("Responding to telemetry request...")

	event := events.PocketEvent{
		SourceModule: events.P2P,
		PocketTopic:  string(events.CONSENSUS_TELEMETRY_MESSAGE),

		NetworkConnection: conn,
	}
	m.GetPocketBusMod().PublishEventToBus(&event)
}
