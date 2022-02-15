package pre_p2p

import (
	"fmt"
	"log"
	"net"
	"pocket/p2p/pre_p2p/pre_p2p_types"
	"pocket/shared/config"
	"pocket/shared/modules"
	"pocket/shared/types"
	"strconv"

	"google.golang.org/protobuf/proto"
)

type networkModule struct {
	modules.NetworkModule
	pocketBusMod modules.BusModule

	listener *net.TCPListener
	network  pre_p2p_types.Network
	nodeId   pre_p2p_types.NodeId
}

func Create(cfg *config.Config) (m modules.NetworkModule, err error) {
	log.Println("Creating network module")
	p2pState := GetPocketState()
	p2pState.LoadStateFromConfig(cfg)
	tcpAddr, _ := net.ResolveTCPAddr("tcp4", fmt.Sprintf(":%d", cfg.PREP2P.ConsensusPort))
	l, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return nil, err
	}

	m = &networkModule{
		listener: l,
		network:  ConnectToNetwork(state.ValidatorMap),
		nodeId:   pre_p2p_types.NodeId(cfg.Consensus.NodeId),
	}

	return m, nil
}

func (m *networkModule) SetPocketBusMod(pocketBus modules.BusModule) {
	m.pocketBusMod = pocketBus
}

func (m *networkModule) GetBus() modules.BusModule {
	if m.pocketBusMod == nil {
		log.Fatalf("PocketBus is not initialized")
	}
	return m.pocketBusMod
}

func (m *networkModule) Start() error {
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

func (m *networkModule) Stop() error {
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

func (m *networkModule) BroadcastMessage(msg *types.NetworkMessage) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	return m.network.NetworkBroadcast(data, m.nodeId)
}

func (m *networkModule) Send(addr string, msg *types.NetworkMessage) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	// pre_p2p hack
	nodeIdInt, err := strconv.Atoi(addr)
	if err != nil {
		return err
	}
	destNodeId := pre_p2p_types.NodeId(nodeIdInt)

	return m.network.NetworkSend(data, destNodeId)
}

func (m *networkModule) GetNetwork() pre_p2p_types.Network {
	return m.network
}
