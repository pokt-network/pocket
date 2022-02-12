package prep2p

import (
	"fmt"
	"log"
	"net"

	"pocket/consensus/pkg/config"
	"pocket/consensus/pkg/types"
	"pocket/prep2p/pre_p2p_types"
	"pocket/shared"
	"pocket/shared/context"
	"pocket/shared/messages"
	"pocket/shared/modules"

	"google.golang.org/protobuf/proto"
)

type networkModule struct {
	modules.NetworkModule
	pocketBusMod modules.PocketBusModule

	listener *net.TCPListener
	network  pre_p2p_types.Network
	nodeId   types.NodeId
}

func Create(cfg *config.Config) (m modules.NetworkModule, err error) {
	log.Println("Creating network module")

	tcpAddr, _ := net.ResolveTCPAddr("tcp4", fmt.Sprintf(":%d", cfg.PREP2P.ConsensusPort))
	l, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return nil, err
	}

	state := shared.GetPocketState()

	m = &networkModule{
		listener: l,
		network:  ConnectToNetwork(state.ValidatorMap),
		nodeId:   cfg.Consensus.NodeId,
	}

	return m, nil
}

func (m *networkModule) SetPocketBusMod(pocketBus modules.PocketBusModule) {
	m.pocketBusMod = pocketBus
}

func (m *networkModule) GetPocketBusMod() modules.PocketBusModule {
	if m.pocketBusMod == nil {
		log.Fatalf("PocketBus is not initialized")
	}
	return m.pocketBusMod
}

func (m *networkModule) Start(ctx *context.PocketContext) error {
	log.Println("Starting network module")

	// Struct telementry TCP server.
	go func() {
		tcpAddr, _ := net.ResolveTCPAddr("tcp4", fmt.Sprintf(":%d", 9080))
		l, err := net.ListenTCP("tcp", tcpAddr)
		if err != nil {
			log.Fatal("Listen error: ", err)
		}
		for {
			conn, err := l.AcceptTCP()
			if err != nil {
				log.Println("Error accepting connection: ", err)
				continue
			}
			go m.respondToTelemetryMessage(conn)
		}
	}()

	// Start consensus TCP server
	go func() {
		for {
			conn, err := m.listener.AcceptTCP()
			if err != nil {
				log.Println("Error accepting connection: ", err)
				continue
			}
			go m.handleNetworkMessage(conn)
		}
	}()

	return nil
}

func (m *networkModule) Stop(ctx *context.PocketContext) error {
	log.Println("Stopping network module")
	if err := m.listener.Close(); err != nil {
		return err
	}
	return nil
}

// func (m *networkModule) Broadcast(ctx *context.PocketContext, message *pre_p2p_types.NetworkMessage) error {
// 	data, err := EncodeNetworkMessage(message)
// 	if err != nil {
// 		return err
// 	}
// 	return m.network.NetworkBroadcast(data, m.nodeId)
// }

// func (m *networkModule) Send(ctx *context.PocketContext, message *pre_p2p_types.NetworkMessage, destNodeId types.NodeId) error {
// 	data, err := EncodeNetworkMessage(message)
// 	if err != nil {
// 		return err
// 	}
// 	return m.network.NetworkSend(data, destNodeId)
// }

func (m *networkModule) BroadcastMessage(msg *messages.NetworkMessage) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	return m.network.NetworkBroadcast(data, m.nodeId)
}

func (m *networkModule) Send(addr string, msg *messages.NetworkMessage) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	destNodeId := types.NodeId(1) // Get from addr
	return m.network.NetworkSend(data, destNodeId)
}

func (m *networkModule) GetNetwork() pre_p2p_types.Network {
	return m.network
}
