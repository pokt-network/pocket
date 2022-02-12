package p2p

import (
	"fmt"
	"net"
	"pocket/shared/messages"
	"strings"
	"sync"
	"time"
)

type work struct {
	nonce  uint32
	decode bool // should decode data using the domain codec or not
	data   []byte
}

type gater struct {
	GaterModule

	id           uint64
	protocol     string
	address      string
	externaladdr string

	c *dcodec

	inbound  iomap
	outbound iomap

	peerlist plist

	sink chan work

	listener  *net.TCPListener
	listening bool

	err    error
	done   chan uint
	ready  chan uint
	closed chan uint
}

type GaterModule interface {
	Config(protocol, address, external string, peers []string)
	Init() error

	Listen() error
	Ready() <-chan uint
	Close()
	Done() <-chan uint

	Send(addr string, msg []byte, wrapped bool) error

	BroadcastTempWrapper(m messages.NetworkMessage) error // TODO: hamza to refactor
	Broadcast(m message, isroot bool) error

	Handle()

	Request(addr string, msg []byte, wrapped bool) ([]byte, error)
	Respond(nonce uint32, iserroreof bool, addr string, msg []byte, wrapped bool) error

	Pong(msg message) error
	Ping(addr string) (bool, error)
}

func (g *gater) Config(protocol, address, external string, peers []string) {
	g.protocol = protocol
	g.address = address
	g.externaladdr = external
	g.peerlist = plist{}

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
	// TODO: revisit this
	cm := &churnmgmt{}

	churnmsg := cm.message(0, Ping, 0)

	_, err := g.c.register(churnmsg, cm.encode, cm.decode)
	if err != nil {
		return err
	}

	return nil
}

func (g *gater) Listen() error {
	defer func() {
		g.listener = nil
		g.listening = false
		close(g.closed)
	}()

	// add address validation and parsing
	listener, err := net.Listen(g.protocol, g.address)
	if err != nil {
		fmt.Println("Error:", err.Error())
	}

	g.listener = listener.(*net.TCPListener)
	g.listening = true

	close(g.ready)

	fmt.Println("Listening at", g.protocol, g.address, "...")
	for stop := false; !stop; {
		select {
		case <-g.done:
			stop = true
			break
		default:
			{
				conn, err := g.listener.Accept()
				if err != nil && g.listening {
					fmt.Println("Error receiving an inbound connection: ", err.Error())
					// TODO ignore use of closed network connection error when listener has closed
					g.err = err
					break // report error
				}

				if !g.listening {
					break
				}

				addr := conn.RemoteAddr().String()
				go g.handleInbound(conn, addr)
			}
		}
	}

	return nil
}

func (g *gater) Ready() <-chan uint {
	return g.ready
}

