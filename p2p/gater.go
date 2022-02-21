package p2p

import (
	"fmt"
	"net"
	"pocket/p2p/pre_p2p/types"
	"strings"
	"sync"
	"time"

	"google.golang.org/protobuf/types/known/anypb"

	"go.uber.org/atomic"
	"google.golang.org/protobuf/types/known/anypb"
)

type work struct {
	nonce  uint32
	decode bool // should decode data using the domain codec or not
	data   []byte
	addr   string // temporary until we start using ids
}

type GaterModule interface {
	Config(protocol, address, external string, peers []string)
	Init() error

	Listen() error
	Ready() <-chan uint
	Close()
	Done() <-chan uint

	Send(addr string, msg []byte, wrapped bool) error
	Broadcast(m *types.NetworkMessage, isroot bool) error

	On(GaterEvent, func(...interface{}))
	Handle()

	Request(addr string, msg []byte, wrapped bool) ([]byte, error)
	Respond(nonce uint32, iserroreof bool, addr string, msg []byte, wrapped bool) error

	Pong(msg message) error
	Ping(addr string) (bool, error)

	Log(m ...interface{})
	SetLogger(func(m ...interface{}))
}

type GaterEvent int

const (
	BroadcastDoneEvent GaterEvent = iota
	PeerConnectedEvent
	PeerDisconnectedEvent
)

type gater struct {
	GaterModule

	id           uint64
	protocol     string
	address      string
	externaladdr string

	inbound  iomap
	outbound iomap

	c *dcodec

	peerlist *plist

	sink chan work

	listener struct {
		sync.Mutex
		*net.TCPListener
	}
	listening atomic.Bool

	done   chan uint
	ready  chan uint
	closed chan uint

	handlers map[GaterEvent][]func(...interface{})

	logger struct {
		sync.RWMutex
		print func(...interface{}) (int, error)
	}

	err struct {
		sync.Mutex
		error
	}
	errored chan uint
}

func (g *gater) Config(protocol, address, external string, peers []string) {
	g.protocol = protocol
	g.address = address
	g.externaladdr = external
	g.peerlist = &plist{elements: make([]peer, 0)}

	// this is a hack to get going no more no less
	for i, p := range peers {
		pr := peer{id: uint64(i + 1), address: p}
		port := strings.Split(pr.address, ":")
		myport := strings.Split(g.address, ":")
		if port[1] == myport[1] {
			g.id = pr.id
		}
		g.peerlist.add(pr)
	}
}

func (g *gater) Init() error {
	pbuffmsnger := &pbuff{}
	msg := pbuffmsnger.message(int32(0), 1, types.PocketTopic_P2P, "", "")
	_, err := g.c.register(*msg, pbuffmsnger.encode, pbuffmsnger.decode)
	if err != nil {
		return err
	}

	return nil
}

func (g *gater) Listen() error {
	defer func() {
		g.listening.Store(false)
		close(g.closed)
	}()

	// add address validation and parsing
	listener, err := net.Listen(g.protocol, g.address)
	if err != nil {
		g.Log("Error:", err.Error())
		return err
	}

	g.listener.Lock()
	g.listener.TCPListener = listener.(*net.TCPListener)
	g.listener.Unlock()
	g.Log("prehere")
	g.listening.Store(true)

	close(g.ready)

	g.Log("here?")
	g.Log("Listening at", g.protocol, g.address, "...")
	for stop := false; !stop; {
		select {
		case <-g.done:
			stop = true
			break
		default:
			{
				conn, err := g.listener.Accept()
				if err != nil && g.listening.Load() {
					g.Log("Error receiving an inbound connection: ", err.Error())
					// TODO ignore use of closed network connection error when listener has closed
					g.error(err)
					break // report error
				}

				if !g.listening.Load() {
					break
				}

				addr := conn.RemoteAddr().String()
				go g.handleInbound(conn, addr)
			}
		}
	}

	g.listener.Lock()
	g.listener.TCPListener = nil
	g.listener.Unlock()

	return nil
}

func (g *gater) Ready() <-chan uint {
	return g.ready
}

func (g *gater) Close() {
	g.done <- 1
	g.closed <- 1
	g.listening.Store(false)
	g.listener.Close()
	close(g.done)
}

func (g *gater) Done() <-chan uint {
	return g.closed
}

func (g *gater) Send(addr string, msg []byte, wrapped bool) error {
	pipe, derr := g.dial(addr)
	if derr != nil {
		return derr
	}

	_, werr := pipe.write(msg, false, 0, wrapped)
	if werr != nil {
		return werr
	}

	if pipe.err.error != nil {
		return pipe.err.error
	}

	return nil
}

// TODO: this: msg []byte, wrapped bool) is repeat everything, maybe a struct for this?
func (g *gater) Request(addr string, msg []byte, wrapped bool) ([]byte, error) {
	pipe, derr := g.dial(addr)
	if derr != nil {
		return nil, derr
	}

	response, rerr := pipe.ackwrite(msg, wrapped)
	if rerr != nil {
		return nil, rerr
	}

	return response.data, nil
}

