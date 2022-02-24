package p2p

import (
	"net"
	"pocket/p2p/types"
	"sync"
	"time"

	"google.golang.org/protobuf/types/known/anypb"
)

func (m *networkModule) handshake() {}

func (m *networkModule) dial(addr string) (*socket, error) {
	// TODO(derrandz): this is equivalent to maxRetries = 1, add logic for > 1
	// TODO(derrandz): should we explictly tell dial to use either inbound or outbound?
	exists := m.inbound.peak(addr)
	m.log("Peaked into inbound clients map for", addr, "found=", exists)
	if exists {
		pipe, _ := m.inbound.get(addr)
		return pipe, nil
	}

	pipe, exists := m.outbound.get(addr)
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

func (m *networkModule) send(addr string, msg []byte, wrapped bool) error {
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

func (m *networkModule) listen() error {
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
func (m *networkModule) request(addr string, msg []byte, wrapped bool) ([]byte, error) {
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

func (m *networkModule) respond(nonce uint32, iserroreof bool, addr string, msg []byte, wrapped bool) error {
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

func (m *networkModule) ping(addr string) (bool, error) {
	// TODO(derrandz): refactor this to use types.NetworkMessage
	var pongbytes []byte

	pingmsg := types.NetworkMessage{Topic: types.PocketTopic_P2P_PING, Source: m.address, Destination: addr}
	pingbytes, err := m.c.encode(pingmsg)

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
		response, err := m.request(addr, pingbytes, true)

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
		pong, err := m.c.decode(pongbytes)
		pongmsg := pong.(types.NetworkMessage)

		if err != nil {
			return false, err
		}

		var valid bool
		if pongmsg.Topic != types.PocketTopic_P2P_PONG {
			valid = true
		}

		return valid, nil
	}
} // TODO(derrandz): should we use UDP requests for ping?

// TODO(derrandz): test
func (m *networkModule) pong(msg types.NetworkMessage) error {
	// TODO(derrandz): refactor to use networkMessage
	if msg.IsRequest() && msg.Topic == types.PocketTopic_P2P_PING {
		pongmsg := types.NetworkMessage{
			Nonce:       msg.Nonce,
			Topic:       types.PocketTopic_P2P_PONG,
			Source:      m.address,
			Destination: msg.Source,
		}
		pongbytes, err := m.c.encode(pongmsg)

		if err != nil {
			return err
		}

		err = m.respond(uint32(msg.Nonce), false, msg.Source, pongbytes, true)

		if err != nil {
			return err
		}
	}
	return nil
}

// Discuss: why is m not a pointer?
func (m *networkModule) broadcast(msg *types.NetworkMessage, isroot bool) error {
	m.clog(isroot, "Starting gossip round, level is", msg.Level)

	// var mmutex sync.Mutex

	var list *types.Peerlist = m.peerlist
	var sourceAddr string = m.externaladdr

	var toplevel int = int(getTopLevel(list))
	var currentlevel int = toplevel - 1

	if !isroot {
		m.clog(isroot, "Received gossip message from level", msg.Level)
		currentlevel = int(msg.Level)
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

	for ; currentlevel > 0; currentlevel-- {
		msg.Level = int32(currentlevel)
		m.clog(isroot, "Gossiping to peers at level", currentlevel, "message level=", msg.Level)
		targetlist := getTargetList(list, m.id, toplevel, currentlevel)

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
		var l_msg, r_msg types.NetworkMessage
		var l_data, r_data []byte

		{
			l_msg = *msg
			l_msg.Source = sourceAddr
			l_msg.Destination = left.Addr()
			encm, errenc := m.c.encode(l_msg)
			if errenc != nil {
				return errenc
			}
			l_data = encm
		}

		{
			r_msg = *msg
			r_msg.Source = sourceAddr
			r_msg.Destination = right.Addr()
			encm, errenc := m.c.encode(r_msg)
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
		m.clog(isroot, "Got acks from left and right")

		if l_err != nil {
			return l_err
		}
		m.clog(isroot, "(raintree) left peer: ACK")

		if r_err != nil {
			return r_err
		}
		m.clog(isroot, "(raintree) right peer: ACK")
	}

	m.clog(isroot, "Done broadcasting")

	for _, handler := range m.handlers[types.BroadcastDoneEvent] {
		handler(m)
	}
	return nil
}

func (m *networkModule) handle() {
	var msg *types.NetworkMessage
	var mx sync.Mutex

	m.log("Handling...")
	for w := range m.sink {
		nonce, data, srcaddr, encoded := (&w).Implode()

		if encoded {
			decoded, err := m.c.decode(data)
			if err != nil {
				m.log("Error decoding data", err.Error())
				continue
			}

			mx.Lock()
			msgi := decoded.(types.NetworkMessage)
			msg = &msgi
			msg.Nonce = int32(nonce)
			mx.Unlock()
		} else {
			msg.Data = &anypb.Any{}
			msg.Nonce = int32(nonce)
			msg.Source = srcaddr
		}

		switch msg.Topic {

		case types.PocketTopic_CONSENSUS:
			mx.Lock()
			ack := &types.NetworkMessage{Nonce: msg.Nonce, Level: msg.Level, Topic: types.PocketTopic_CONSENSUS, Source: m.externaladdr, Destination: msg.Source}
			encoded, err := m.c.encode(*ack)
			if err != nil {
				m.log("Error encoding m for gossipaCK", err.Error())
			}

			err = m.respond(uint32(msg.Nonce), false, srcaddr, encoded, true)
			mx.Unlock()
			if err != nil {
				m.log("Error encoding msg for gossipaCK", err.Error())
			}
			<-time.After(time.Millisecond * 5)
			m.log("Acked to", ack.Destination)
			go m.broadcast(&types.NetworkMessage{
				Nonce:       msg.Nonce,
				Level:       msg.Level,
				Source:      msg.Source,
				Topic:       msg.Topic,
				Destination: msg.Destination,
				Data:        msg.Data,
			}, false)

		default:
			m.log("Unrecognized message topic", msg.Topic, "from", msg.Source, "to", msg.Destination, "@node", m.address)
		}
	}
}

func (m *networkModule) handleInbound(conn net.Conn, addr string) {
	pipe, exists := m.inbound.get(addr)
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

func (m *networkModule) on(e types.PeerEvent, handler func(...interface{})) {
	if m.handlers != nil {
		if hmap, exists := m.handlers[e]; exists {
			hmap = append(hmap, handler)
		} else {
			m.handlers[e] = append(make([]func(...interface{}), 0), handler)
		}
	}
}

func (m *networkModule) peerConnected(p *socket) error {
	m.log("Peer connected", p.addr)
	return nil
}

func (m *networkModule) peerDisconnected(p *socket) error {
	return nil
}

func (m *networkModule) initConnectionPools() {
	socketFactory := func() interface{} {
		sck := NewSocket(m.config)
		return interface{}(sck) // TODO(derrandz): remember to change this if you end up using unsafe_ptr instead of interface{}
	}

	m.inbound = NewRegistry(m.config.MaxInbound, socketFactory)
	m.outbound = NewRegistry(m.config.MaxInbound, socketFactory)
}
