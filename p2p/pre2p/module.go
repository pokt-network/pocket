package pre2p

// TODO(team): This is a Top Level Module since it is a temporary parallel to the
// real `p2p` module. It should be removed once the real `p2p` module is ready but
// is meant to be a "real" replacement for now.

import (
	"fmt"
	"log"
	"net"
	"pocket/p2p/pre2p/types"
	"pocket/shared/config"
	"pocket/shared/modules"

	"google.golang.org/protobuf/types/known/anypb"

	"strconv"

	"google.golang.org/protobuf/proto"
)

var _ modules.P2PModule = &p2pModule{}

type p2pModule struct {
	pocketBusMod modules.Bus

	listener *net.TCPListener
	network  types.Network
	nodeId   types.NodeId
}

func Create(cfg *config.Config) (m modules.P2PModule, err error) {
	log.Println("Creating network module")

	p2pState := GetTestState()
	p2pState.LoadStateFromConfig(cfg)

	tcpAddr, _ := net.ResolveTCPAddr("tcp4", fmt.Sprintf(":%d", cfg.PREP2P.ConsensusPort))
	l, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return nil, err
	}

	m = &p2pModule{
		listener: l,
		network:  ConnectToValidatorNetwork(state.ValidatorMap),
		nodeId:   types.NodeId(cfg.Consensus.NodeId),
	}

	return m, nil
}

func (m *p2pModule) SetBus(pocketBus modules.Bus) {
	m.pocketBusMod = pocketBus
}

func (m *p2pModule) GetBus() modules.Bus {
	if m.pocketBusMod == nil {
		log.Fatalf("PocketBus is not initialized")
	}
	return m.pocketBusMod
}

func (m *p2pModule) Start() error {
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

func (m *p2pModule) Stop() error {
	log.Println("Stopping network module")
	if err := m.listener.Close(); err != nil {
		return err
	}
	return nil
}

func (m *p2pModule) BroadcastMessage(msg *anypb.Any, topic string) error {
	c := &types.P2PMessage{
		Topic: topic, // TODO topic is either P2P (from this module) or consensus
		Data:  msg,
	}
	data, err := proto.Marshal(c)
	if err != nil {
		return err
	}
	log.Println("broadcasting message to network")
	return m.network.NetworkBroadcast(data, m.nodeId)
}

func (m *p2pModule) Send(addr string, msg *anypb.Any, topic string) error {
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

func (m *p2pModule) GetAddrBook() []*types.NetworkPeer {
	return m.network.GetAddrBook()
}
