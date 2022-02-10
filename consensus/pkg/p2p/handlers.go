package p2p

import (
	"io/ioutil"
	"log"
	"net"

	"pocket/consensus/pkg/shared/events"
)

func (m *networkModule) handleNetworkMessage(conn net.Conn) {
	defer conn.Close()

	data, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Println("Error reading from conn: ", err)
		return
	}

	networkMessage, err := DecodeNetworkMessage(data)
	if err != nil {
		log.Println("Error decoding network message: ", err)
		return
	}

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
		PocketTopic:  events.CONSENSUS_TELEMETRY_MESSAGE,

		NetworkConnection: conn,
	}
	m.GetPocketBusMod().PublishEventToBus(&event)
}
