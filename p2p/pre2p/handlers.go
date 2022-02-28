package pre2p

import (
	"io/ioutil"
	"log"
	"net"

	pre2ptypes "pocket/pre2p/types"
	"pocket/shared/types"

	"google.golang.org/protobuf/proto"
)

func (m *p2pModule) handleNetworkMessage(conn net.Conn) {
	defer conn.Close()

	data, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Println("Error reading from conn: ", err)
		return
	}

	networkMessage := pre2ptypes.P2PMessage{}
	if err := proto.Unmarshal(data, &networkMessage); err != nil {
		panic(err)
	}
	if err != nil {
		log.Println("Error decoding network message: ", err)
		return
	}

	event := types.Event{
		SourceModule: types.P2P,
		PocketTopic:  networkMessage.Topic,
		MessageData:  networkMessage.Data,
	}

	m.GetBus().PublishEventToBus(&event)
}
