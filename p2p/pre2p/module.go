package pre2p

// TODO(team): This is a a temporary parallel to the real `p2p` module.
// It should be removed once the real `p2p` module is ready but is meant
// to be a "real" replacement for now.

import (
	"fmt"
	"log"
	"net"

	pre2ptypes "github.com/pokt-network/pocket/p2p/pre2p/types"

	"github.com/pokt-network/pocket/shared/config"
	pcrypto "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/types"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

var _ modules.P2PModule = &p2pModule{}

type p2pModule struct {
	pocketBusMod modules.Bus

	listener *net.TCPListener
	network  pre2ptypes.Network
	address  pcrypto.Address
}

func Create(cfg *config.Config) (m modules.P2PModule, err error) {
	log.Println("Creating network module")

	p2pState := GetTestState()
	p2pState.LoadStateFromConfig(cfg)

	tcpAddr, _ := net.ResolveTCPAddr("tcp4", fmt.Sprintf(":%d", cfg.Pre2P.ConsensusPort))
	l, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return nil, err
	}

	m = &p2pModule{
		listener: l,
		network:  ConnectToValidatorNetwork(state.ValidatorMap),
		address:  cfg.PrivateKey.Address(),
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

func (m *p2pModule) BroadcastMessage(msg *anypb.Any, topic types.PocketTopic) error {
	c := &types.PocketEvent{
		Topic: topic,
		Data:  msg,
	}
	data, err := proto.Marshal(c)
	if err != nil {
		return err
	}
	log.Println("broadcasting message to network")
	return m.network.NetworkBroadcast(data)
}

func (m *p2pModule) Send(addr pcrypto.Address, msg *anypb.Any, topic types.PocketTopic) error {
	c := &types.PocketEvent{
		Topic: topic,
		Data:  msg,
	}
	data, err := proto.Marshal(c)
	if err != nil {
		return err
	}

	return m.network.NetworkSend(data, addr)
}

func (m *p2pModule) GetAddrBook() []*pre2ptypes.NetworkPeer {
	return m.network.GetAddrBook()
}
