package p2p

import (
	"net"
	"sync"

	"github.com/pokt-network/pocket/p2p/types"
	shared "github.com/pokt-network/pocket/shared/types"

	"google.golang.org/protobuf/types/known/anypb"
)

func (m *p2pModule) handshake() {}

func (m *p2pModule) dial(addr string) (*socket, error) {
	// TODO(derrandz): this is equivalent to maxRetries = 1, add logic for > 1
	// TODO(derrandz): should we explictly tell dial to use either inbound or outbound?
	exists := m.inbound.Peak(addr)
	m.log("Peaked into inbound clients map for", addr, "found=", exists)
	if exists {
		obj, _ := m.inbound.Get(addr)
		pipe := obj.(*socket)
		return pipe, nil
	}

	obj, exists := m.outbound.Get(addr)
	pipe := obj.(*socket)
	if exists {
		return pipe, nil
	}

	pipe.network = m
	go pipe.open(OutboundIoPipe, addr, nil, m.peerConnected, m.peerDisconnected)

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
		m.log("Error openning pipe", err.Error())
		pipe.close()
		<-pipe.closed
		pipe = nil
	}

	return pipe, err
}

func (m *p2pModule) send(addr string, msg []byte, wrapped bool) error {
	pipe, derr := m.dial(addr)
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

func (m *p2pModule) listen() error {
	defer func() {
		m.isListening.Store(false)
		close(m.closed)
	}()

	// add address validation and parsing
	listener, err := net.Listen(m.protocol, m.address)
	if err != nil {
		m.log("Error:", err.Error())
		return err
	}

	m.listener.Lock()
	m.listener.TCPListener = listener.(*net.TCPListener)
	m.listener.Unlock()
	m.isListening.Store(true)

	close(m.ready)

	m.log("Listening at", m.protocol, m.address, "...")
	for stop := false; !stop; {
		select {
		case <-m.done:
			stop = true
			break
		default:
			{
				conn, err := m.listener.Accept()
				if err != nil && m.isListening.Load() {
					m.log("Error receiving an inbound connection: ", err.Error())
					// TODO(derrandz) ignore use of closed network connection error when listener has closed
					m.error(err)
					break // report error
				}

				if !m.isListening.Load() {
					break
				}

				addr := conn.RemoteAddr().String()
				go m.handleInbound(conn, addr)
			}
		}
	}

	m.listener.Lock()
	m.listener.TCPListener = nil
	m.listener.Unlock()

	return nil
}

// TODO(derrandz): this: msg []byte, wrapped bool) is repeat everything, maybe a struct for this?
func (m *p2pModule) request(addr string, msg []byte, wrapped bool) ([]byte, error) {
	pipe, derr := m.dial(addr)
	if derr != nil {
		return nil, derr
	}
	var response types.Work

	response, rerr := pipe.ackwrite(msg, wrapped)
	if rerr != nil {
		return nil, rerr
	}

	return response.Bytes(), nil
}

func (m *p2pModule) respond(nonce uint32, iserroreof bool, addr string, msg []byte, wrapped bool) error {
	pipe, derr := m.dial(addr)
	if derr != nil {
		return derr
	}

	_, werr := pipe.write(msg, iserroreof, nonce, wrapped)
	if werr != nil {
		return werr
	}

	return nil
}

func (m *p2pModule) broadcast(msg *types.P2PMessage, isRoot bool) error {

	// var mmutex sync.Mutex

	var list *types.Peerlist = m.peerlist
	var sourceAddr string = m.externaladdr

	var topLevel int = int(getTopLevel(list))
	var currentLevel int = topLevel - 1

	if msg.Metadata.Level == int32(topLevel) && msg.Metadata.Source == "" { // TODO(Derrandz): m.config.Address
		isRoot = true
	}

	m.clog(isRoot, "Starting gossip round, level is", msg.Metadata.Level)

	if !isRoot {
		m.clog(isRoot, "Received gossip message from level", msg.Metadata.Level)
		currentLevel = int(msg.Metadata.Level)
	}

	SEND_AND_WAIT_FOR_ACK := func(encodedMsg []byte, wg *sync.WaitGroup, target *types.Peer, ack *[]byte, err *error) {
		response, reqerr := m.request(target.Addr(), encodedMsg, true)

		if reqerr != nil {
			*err = reqerr
			wg.Done()
			return
		}

		*ack = response

		wg.Done()
		return
	}

	for ; currentLevel > 0; currentLevel-- {
		msg.Metadata.Level = int32(currentLevel)
		m.clog(isRoot, "Gossiping to peers at level", currentLevel, "message level=", msg.Metadata.Level)
		targetlist := getTargetList(list, m.id, topLevel, currentLevel)

		var left, right *types.Peer
		{
			lpos := pickLeft(m.id, targetlist)
			rpos := pickRight(m.id, targetlist)

			left = targetlist.Get(lpos)
			right = targetlist.Get(rpos)
		}

		var wg sync.WaitGroup

		var l_ack, r_ack []byte
		var l_err, r_err error = nil, nil
		var l_msg, r_msg types.P2PMessage
		var l_data, r_data []byte

		{
			l_msg = *msg
			l_msg.Metadata.Source = sourceAddr
			l_msg.Metadata.Destination = left.Addr()
			encm, errenc := m.c.Encode(l_msg)
			if errenc != nil {
				return errenc
			}
			l_data = encm
		}

		{
			r_msg = *msg
			r_msg.Metadata.Source = sourceAddr
			r_msg.Metadata.Destination = right.Addr()
			encm, errenc := m.c.Encode(r_msg)
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
		m.clog(isRoot, "Got acks from left and right")

		if l_err != nil {
			return l_err
		}
		m.clog(isRoot, "(raintree) left peer: ACK")

		if r_err != nil {
			return r_err
		}
		m.clog(isRoot, "(raintree) right peer: ACK")
	}

	m.clog(isRoot, "Done broadcasting")

	for _, handler := range m.handlers[types.BroadcastDoneEvent] {
		handler(m)
	}
	return nil
}

func (m *p2pModule) handle() {
	var msg *types.P2PMessage
	var mx sync.Mutex

	m.log("Handling...")
	for w := range m.sink {
		nonce, data, srcaddr, encoded := (&w).Implode()

		if encoded {
			decoded, err := m.c.Decode(data)
			if err != nil {
				m.log("Error decoding data", err.Error())
				continue
			}

			mx.Lock()
			msgi := decoded.(types.P2PMessage)
			msg = &msgi
			msg.Metadata.Nonce = int32(nonce)
			mx.Unlock()
		} else {
			msg.Payload.Data = &anypb.Any{}
			msg.Metadata.Nonce = int32(nonce)
			msg.Metadata.Source = srcaddr
		}

		switch msg.Payload.Topic {

		case shared.PocketTopic_CONSENSUS_MESSAGE_TOPIC:
			mx.Lock()
			md := &types.Metadata{
				Nonce:       msg.Metadata.Nonce,
				Level:       msg.Metadata.Level,
				Source:      m.externaladdr,
				Destination: msg.Metadata.Source,
			}
			pl := &shared.PocketEvent{
				Topic: shared.PocketTopic_CONSENSUS_MESSAGE_TOPIC,
			}
			ack := &types.P2PMessage{Metadata: md, Payload: pl}
			encoded, err := m.c.Encode(*ack)
			if err != nil {
				m.log("Error encoding m for gossipaCK", err.Error())
			}

			err = m.respond(uint32(msg.Metadata.Nonce), false, srcaddr, encoded, true)
			if err != nil {
				m.log("Error encoding msg for gossipaCK", err.Error())
			}

			mx.Unlock()

			m.log("Acked to", ack.Metadata.Destination)

			go m.broadcast(msg, false)

		default:
			m.log("Unrecognized message topic", msg.Payload.Topic, "from", msg.Metadata.Source, "to", msg.Metadata.Destination, "@node", m.address)
		}
	}
}

func (m *p2pModule) handleInbound(conn net.Conn, addr string) {
	var pipe *socket
	obj, exists := m.inbound.Get(addr)
	pipe = obj.(*socket)
	if !exists {
		pipe.network = m
		go pipe.open(InboundIoPipe, addr, conn, m.peerConnected, m.peerDisconnected)

		var err error
		select {
		case <-pipe.done:
			err = pipe.err
		case <-pipe.errored:
			err = pipe.err
		case <-pipe.ready:
			err = nil
		}

		m.log("New connection from", addr, err)
		if err != nil {
			pipe.close()
			<-pipe.closed
			pipe = nil
			m.error(err)
		}

	}
}

func (m *p2pModule) on(e types.PeerEvent, handler func(...interface{})) {
	if m.handlers != nil {
		if hmap, exists := m.handlers[e]; exists {
			hmap = append(hmap, handler)
		} else {
			m.handlers[e] = append(make([]func(...interface{}), 0), handler)
		}
	}
}

func (m *p2pModule) peerConnected(p *socket) error {
	m.log("Peer connected", p.addr)
	return nil
}

func (m *p2pModule) peerDisconnected(p *socket) error {
	return nil
}

func (m *p2pModule) initializePools() {
	socketFactory := func() interface{} {
		sck := NewSocket(m.config.BufferSize, m.config.WireHeaderLength, m.config.TimeoutInMs)
		return interface{}(sck) // TODO(derrandz): remember to change this if you end up using unsafe_ptr instead of interface{}
	}

	m.inbound = types.NewRegistry(m.config.MaxInbound, socketFactory)
	m.outbound = types.NewRegistry(m.config.MaxOutbound, socketFactory)
}

func newP2PModule() *p2pModule {
	return &p2pModule{
		c: NewTypesCodec(),

		sink: make(chan types.Work, 100), // TODO(derrandz): rethink whether this should be buffered

		peerlist: nil,

		done:   make(chan uint, 1),
		ready:  make(chan uint, 1),
		closed: make(chan uint, 1),

		handlers: make(map[types.PeerEvent][]func(...interface{}), 0),
		errored:  make(chan uint, 1),
	}
}