func (g *gater) Close() {
	g.done <- 1
	g.closed <- 1
	g.listening = false
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
	fmt.Println("Respond dialing", addr)
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

func (g *gater) BroadcastTempWrapper(msg *messages.NetworkMessage) error {
	m := message{
		payload: msg.Data,
		topic:   Topic(msg.Topic),
	}
	return g.Broadcast(m, false)

}

// Discuss: why is m not a pointer?
func (g *gater) Broadcast(m message, isroot bool) error {
	var toplevel int

	if isroot {
		maxlevel := getTopLevel(g.peerlist)
		toplevel = int(maxlevel)
	} else {
		fmt.Println("Not root, propagating down")
		toplevel = int(m.level)
	}

	source := g.externaladdr
	list := g.peerlist

	for currentlevel := toplevel - 1; currentlevel > 0; currentlevel-- {
		fmt.Println("New send")
		targetlist := getTargetList(list, g.id, toplevel, currentlevel)

		lpos := pickLeft(g.id, targetlist)
		rpos := pickRight(g.id, targetlist)

		left := targetlist[lpos]
		right := targetlist[rpos]

		m.level = uint16(currentlevel)

		var l_ack, r_ack []byte
		var l_err, r_err error = nil, nil
		var wg sync.WaitGroup

		wg.Add(1)
		go func(ack *[]byte, err *error, msg message) {

			msg.source = source
			msg.destination = left.address
			encm, _ := g.c.encode(msg)

			// just a hack
			if len(encm) != ReadBufferSize {
				encm = append(encm, make([]byte, ReadBufferSize-len(encm))...)
			}
			// just a hack

			fmt.Println("Requesting", left.address)
			response, reqerr := g.Request(left.address, encm, true)
			fmt.Println("Received resposne from", left.address)

			if reqerr != nil {
				*err = reqerr
			}

			*ack = response
			wg.Done()
		}(&l_ack, &l_err, m)

		wg.Add(1)
		go func(ack *[]byte, err *error, msg message) {

			msg.source = source
			msg.destination = right.address
			encm, _ := g.c.encode(msg)

			// just a hack
			if len(encm) != ReadBufferSize {
				encm = append(encm, make([]byte, ReadBufferSize-len(encm))...)
			}
			// just a hack

			fmt.Println("Requesting", right.address)
			response, reqerr := g.Request(right.address, encm, true)
			fmt.Println("Received resposne from", right.address)

			if reqerr != nil {
				*err = reqerr
			}
			*ack = response
			wg.Done()
		}(&r_ack, &r_err, m)

		wg.Wait()

		if l_err != nil {
			// pick next one but for send only (no ack)
			fmt.Println("Left failed to ack", l_err.Error())
		} else {
			fmt.Println("Left has acked", l_ack[:8])
		}

		if r_err != nil {
			// pick next one but for send only (no ack)
			fmt.Println("Right failed to ack", r_err.Error())
		} else {
			fmt.Println("Right has acked", l_ack[:8])
		}
	}

	// a hack to achieve full coverage like a redundancy layer
	sl := list.slice()
	for i := 0; i < len(sl); i++ {
		p := sl[i]
		if p.address != source {
			fmt.Println("redundancy", p, source)
			m.source = source
			m.destination = p.address
			m.level = 0

			encm, _ := g.c.encode(m)

			// just a hack
			if len(encm) != ReadBufferSize {
				encm = append(encm, make([]byte, ReadBufferSize-len(encm))...)
			}
			// just a hack

			fmt.Println("Requesting", p.address)
			reqerr := g.Send(p.address, encm, true)
			if reqerr != nil {
				fmt.Println(reqerr)
			}
		}
	}

	fmt.Println("Done broadcasting")

	return nil
}

func (g *gater) Handshake() {}
func (g *gater) Handle() {
	var msg message

	fmt.Println("Handling...")
	for w := range g.sink {
		nonce, data, decode := w.nonce, w.data, w.decode

		if decode {
			decoded, err := g.c.decode(data)
			if err != nil {
				fmt.Println("Error decoding data", err.Error())
				panic("D")
				//continue
			}

			msg = decoded.(message)
			msg.nonce = nonce
		} else {
			msg.payload = data
		}
		fmt.Println("msg:", msg)

		switch msg.action {

		case Gossip:
			fmt.Println("Received a gossip message")
			go func() {
				fmt.Println("Acking...")
				ack := (&gossip{}).message(msg.nonce, GossipACK, msg.level, g.externaladdr, msg.source)

				encoded, err := g.c.encode(ack)
				if err != nil {
					fmt.Println("Error encoding msg for gossipaCK", err.Error())
				}

				// just a hack
				encoded = append(encoded, make([]byte, ReadBufferSize-len(encoded))...)
				// just a hack

				err = g.Respond(msg.nonce, false, ack.destination, encoded, true)
				if err != nil {
					fmt.Println("Error encoding msg for gossipaCK", err.Error())
				}
			}()
			go g.Broadcast(msg, false)

		default:
			fmt.Println("Unrecognized message topic")
		}
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

		listener:  nil,
		listening: false,

		err:    nil,
		done:   make(chan uint, 1),
		ready:  make(chan uint, 1),
		closed: make(chan uint, 1),
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

		fmt.Println("New connection from", addr, err)
		if err != nil {
			pipe.close()
			<-pipe.closed
			pipe = nil
			g.err = err
		}

	}
}

func (g *gater) peerConnected(p *io) error {
	return nil
}

func (g *gater) peerDisconnected(p *io) error {
	return nil
}

func (g *gater) dial(addr string) (*io, error) {

	// TODO: this is equivalent to maxRetries = 1, add logic for > 1
	// TODO: should we explictly tell dial to use either inbound or outbound?
	exists := g.inbound.peak(addr)
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
		err = pipe.err
	case <-pipe.errored:
		err = pipe.err
	case <-pipe.ready:
		err = nil
	}

	if err != nil {
		fmt.Println("Error openning pipe", err.Error())
		pipe.close()
		<-pipe.closed
		pipe = nil
	}

	return pipe, err
}
