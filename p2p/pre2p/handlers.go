package pre2p

import (
	"io/ioutil"
	"log"
	"net"

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

	networkMessage := types.PocketEvent{}
	if err := proto.Unmarshal(data, &networkMessage); err != nil {
		panic(err)
	}
	if err != nil {
		log.Println("Error decoding network message: ", err)
		return
	}

	event := types.PocketEvent{
		Module: types.SourceModule_P2P_MODULE,
		Topic:  networkMessage.Topic,
		Data:   networkMessage.Data,
	}

	m.GetBus().PublishEventToBus(&event)
}
