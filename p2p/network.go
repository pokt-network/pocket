package p2p

import (
	"net"

	"github.com/pokt-network/pocket/p2p/types"
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

	pipe.runner = m
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

func (m *p2pModule) handleInbound(conn net.Conn, addr string) {
	var pipe *socket
	obj, exists := m.inbound.Get(addr)
	pipe = obj.(*socket)
	if !exists {
		pipe.runner = m
		go pipe.open(InboundIoPipe, addr, conn, m.peerConnected, m.peerDisconnected)

		var err error
		select {
		case <-pipe.done:
			err = pipe.err.error
		case <-pipe.errored:
			err = pipe.err.error
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

func newP2PModule() *p2pModule {
	return &p2pModule{
		c: types.NewProtoMarshaler(),

		sink: make(chan types.Work, 100), // TODO(derrandz): rethink whether this should be buffered

		peerlist: nil,

		done:   make(chan uint, 1),
		ready:  make(chan uint, 1),
		closed: make(chan uint, 1),

		handlers: make(map[types.PeerEvent][]func(...interface{}), 0),
		errored:  make(chan uint, 1),
	}
}
