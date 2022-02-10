package p2p

import (
	"fmt"
	"log"
	"net"

	"pocket/consensus/pkg/p2p/p2p_types"
	"pocket/consensus/pkg/shared"
	"pocket/consensus/pkg/shared/context"
	"pocket/consensus/pkg/shared/modules"
	"pocket/consensus/pkg/types"
)

type networkModule struct {
	*modules.BasePocketModule
	modules.NetworkModule

	listener *net.TCPListener
	network  p2p_types.Network
	nodeId   types.NodeId
}

func Create(ctx *context.PocketContext, base *modules.BasePocketModule) (m modules.NetworkModule, err error) {
	log.Println("Creating network module")

	cfg := base.GetConfig()

	tcpAddr, _ := net.ResolveTCPAddr("tcp4", fmt.Sprintf(":%d", cfg.P2P.ConsensusPort))
	l, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return nil, err
	}

	state := shared.GetPocketState()

	m = &networkModule{
		BasePocketModule: base,
		listener:         l,
		network:          ConnectToNetwork(state.ValidatorMap),
		nodeId:           cfg.Consensus.NodeId,
	}

	return m, nil
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

func (m *networkModule) GetPocketBusMod() modules.PocketBusModule {
	return m.BasePocketModule.GetPocketBusMod()
}

func (m *networkModule) SetPocketBusMod(bus modules.PocketBusModule) {
	m.BasePocketModule.SetPocketBusMod(bus)
}

func (m *networkModule) Broadcast(ctx *context.PocketContext, message *p2p_types.NetworkMessage) error {
	data, err := EncodeNetworkMessage(message)
	if err != nil {
		return err
	}
	return m.network.NetworkBroadcast(data, m.nodeId)
}

func (m *networkModule) Send(ctx *context.PocketContext, message *p2p_types.NetworkMessage, destNodeId types.NodeId) error {
	data, err := EncodeNetworkMessage(message)
	if err != nil {
		return err
	}
	return m.network.NetworkSend(data, destNodeId)
}

func (m *networkModule) GetNetwork() p2p_types.Network {
	return m.network
}