func (g *gater) Respond(nonce uint32, iserroreof bool, addr string, msg []byte, wrapped bool) error {
	pipe, derr := g.dial(addr)
	if derr != nil {
		return derr
	}

	_, werr := pipe.write(msg, iserroreof, nonce, wrapped)
	if werr != nil {
		return werr
	}

	return nil
}

func (g *gater) Ping(addr string) (bool, error) {
	var pongbytes []byte
	var msngr churnmgmt

	pingmsg := msngr.message(0, Ping, 0, g.address, addr)
	pingbytes, err := g.c.encode(pingmsg)

	if err != nil {
		return false, err
	}

	timedout := make(chan int)
	ponged := make(chan int)
	errored := make(chan error)

	go func() {
		<-time.After(time.Millisecond * 500)
		timedout <- 1
	}()

	go func() {
		response, err := g.Request(addr, pingbytes, true)

		if err != nil {
			errored <- err
		}

		pongbytes = response
		ponged <- 1
	}()

	select {

	case <-timedout:
		return false, nil

	case err := <-errored:
		return false, err

	case <-ponged:
		pong, err := g.c.decode(pongbytes)
		pongmsg := pong.(message)

		if err != nil {
			return false, err
		}

		var valid bool
		if strings.Compare(string(pongmsg.action), string(Pong)) != 0 {
			valid = true
		}

		return valid, nil
	}
} // TODO: should we use UDP requests for ping?

// TODO: test
func (g *gater) Pong(msg message) error {
	if msg.IsRequest() && msg.action == Ping {
		messaging := &churnmgmt{}
		pongmsg := messaging.message(msg.nonce, Pong, 0, msg.destination, msg.source)
		pongbytes, err := g.c.encode(pongmsg)

		if err != nil {
			return err
		}

		err = g.Respond(msg.nonce, false, msg.source, pongbytes, true)

		if err != nil {
			return err
		}
	}
	return nil
}

// Discuss: why is m not a pointer?
func (g *gater) Broadcast(m *types.NetworkMessage, isroot bool) error {
	g.CondLog(isroot, "Starting gossip round, level is", m.Level)

	// var mmutex sync.Mutex

	var list *plist = g.peerlist
	var sourceAddr string = g.externaladdr

	var toplevel int = int(getTopLevel(list))
	var currentlevel int = toplevel - 1

	if !isroot {
		g.CondLog(isroot, "Received gossip message from level", m.Level)
		currentlevel = int(m.Level)
	}

	SEND_AND_WAIT_FOR_ACK := func(encryptedmsg []byte, wg *sync.WaitGroup, target *peer, ack *[]byte, err *error) {
		response, reqerr := g.Request(target.address, encryptedmsg, true)

		if reqerr != nil {
			*err = reqerr
			wg.Done()
			return
		}

		*ack = response

		wg.Done()
		return
	}

	for ; currentlevel > 0; currentlevel-- {
		m.Level = int32(currentlevel)
		g.CondLog(isroot, "Gossiping to peers at level", currentlevel, "message level=", m.Level)
		targetlist := getTargetList(list, g.id, toplevel, currentlevel)

		var left, right *peer
		{
			lpos := pickLeft(g.id, targetlist)
			rpos := pickRight(g.id, targetlist)

			left = targetlist.get(lpos)
			right = targetlist.get(rpos)
		}

		var wg sync.WaitGroup

		var l_ack, r_ack []byte
		var l_err, r_err error = nil, nil
		var l_msg, r_msg types.NetworkMessage
		var l_data, r_data []byte

		{
			l_msg = *m
			l_msg.Source = sourceAddr
			l_msg.Destination = left.address
			encm, errenc := g.c.encode(l_msg)
			if errenc != nil {
				return errenc
			}
			l_data = encm
		}

		{
			r_msg = *m
			r_msg.Source = sourceAddr
			r_msg.Destination = right.address
			encm, errenc := g.c.encode(r_msg)
			if errenc != nil {
				return errenc
			}
			r_data = encm
		}

		wg.Add(1)
		go SEND_AND_WAIT_FOR_ACK(l_data, &wg, left, &l_ack, &l_err)

		wg.Add(1)
		go SEND_AND_WAIT_FOR_ACK(r_data, &wg, right, &r_ack, &r_err)

		wg.Wait()
		g.CondLog(isroot, "Got acks from left and right")

		if l_err != nil {
			return l_err
		}
		g.CondLog(isroot, "(raintree) left peer: ACK")

		if r_err != nil {
			return r_err
		}
		g.CondLog(isroot, "(raintree) right peer: ACK")
	}

	g.CondLog(isroot, "Done broadcasting")

	for _, handler := range g.handlers[BroadcastDoneEvent] {
		handler(m)
	}
	return nil
}

func (g *gater) Handshake() {}

func (g *gater) On(e GaterEvent, handler func(...interface{})) {
	if g.handlers != nil {
		if hmap, exists := g.handlers[e]; exists {
			hmap = append(hmap, handler)
		} else {
			g.handlers[e] = append(make([]func(...interface{}), 0), handler)
		}
	}
}

