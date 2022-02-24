package p2p

import (
	"errors"
	"log"
	"net"
	"pocket/p2p/types"
	"pocket/shared/config"
	"pocket/shared/modules"
	"pocket/shared/types"
	"sync"

	"go.uber.org/atomic"
	"google.golang.org/protobuf/types/known/anypb"
)

var (
	MaxInbound           uint = 128
	MaxOutbound          uint = 128
	WireByteHeaderLength int  = 9
	ReadBufferSize       int  = (1024 * 4)
	WriteBufferSize      int  = (1024 * 4)
	ReadDeadlineMs       int  = 400
)

var (
	Protocol = "tcp"
	Address  = "localhost:30303"
)

type P2PModule struct {
	modules.NetworkModule

	bus    modules.Bus
	config *config.P2PConfig
	c      *dcodec

	id           uint64
	protocol     string
	address      string
	externaladdr string

	inbound  pipemap
	outbound pipemap

	peerlist *types.Peerlist

	listener struct {
		sync.Mutex
		*net.TCPListener
	}
	listening atomic.Bool

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

var (
	ErrNotCreated = errors.New("Module error: P2P Module not created. Trying to start the p2p module before calling create.")
)

func NewP2PModule() *P2PModule {
	return &P2PModule{
		protocol: Protocol,
		address:  Address,

		inbound:  *NewIoMap(MaxInbound),
		outbound: *NewIoMap(MaxOutbound),

		c: NewDomainCodec(),

		peerlist: nil,
		sink:     make(chan types.Work, 100), // TODO: rethink whether this should be buffered

		done:   make(chan uint, 1),
		ready:  make(chan uint, 1),
		closed: make(chan uint, 1),

		handlers: make(map[types.PeerEvent][]func(...interface{}), 0),
		errored:  make(chan uint, 1),
	}
}

func Create(config *config.Config) (modules.NetworkModule, error) {
	module := NewP2PModule()

	module.setLogger(func(args ...interface{}) (int, error) {
		log.Println(args...)
		return 0, nil
	})

	module.config = config.P2P

	return module, nil
}

func (m *P2PModule) Start() error {
	m.log("p.list", m.peerlist)

	m.configure(
		m.config.Protocol,
		m.config.Address,
		m.config.ExternalIp,
		m.config.Peers,
	)
	m.init()

	go m.listen()

	<-m.isReady()

	return nil
}

func (m *P2PModule) Stop() error {
	go m.close()

	<-m.closed
	<-m.done
	<-m.errored

	if m.err.error != nil {
		return m.err.error
	}

	return nil
}

func (m *P2PModule) SetBus(pocketBus modules.Bus) {
	m.bus = pocketBus
}

func (m *P2PModule) GetBus() modules.Bus {
	if m.bus == nil {
		log.Fatalf("PocketBus is not initialized")
	}
	return m.bus
}

func (m *P2PModule) BroadcastMessage(msg *anypb.Any, topic string) error {
	netmsg := &types.NetworkMessage{Data: msg, Topic: types.Topic(topic)}
	return m.broadcast(netmsg, true)
}

func (m *P2PModule) Send(addr string, msg *anypb.Any, topic string) error {
	netmsg := &types.NetworkMessage{Data: msg, Topic: types.Topic(topic)}
	encodedBytes, err := m.c.encode(netmsg)
	if err != nil {
		return err
	}

	return m.send(addr, encodedBytes, true) // true: meaning that this message is already encoded
}

func (m *P2PModule) AckSend(addr string, msg *types.NetworkMessage) (bool, error) {
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
		return true, err // TODO: notice it's true
	}

	ackmsg := ack.(*types.NetworkMessage)

	if ackmsg.Nonce == msg.Nonce {
		return true, nil
	}

	return false, nil
}
