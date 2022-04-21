package p2p

import (
	"fmt"
	"log"
	"net"

	"context"

	"github.com/pokt-network/pocket/p2p/types"
	"google.golang.org/protobuf/proto"
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
	go pipe.open(context.Background(), func() (string, types.SocketType, net.Conn) {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			return "", types.UndefinedSocketType, nil
		}
		return addr, types.Outbound, conn
	}, m.peerConnected, m.peerDisconnected)

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
		<-pipe.done
		pipe = nil
	}

	return pipe, err
}

func (m *p2pModule) send(addr string, msg []byte, wrapped bool) error {
	pipe, derr := m.dial(addr)
	if derr != nil {
		return derr
	}

	_, werr := pipe.writeChunk(msg, false, 0, wrapped)
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

	m.log("Listening on", m.address)
	close(m.ready)

	m.log("Listening at", m.protocol, m.address, "...")
accepter:
	for {
		select {
		case <-m.done:
			break accepter
		default:
			{
				conn, err := m.listener.Accept()
				if err != nil && m.isListening.Load() {
					m.log("Error receiving an inbound connection: ", err.Error())
					// TODO(derrandz) ignore use of closed network connection error when listener has closed
					m.error(err)
					break accepter
				}

				if !m.isListening.Load() {
					break accepter
				}

				addr := conn.RemoteAddr().String()
				go m.poolInbound(conn, addr)
			}
		}
	}

	m.listener.Lock()
	m.listener.Close()
	m.listener.TCPListener = nil
	m.listener.Unlock()
	m.isListening.Store(false)

	m.releaseSockets()

	return nil
}

// TODO(derrandz): this: msg []byte, wrapped bool) is repeat everything, maybe a struct for this?
func (m *p2pModule) request(addr string, msg []byte, wrapped bool) ([]byte, error) {
	pipe, derr := m.dial(addr)
	if derr != nil {
		return nil, derr
	}
	var response types.Packet

	response, rerr := pipe.writeChunkAckful(msg, wrapped)
	if rerr != nil {
		return nil, rerr
	}

	return response.Data, nil
}

func (m *p2pModule) respond(nonce uint32, iserroreof bool, addr string, msg []byte, wrapped bool) error {
	pipe, derr := m.dial(addr)
	if derr != nil {
		return derr
	}

	_, werr := pipe.writeChunk(msg, iserroreof, nonce, wrapped)
	if werr != nil {
		return werr
	}

	return nil
}

func (m *p2pModule) consume() {
	for w := range m.sink {

		if w.IsEncoded {
			p2pMsg := &types.P2PMessage{}

			//err := m.c.Unmarshal(w.Data, &p2pMsg)
			err := proto.Unmarshal(w.Data, p2pMsg)
			if err != nil {
				// TODO(derrandz): this is a place holder error handling pattern as discussed in protocol hours
				log.Fatalf("handleBroadcast: failed to unmarsha received message: %s", err)
				continue
			}

			m.log("Nonce=", w.Nonce, "from=", w.From, "p2pMsg=", p2pMsg)
			m.handle(w.Nonce, w.From, p2pMsg)
		} else {
			m.log(fmt.Sprintf("Consume: received a %d wire-level bytes from %s. Left unhandled.", len(w.Data), w.From))
		}
	}
}

func (m *p2pModule) handle(nonce uint32, sourceAddr string, msg *types.P2PMessage) {
	m.log("Message metadata", msg.Metadata, "broadcast?", msg.Metadata.Broadcast)
	if msg.Metadata.Broadcast {
		err := m.handleBroadcast(nonce, sourceAddr, msg)
		if err != nil {
			m.log("Handle: encountered error while handling broadcast message: %s", err)
		}
	}

	// TODO(derrandz): prevent network from listening if bus is nil and remove this check
	// Temporarily added to fix an isolated test
	//if m.bus != nil {
	//	m.bus.PublishEventToBus(msg.Payload)
	//} else {
	//	log.Fatal("[ERROR]: No bus was set!")
	//}
}

func (m *p2pModule) poolInbound(conn net.Conn, addr string) {
	var pipe *socket
	obj, exists := m.inbound.Get(addr)
	pipe = obj.(*socket)
	if !exists {
		pipe.runner = m
		connect := func() (string, types.SocketType, net.Conn) {
			return addr, types.Inbound, conn
		}
		// TODO(derrandz): pass proper context instead of background
		go pipe.open(context.Background(), connect, m.peerConnected, m.peerDisconnected)

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
			<-pipe.done
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

func (m *p2pModule) peerConnected(ctx context.Context, p *socket) error {
	m.log("Peer connected", p.addr)
	return nil
}

func (m *p2pModule) peerDisconnected(ctx context.Context, p *socket) error {
	return nil
}

func (m *p2pModule) releaseSockets() {
	for _, s := range m.inbound.Elements() {
		sckt := s.(socket)
		sckt.close()
	}

	for _, s := range m.outbound.Elements() {
		sckt := s.(socket)
		sckt.close()
	}
}

func newP2PModule() *p2pModule {
	return &p2pModule{
		c: types.NewProtoMarshaler(),

		sink: make(chan types.Packet, 100), // TODO(derrandz): rethink whether this should be buffered

		peerlist: nil,

		done:   make(chan uint, 1),
		ready:  make(chan uint, 1),
		closed: make(chan uint, 1),

		handlers: make(map[types.PeerEvent][]func(...interface{}), 0),
		errored:  make(chan uint, 1),
	}
}