func (g *gater) Handle() {
	var msg *types.NetworkMessage
	var m sync.Mutex

	g.Log("Handling...")
	for w := range g.sink {
		nonce, data, decode, srcaddr := w.nonce, w.data, w.decode, w.addr

		if decode {
			decoded, err := g.c.decode(data)
			if err != nil {
				g.Log("Error decoding data", err.Error())
				panic("D")
				//continue
			}

			m.Lock()
			msgi := decoded.(types.NetworkMessage)
			msg = &msgi
			msg.Nonce = int32(nonce)
			m.Unlock()
		} else {
			msg.Data = &anypb.Any{}
			msg.Nonce = int32(nonce)
			msg.Source = srcaddr
		}

		switch msg.Topic {

		case types.PocketTopic_CONSENSUS:
			m.Lock()
			ack := &types.NetworkMessage{Nonce: msg.Nonce, Level: msg.Level, Topic: types.PocketTopic_CONSENSUS, Source: g.externaladdr, Destination: msg.Source}
			encoded, err := g.c.encode(*ack)
			if err != nil {
				g.Log("Error encoding m for gossipaCK", err.Error())
			}

			err = g.Respond(uint32(msg.Nonce), false, srcaddr, encoded, true)
			m.Unlock()
			if err != nil {
				g.Log("Error encoding msg for gossipaCK", err.Error())
			}
			<-time.After(time.Millisecond * 5)
			g.Log("Acked to", ack.Destination)
			go g.Broadcast(&types.NetworkMessage{
				Nonce:       msg.Nonce,
				Level:       msg.Level,
				Source:      msg.Source,
				Topic:       msg.Topic,
				Destination: msg.Destination,
				Data:        msg.Data,
			}, false)

		default:
			g.Log("Unrecognized message topic", msg.Topic, "from", msg.Source, "to", msg.Destination, "@node", g.address)
		}
	}
}

func (g *gater) SetLogger(logger func(...interface{}) (int, error)) {
	defer g.logger.Unlock()
	g.logger.Lock()

	g.logger.print = logger
}

func (g *gater) Log(m ...interface{}) {
	defer g.logger.Unlock()
	g.logger.Lock()

	if g.logger.print != nil {
		args := make([]interface{}, 0)
		args = append(args, fmt.Sprintf("[%s]", g.address))
		args = append(args, m...)
		g.logger.print(args...)
	}
}

func (g *gater) CondLog(cond bool, m ...interface{}) {
	if cond {
		g.Log(m)
	}
}

func NewGater() *gater {
	return &gater{
		protocol: Protocol,
		address:  Address,

		inbound:  *NewIoMap(MaxInbound),
		outbound: *NewIoMap(MaxOutbound),

		c: NewDomainCodec(),

		peerlist: nil,
		sink:     make(chan work, 100), // TODO: rethink whether this should be buffered

		done:   make(chan uint, 1),
		ready:  make(chan uint, 1),
		closed: make(chan uint, 1),

		handlers: make(map[GaterEvent][]func(...interface{}), 0),
		errored:  make(chan uint, 1),
	}
}

/*
 @
 @ Internal
 @
*/
func (g *gater) handleInbound(conn net.Conn, addr string) {
	pipe, exists := g.inbound.get(addr)
	if !exists {
		pipe.g = g
		go pipe.open(InboundIoPipe, addr, conn, g.peerConnected, g.peerDisconnected)

		var err error
		select {
		case <-pipe.done:
			err = pipe.err
		case <-pipe.errored:
			err = pipe.err
		case <-pipe.ready:
			err = nil
		}

		g.Log("New connection from", addr, err)
		if err != nil {
			pipe.close()
			<-pipe.closed
			pipe = nil
			g.error(err)
		}

	}
}

func (g *gater) peerConnected(p *io) error {
	g.Log("Peer connected", p.addr)
	return nil
}

func (g *gater) peerDisconnected(p *io) error {
	return nil
}

func (g *gater) dial(addr string) (*io, error) {

	// TODO: this is equivalent to maxRetries = 1, add logic for > 1
	// TODO: should we explictly tell dial to use either inbound or outbound?
	exists := g.inbound.peak(addr)
	g.Log("Peaked into inbound clients map for", addr, "found=", exists)
	if exists {
		pipe, _ := g.inbound.get(addr)
		return pipe, nil
	}

	pipe, exists := g.outbound.get(addr)
	if exists {
		return pipe, nil
	}

	pipe.g = g
	go pipe.open(OutboundIoPipe, addr, nil, g.peerConnected, g.peerDisconnected)

	var err error
	select {
	case <-pipe.done:
		err = pipe.err.error
	case <-pipe.errored:
		err = pipe.err.error
	case <-pipe.ready:
		err = nil
	}

	if err != nil {
		g.Log("Error openning pipe", err.Error())
		pipe.close()
		<-pipe.closed
		pipe = nil
	}

	return pipe, err
}

func (g *gater) error(err error) {
	defer g.err.Unlock()
	g.err.Lock()

	if g.err.error != nil {
		g.err.error = err
	}

	g.errored <- 1
}
