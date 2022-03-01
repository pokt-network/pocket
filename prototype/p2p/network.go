package p2p

import (
	"net"
	"pocket/p2p/types"
	"sync"
	"time"

	"google.golang.org/protobuf/types/known/anypb"
)

func (g *networkModule) handshake() {}

func (g *networkModule) dial(addr string) (*netpipe, error) {
	// TODO(derrandz): this is equivalent to maxRetries = 1, add logic for > 1
	// TODO(derrandz): should we explictly tell dial to use either inbound or outbound?
	exists := g.inbound.peak(addr)
	g.log("Peaked into inbound clients map for", addr, "found=", exists)
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
		g.log("Error openning pipe", err.Error())
		pipe.close()
		<-pipe.closed
		pipe = nil
	}

	return pipe, err
}

func (g *networkModule) send(addr string, msg []byte, wrapped bool) error {
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

func (g *networkModule) listen() error {
	defer func() {
		g.listening.Store(false)
		close(g.closed)
	}()

	// add address validation and parsing
	listener, err := net.Listen(g.protocol, g.address)
	if err != nil {
		g.log("Error:", err.Error())
		return err
	}

	g.listener.Lock()
	g.listener.TCPListener = listener.(*net.TCPListener)
	g.listener.Unlock()
	g.listening.Store(true)

	close(g.ready)

	g.log("Listening at", g.protocol, g.address, "...")
	for stop := false; !stop; {
		select {
		case <-g.done:
			stop = true
			break
		default:
			{
				conn, err := g.listener.Accept()
				if err != nil && g.listening.Load() {
					g.log("Error receiving an inbound connection: ", err.Error())
					// TODO(derrandz) ignore use of closed network connection error when listener has closed
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

// TODO(derrandz): this: msg []byte, wrapped bool) is repeat everything, maybe a struct for this?
func (g *networkModule) request(addr string, msg []byte, wrapped bool) ([]byte, error) {
	pipe, derr := g.dial(addr)
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

func (g *networkModule) respond(nonce uint32, iserroreof bool, addr string, msg []byte, wrapped bool) error {
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

func (g *networkModule) ping(addr string) (bool, error) {
	// TODO(derrandz): refactor this to use types.NetworkMessage
	var pongbytes []byte

	pingmsg := types.NetworkMessage{Topic: types.PocketTopic_P2P_PING, Source: g.address, Destination: addr}
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
		response, err := g.request(addr, pingbytes, true)

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
func (g *networkModule) pong(msg types.NetworkMessage) error {
	// TODO(derrandz): refactor to use networkMessage
	if msg.IsRequest() && msg.Topic == types.PocketTopic_P2P_PING {
		pongmsg := types.NetworkMessage{
			Nonce:       msg.Nonce,
			Topic:       types.PocketTopic_P2P_PONG,
			Source:      g.address,
			Destination: msg.Source,
		}
		pongbytes, err := g.c.encode(pongmsg)

		if err != nil {
			return err
		}

		err = g.respond(uint32(msg.Nonce), false, msg.Source, pongbytes, true)

		if err != nil {
			return err
		}
	}
	return nil
}

// Discuss: why is m not a pointer?
func (g *networkModule) broadcast(m *types.NetworkMessage, isroot bool) error {
	g.clog(isroot, "Starting gossip round, level is", m.Level)

	// var mmutex sync.Mutex

	var list *types.Peerlist = g.peerlist
	var sourceAddr string = g.externaladdr

	var toplevel int = int(getTopLevel(list))
	var currentlevel int = toplevel - 1

	if !isroot {
		g.clog(isroot, "Received gossip message from level", m.Level)
		currentlevel = int(m.Level)
	}

	SEND_AND_WAIT_FOR_ACK := func(encodedMsg []byte, wg *sync.WaitGroup, target *types.Peer, ack *[]byte, err *error) {
		response, reqerr := g.request(target.Addr(), encodedMsg, true)

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
		g.clog(isroot, "Gossiping to peers at level", currentlevel, "message level=", m.Level)
		targetlist := getTargetList(list, g.id, toplevel, currentlevel)

		var left, right *types.Peer
		{
			lpos := pickLeft(g.id, targetlist)
			rpos := pickRight(g.id, targetlist)

			left = targetlist.Get(lpos)
			right = targetlist.Get(rpos)
		}

		var wg sync.WaitGroup

		var l_ack, r_ack []byte
		var l_err, r_err error = nil, nil
		var l_msg, r_msg types.NetworkMessage
		var l_data, r_data []byte

		{
			l_msg = *m
			l_msg.Source = sourceAddr
			l_msg.Destination = left.Addr()
			encm, errenc := g.c.encode(l_msg)
			if errenc != nil {
				return errenc
			}
			l_data = encm
		}

		{
			r_msg = *m
			r_msg.Source = sourceAddr
			r_msg.Destination = right.Addr()
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
		g.clog(isroot, "Got acks from left and right")

		if l_err != nil {
			return l_err
		}
		g.clog(isroot, "(raintree) left peer: ACK")

		if r_err != nil {
			return r_err
		}
		g.clog(isroot, "(raintree) right peer: ACK")
	}

	g.clog(isroot, "Done broadcasting")

	for _, handler := range g.handlers[types.BroadcastDoneEvent] {
		handler(m)
	}
	return nil
}

func (g *networkModule) handle() {
	var msg *types.NetworkMessage
	var m sync.Mutex

	g.log("Handling...")
	for w := range g.sink {
		nonce, data, srcaddr, encoded := (&w).Implode()

		if encoded {
			decoded, err := g.c.decode(data)
			if err != nil {
				g.log("Error decoding data", err.Error())
				continue
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
				g.log("Error encoding m for gossipaCK", err.Error())
			}

			err = g.respond(uint32(msg.Nonce), false, srcaddr, encoded, true)
			m.Unlock()
			if err != nil {
				g.log("Error encoding msg for gossipaCK", err.Error())
			}
			<-time.After(time.Millisecond * 5)
			g.log("Acked to", ack.Destination)
			go g.broadcast(&types.NetworkMessage{
				Nonce:       msg.Nonce,
				Level:       msg.Level,
				Source:      msg.Source,
				Topic:       msg.Topic,
				Destination: msg.Destination,
				Data:        msg.Data,
			}, false)

		default:
			g.log("Unrecognized message topic", msg.Topic, "from", msg.Source, "to", msg.Destination, "@node", g.address)
		}
	}
}

func (g *networkModule) handleInbound(conn net.Conn, addr string) {
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

		g.log("New connection from", addr, err)
		if err != nil {
			pipe.close()
			<-pipe.closed
			pipe = nil
			g.error(err)
		}

	}
}

func (g *networkModule) on(e types.PeerEvent, handler func(...interface{})) {
	if g.handlers != nil {
		if hmap, exists := g.handlers[e]; exists {
			hmap = append(hmap, handler)
		} else {
			g.handlers[e] = append(make([]func(...interface{}), 0), handler)
		}
	}
}

func (g *networkModule) peerConnected(p *netpipe) error {
	g.log("Peer connected", p.addr)
	return nil
}

func (g *networkModule) peerDisconnected(p *netpipe) error {
	return nil
}
