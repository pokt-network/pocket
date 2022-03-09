package p2p

import (
	"log"
	"net"
	"sync"

	"github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/shared/config"
	pcrypto "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
	shared "github.com/pokt-network/pocket/shared/types"
	"go.uber.org/atomic"
	"google.golang.org/protobuf/types/known/anypb"
)

type p2pModule struct {
	bus    modules.Bus
	config *config.P2PConfig
	c      types.Marshaler

	id           uint64
	protocol     string
	address      string
	externaladdr string

	inbound  *types.Registry
	outbound *types.Registry

	peerlist *types.Peerlist

	listener struct {
		sync.Mutex
		*net.TCPListener
	}
	isListening atomic.Bool

	done    chan uint
	ready   chan uint
	closed  chan uint
	errored chan uint

	sink     chan types.Work
	handlers map[types.PeerEvent][]func(...interface{})

	logger struct {
		sync.RWMutex
		print func(...interface{}) (int, error)
	}

	err struct {
		sync.Mutex
		error
	}
}

var _ modules.P2PModule = &p2pModule{}
var _ types.Runner = &p2pModule{}

var networkLogger func(...interface{}) (int, error) = func(args ...interface{}) (int, error) {
	log.Println(args...)
	return 0, nil
}

func Create(config *config.Config) (modules.P2PModule, error) {
	m := newP2PModule()

	m.setLogger(networkLogger)

	if err := m.initialize(config.P2P); err != nil { // TODO(derrandz): Should initialize also include logger intialization?
		return nil, err
	}

	return m, nil
}

func (m *p2pModule) Start() error {
	m.log("Starting the P2P Module.")
	go m.listen()

	<-m.isReady()

	return nil
}

func (m *p2pModule) Stop() error {
	go m.close()

	<-m.closed
	<-m.done
	<-m.errored

	if m.err.error != nil {
		return m.err.error
	}

	return nil
}

func (m *p2pModule) SetBus(pocketBus modules.Bus) {
	m.bus = pocketBus
}

func (m *p2pModule) GetBus() modules.Bus {
	if m.bus == nil {
		log.Fatalf("PocketBus is not initialized")
	}
	return m.bus
}

func (m *p2pModule) BroadcastMessage(data *anypb.Any, topic shared.PocketTopic) error {
	return m.Broadcast(data, topic)
}

func (m *p2pModule) Broadcast(data *anypb.Any, topic shared.PocketTopic) error {
	metadata := &types.Metadata{}
	payload := &shared.PocketEvent{Data: data, Topic: topic}
	p2pmsg := &types.P2PMessage{Metadata: metadata, Payload: payload}

	return m.broadcast(p2pmsg, true)
}

func (m *p2pModule) Send(addr pcrypto.Address, data *anypb.Any, topic shared.PocketTopic) error {
	metadata := &types.Metadata{}
	payload := &shared.PocketEvent{Data: data, Topic: topic}
	p2pmsg := &types.P2PMessage{Metadata: metadata, Payload: payload}
	encodedBytes, err := m.c.Marshal(p2pmsg)
	if err != nil {
		return err
	}

	// TODO(derrandz): look into using pcrypto.Address
	return m.send("", encodedBytes, true) // true: meaning that this message is already encoded
}

func (m *p2pModule) AckSend(addr string, msg *types.P2PMessage) (bool, error) {
	encodedBytes, err := m.c.Marshal(msg)
	if err != nil {
		return false, err
	}

	response, err := m.request(addr, encodedBytes, true) // true: meaning that this message is already encoded
	if err != nil {
		return false, err
	}

	var ack *types.P2PAckMessage
	err = m.c.Unmarshal(response, ack)
	if err != nil {
		return true, err // TODO(derrandz): notice it's true
	}

	if ack.Ackee == msg.Metadata.Source {
		return true, nil
	}

	return false, nil
}
