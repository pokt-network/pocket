package p2p

import (
	"log"
	"net"
	"pocket/p2p/types"
	"pocket/shared/config"
	"pocket/shared/modules"
	"reflect"
	"sync"

	"github.com/pokt-network/pocket/shared/config"
	pcrypto "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/types"
	"go.uber.org/atomic"
	"google.golang.org/protobuf/types/known/anypb"
)

type networkModule struct {
	bus    modules.Bus
	config *config.P2PConfig
	c      *dcodec

	id           uint64
	protocol     string
	address      string
	externaladdr string

	inbound  *Registry
	outbound *Registry

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

func newNetworkModule() *networkModule {
	return &networkModule{
		c: NewDomainCodec(),

		sink: make(chan types.Work, 100), // TODO(derrandz): rethink whether this should be buffered

		peerlist: nil,

		done:   make(chan uint, 1),
		ready:  make(chan uint, 1),
		closed: make(chan uint, 1),

		handlers: make(map[types.PeerEvent][]func(...interface{}), 0),
		errored:  make(chan uint, 1),
	}
}

var _ modules.NetworkModule = &networkModule{}

var networkLogger func(...interface{}) (int, error) = func(args ...interface{}) (int, error) {
	log.Println(args...)
	return 0, nil
}

func Create(config *config.Config) (modules.NetworkModule, error) {
	m := newNetworkModule()

	m.setLogger(networkLogger)

	m.config = config.P2P

	if err := m.validateConfig(); err != nil {
		return nil, err
	}

	m.configure(
		m.config.Protocol,
		m.config.Address,
		m.config.ExternalIp,
		m.config.Peers,
	)

	m.init()

	m.initConnectionPools()

	return m, nil
}

func (m *networkModule) validateConfig() error {
	requiredCfgEntries := []string{
		"MaxInbound",
		"MaxOutbound",
		"WireByteHeaderLength",
		"ReadBufferSize",
		"WriteBufferSize",
		"ReadDeadlineMs",
		"ReadBufferSize",
		"ReadDeadlineMs",
	}

	cfg := reflect.ValueOf(m.config).Elem()

	for _, name := range requiredCfgEntries {
		field := cfg.FieldByName(name)
		if field == (reflect.Value{}) {
			return ErrMissingConfigField(name)
		}
	}

	return nil
}

func (m *networkModule) Start() error {
	m.log("Starting the P2P Module.")
	go m.listen()

	<-m.isReady()

	return nil
}

func (m *networkModule) Stop() error {
	go m.close()

	<-m.closed
	<-m.done
	<-m.errored

	if m.err.error != nil {
		return m.err.error
	}

	return nil
}

func (m *networkModule) SetBus(pocketBus modules.Bus) {
	m.bus = pocketBus
}

func (m *networkModule) GetBus() modules.Bus {
	if m.bus == nil {
		log.Fatalf("PocketBus is not initialized")
	}
	return m.bus
}

func (m *networkModule) BroadcastMessage(msg *anypb.Any, topic string) error {
	netmsg := &types.NetworkMessage{Data: msg, Topic: types.Topic(topic)}
	return m.broadcast(netmsg, true)
}

func (m *p2pModule) Send(addr pcrypto.Address, msg *anypb.Any, topic types.PocketTopic) error {
	panic("Send not implemented")
	netmsg := &types.NetworkMessage{Data: msg, Topic: types.Topic(topic)}
	encodedBytes, err := m.c.encode(netmsg)
	if err != nil {
		return err
	}

	return m.send(addr, encodedBytes, true) // true: meaning that this message is already encoded
}

func (m *networkModule) AckSend(addr string, msg *types.NetworkMessage) (bool, error) {
	encodedBytes, err := m.c.encode(msg)
	if err != nil {
		return false, err
	}

	response, err := m.request(addr, encodedBytes, true) // true: meaning that this message is already encoded
	if err != nil {
		return false, err
	}

	ack, err := m.c.decode(response)
	if err != nil {
		return true, err // TODO(derrandz): notice it's true
	}

	ackmsg := ack.(*types.NetworkMessage)

	if ackmsg.Nonce == msg.Nonce {
		return true, nil
	}

	return false, nil
}
