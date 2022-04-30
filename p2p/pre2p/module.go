package pre2p

// TODO(team): This is a a temporary parallel to the real `p2p` module.
// It should be removed once the real `p2p` module is ready but is meant
// to be a "real" replacement for now.

import (
	"fmt"
	"log"
	"net"

	"github.com/pokt-network/pocket/p2p/pre2p/raintree"
	typesPre2P "github.com/pokt-network/pocket/p2p/pre2p/types"

	"github.com/pokt-network/pocket/shared/config"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/types"
	typesGenesis "github.com/pokt-network/pocket/shared/types/genesis"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

var _ modules.P2PModule = &p2pModule{}

type p2pModule struct {
	bus modules.Bus

	listener *net.TCPListener
	address  cryptoPocket.Address

	network typesPre2P.Network
}

func Create(cfg *config.Config) (m modules.P2PModule, err error) {
	log.Println("Creating network module")

	tcpAddr, _ := net.ResolveTCPAddr(NetworkProtocol, fmt.Sprintf(":%d", cfg.Pre2P.ConsensusPort))
	l, err := net.ListenTCP(NetworkProtocol, tcpAddr)
	if err != nil {
		return nil, err
	}

	testState := typesGenesis.GetNodeState(nil)
	addrBook, err := ValidatorMapToAddrBook(testState.ValidatorMap)
	if err != nil {
		return nil, err
	}
	var network typesPre2P.Network
	if cfg.Pre2P.UseRainTree {
		selfAddr := cryptoPocket.Address(cfg.PrivateKey.Address())
		network = raintree.NewRainTreeNetwork(selfAddr, addrBook)
	} else {
		network = NewNetwork(addrBook)
	}

	m = &p2pModule{
		listener: l,
		network:  network,
		address:  cfg.PrivateKey.Address(),
	}

	return m, nil
}

func (m *p2pModule) SetBus(bus modules.Bus) {
	m.bus = bus
}

func (m *p2pModule) GetBus() modules.Bus {
	if m.bus == nil {
		log.Fatalf("PocketBus is not initialized")
	}
	return m.bus
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

func (m *p2pModule) Broadcast(msg *anypb.Any, topic types.PocketTopic) error {
	// TODO(olshansky): This should not be a separate interface from `NetworkBroadcast`
	if topic == types.PocketTopic_P2P_PROPAGATE_TOPIC {
		return m.network.NetworkPropagate(msg)
	}

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

func (m *p2pModule) Send(addr cryptoPocket.Address, msg *anypb.Any, topic types.PocketTopic) error {
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

func (m *p2pModule) GetAddrBook() typesPre2P.AddrBook {
	return m.network.GetAddrBook()
}
