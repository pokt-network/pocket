package p2p

import (
	"fmt"
	"log"
	"net"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/shared/config"
	"github.com/pokt-network/pocket/shared/modules"
	mocks "github.com/pokt-network/pocket/shared/modules/mocks"
	commonTypes "github.com/pokt-network/pocket/shared/types"
	shared "github.com/pokt-network/pocket/shared/types"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

func Setup_TestP2PModule_Case() modules.P2PModule {
	// generate the peer list
	peers := []string{}
	for i := 0; i < 27; i++ {
		peers = append(peers, fmt.Sprintf("%d@127.0.0.1:%d", i+1, i+1+10000))
	}

	m, err := Create(&config.Config{
		P2P: &config.P2PConfig{
			ExternalIp: "127.0.0.1:10001",
			BufferSize: 1024,
			Peers:      peers,
			ID:         1,
			Redundancy: false,
		},
	})

	if err != nil {
		log.Fatal("Fatal! could not instantiate the p2p module")
	}

	return m
}

func TestP2PModule_Start(t *testing.T) {
	m := Setup_TestP2PModule_Case()

	ctrl := gomock.NewController(t)
	bus := mocks.NewMockBus(ctrl)
	m.SetBus(bus)

	err := m.Start()

	assert.Nil(
		t,
		err,
		"TestP2PModule_Start: the p2p module should start error free",
	)

	mm := m.(*p2pModule)

	assert.NotNil(
		t,
		mm.node,
		"TestP2PModule_Start: the p2p module should have a node",
	)

	assert.True(
		t,
		mm.node.IsRunning(),
		"TestP2PModule_Start: the p2p module should be running",
	)

	Teardown_TestP2PModule_Case(m)
}

func Teardown_TestP2PModule_Case(m modules.P2PModule) {
	m.Stop()
}

func Setup_TestP2PModule_Broadcast() (modules.P2PModule, []net.Listener, []net.Conn) {
	// generate the peer list
	peers := []string{}
	for i := 0; i < 27; i++ {
		peers = append(peers, fmt.Sprintf("%d@127.0.0.1:%d", i+1, i+1+10000))
	}

	m, err := Create(&config.Config{
		P2P: &config.P2PConfig{
			ExternalIp: "127.0.0.1:10001",
			BufferSize: 1024,
			Peers:      peers,
		},
	})

	if err != nil {
		log.Fatal("Fatal! could not instantiate the p2p module")
	}

	mm := m.(*p2pModule)
	node := mm.node.(*p2pNode)
	node.ID = 1

	// create 26 peers
	listeners := make([]net.Listener, 26)
	for i := 0; i < 26; i++ {
		l, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", 10000+i+2))
		if err != nil {
			log.Fatal("Setup_TestP2PNode_Broadcast() failed to listen for connections coming from the p2p node")
		}
		listeners[i] = l
	}

	// connect the p2p peer to the other 26
	for i := 0; i < 26; i++ {
		err := node.Dial(listeners[i].Addr().String())
		if err != nil {
			log.Fatal("Setup_TestP2PNode_Broadcast() p2p node failed to dial the mock peers")
		}
	}

	connections := make([]net.Conn, 26)
	for i := 0; i < 26; i++ {
		conn, err := listeners[i].Accept()
		if err != nil {
			log.Fatal("Setup_TestP2PNode_Broadcast() mock server(s) failed to accept the p2p node's connection")
		}
		connections[i] = conn
	}

	// modify the peer list to give them proper ids
	for k := range node.peers.m {
		peer := node.peers.m[k]
		if peer.Conn != nil {
			peerId := peer.Conn.RemoteAddr().(*net.TCPAddr).Port - 10000
			node.peers.m[k].ID = peerId
		}
	}

	return m, listeners, connections
}

func TestP2PModule_Broadcast(t *testing.T) {
	m, listeners, connections := Setup_TestP2PModule_Broadcast()

	ctrl := gomock.NewController(t)
	bus := mocks.NewMockBus(ctrl)
	m.SetBus(bus)

	mm := m.(*p2pModule)
	node := mm.node.(*p2pNode)

	err := m.Start()

	if err != nil {
		log.Fatal("TestP2PModule_Broadcast: the p2p module should start error free")
	}

	data, topic := &anypb.Any{}, shared.PocketTopic_CONSENSUS_MESSAGE_TOPIC
	msg := types.NewP2PMessage(0, 2, connections[5].LocalAddr().String(), node.config["address"].(string), &commonTypes.PocketEvent{
		Topic: topic,
		Data:  data,
	})
	msg.MarkAsBroadcastMessage()

	msgBytes, err := proto.Marshal(msg)
	if err != nil {
		log.Fatal("TestP2PModule_Broadcast: mock peer failed to marshal the broadcast message before send")
	}
	encodedMsg := node.encode(false, 0, msgBytes, true)

	bus.EXPECT().PublishEventToBus(msg.Payload).Times(1)

	_, err = connections[5].Write(encodedMsg) // these peers are already connected to the p2p node

	assert.Nil(
		t,
		err,
		"TestP2PModule_Broadcast: mock peer failed to write the broadcast message to the p2p node",
	)

	<-time.After(time.Millisecond * 1000)

	Teardown_TestP2PModule_Broadcast(m, listeners, connections)
}

func Teardown_TestP2PModule_Broadcast(m modules.P2PModule, listeners []net.Listener, connections []net.Conn) {
	// stop all connections
	for i := 0; i < 26; i++ {
		connections[i].Close()
	}

	// stop all listeners
	for i := 0; i < 26; i++ {
		listeners[i].Close()
	}

	// stop the module
	m.Stop()
}
