package pre2p

import (
	"fmt"
	"log"
	"net"
	"pocket/pre2p/types"
	"pocket/shared/config"
	"pocket/shared/modules"

	"google.golang.org/protobuf/types/known/anypb"

	"strconv"

	"google.golang.org/protobuf/proto"
)

var _ modules.NetworkModule = &networkModule{}

type networkModule struct {
	pocketBusMod modules.Bus

	listener *net.TCPListener
	network  types.Network
	nodeId   types.NodeId
}

func Create(cfg *config.Config) (m modules.NetworkModule, err error) {
	log.Println("Creating network module")

	p2pState := GetTestState()
	p2pState.LoadStateFromConfig(cfg)

	tcpAddr, _ := net.ResolveTCPAddr("tcp4", fmt.Sprintf(":%d", cfg.PREP2P.ConsensusPort))
	l, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return nil, err
	}

	m = &networkModule{
		listener: l,
		network:  ConnectToValidatorNetwork(state.ValidatorMap),
		nodeId:   types.NodeId(cfg.Consensus.NodeId),
	}

	return m, nil
}

func (m *networkModule) SetBus(pocketBus modules.Bus) {
	m.pocketBusMod = pocketBus
}

func (m *networkModule) GetBus() modules.Bus {
	if m.pocketBusMod == nil {
		log.Fatalf("PocketBus is not initialized")
	}
	return m.pocketBusMod
}

func (m *networkModule) Start() error {
	log.Println("Starting network module")

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

func (m *networkModule) BroadcastMessage(msg *anypb.Any, topic string) error {
	c := &types.P2PMessage{
		Topic: topic, // TODO topic is either P2P (from this module) or consensus
		Data:  msg,
	}
	data, err := proto.Marshal(c)
	if err != nil {
		return err
	}
	fmt.Println("broadcasting message to network")
	return m.network.NetworkBroadcast(data, m.nodeId)
}

func (m *networkModule) Send(addr string, msg *anypb.Any, topic string) error {
	c := &types.P2PMessage{
		Topic: topic, // TODO(discuss): Is this the approach we want to go with for P2PMessages?
		Data:  msg,
	}
	data, err := proto.Marshal(c)
	if err != nil {
		return err
	}

	// TODO(olshansky): This is a hack while the consensus module is still dependant on `NodeId`.
	nodeIdInt, err := strconv.Atoi(addr)
	if err != nil {
		return err
	}
	destNodeId := types.NodeId(nodeIdInt)

	return m.network.NetworkSend(data, destNodeId)
}

func (m *networkModule) GetAddrBook() []*types.NetworkPeer {
	return m.network.GetAddrBook()
}
