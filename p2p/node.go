package p2p

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/pokt-network/pocket/p2p/types"
	"go.uber.org/atomic"
	"google.golang.org/protobuf/proto"
)

type (
	P2PNode interface {
		types.Logger
		Start() error
		Stop()
		Serve()
		Handle()
		OnNewMessage(MessageHandler)
		IsRunning() bool
		HandleConnection(Direction, net.Conn, chan struct{})
		Dial(string) error
		Send(uint32, string, []byte, bool) error
		SendMessage(uint32, string, *types.P2PMessage) error
		Request(context.Context, string, []byte, bool) ([]byte, error)
		RequestMessage(context.Context, string, *types.P2PMessage) (*types.P2PMessage, error)
		Ping(string) error
		Pong(uint32, string) error
		Broadcast([]byte, bool, int, bool) error
		BroadcastMessage(*types.P2PMessage, bool, int) error
		Address() string
		SetId(id int)
		AddMsgToHistory(int64)
		IsMsgInHistory(int64) bool
	}
	p2pNode struct {
		sync.Mutex
		net.Listener
		net.Dialer
		*wireCodec
		types.Logger

		config map[string]interface{}
		ID     int

		peerList []peerInfo
		peers    *peerMap
		requests *requestMap

		wg        sync.WaitGroup
		quit      chan struct{}
		isRunning atomic.Bool

		sink chan Packet

		handlers []MessageHandler

		historyLock     sync.Mutex
		messagesHistory map[int64]bool
	}

	p2pConn struct {
		net.Conn
		context.Context
		cancel context.CancelFunc

		direction Direction
		address   string
		ID        int

		// io
		readBuffer     []byte
		writeBuffer    []byte
		writeNonce     uint32
		writeIsWrapped bool
		writeMutex     sync.Mutex
		writeSignals   chan struct{}

		reader *bufio.Reader
		writer *bufio.Writer
	}

	peerMap struct {
		sync.RWMutex
		m map[string]*p2pConn
	}

	peerInfo struct {
		ID      int
		address string
	}
	requestMap struct {
		sync.RWMutex
		m           map[uint32]chan []byte
		numRequests uint32
	}

	Packet struct {
		Nonce   uint32
		IsProto bool
		Data    []byte
	}

	MessageHandler func(*types.P2PMessage)

	Direction string
)

const (
	DirectionInbound  Direction = "inbound"
	DirectionOutbound Direction = "outbound"
)

var (
	PingNonce uint32 = binary.BigEndian.Uint32([]byte("ping"))
	PongNonce uint32 = binary.BigEndian.Uint32([]byte("pong"))
)

// fancy golang interface-implementation checking
var _ P2PNode = &p2pNode{}

// constructors

func NewP2PNode(config map[string]interface{}) *p2pNode {
	node := &p2pNode{
		ID:              config["id"].(int),
		Mutex:           sync.Mutex{},
		Listener:        nil,
		Dialer:          net.Dialer{},
		wireCodec:       NewWireCodec(),
		Logger:          types.NewLogger(os.Stdout),
		requests:        NewRequestMap(),
		peers:           NewPeerMap(),
		config:          config,
		wg:              sync.WaitGroup{},
		quit:            make(chan struct{}, 1),
		sink:            make(chan Packet, 100),
		handlers:        make([]MessageHandler, 0),
		peerList:        make([]peerInfo, 0),
		messagesHistory: make(map[int64]bool),
	}

	// TODO(derrandz): refactor when the addressbook is spec'd out
	if _, exists := config["peers"]; exists && len(config["peers"].([]string)) > 0 {
		for _, peerString := range config["peers"].([]string) {
			idAndAddr := strings.Split(peerString, "@")
			id, _ := strconv.Atoi(idAndAddr[0])
			node.peerList = append(node.peerList, peerInfo{
				ID:      id,
				address: idAndAddr[1],
			})
		}
	}

	return node
}

func NewP2PConn(direction Direction, c net.Conn, readBufferSize, writeBufferSize int) *p2pConn {
	var address string
	ctx, cancel := context.WithCancel(context.Background())
	if c != nil {
		address = c.RemoteAddr().String()
	}
	return &p2pConn{
		Conn:         c,
		Context:      ctx,
		cancel:       cancel,
		direction:    direction,
		address:      address,
		readBuffer:   make([]byte, readBufferSize, readBufferSize),
		writeBuffer:  make([]byte, 0),
		writeNonce:   0,
		writeSignals: make(chan struct{}, 1),
		reader:       bufio.NewReader(c),
		writer:       bufio.NewWriter(c),
	}
}

func NewPeerMap() *peerMap {
	return &peerMap{m: make(map[string]*p2pConn)}
}

func NewRequestMap() *requestMap {
	return &requestMap{m: make(map[uint32]chan []byte)}
}

func NewPacket(isProto bool, Nonce uint32, data []byte) *Packet {
	return &Packet{
		IsProto: isProto,
		Nonce:   Nonce,
		Data:    data,
	}
}

// the P2PNode interface implementation

func CreateP2PNode(config map[string]interface{}) P2PNode {
	return NewP2PNode(config)
}

func (n *p2pNode) Start() error {
	tcpAddr, err := net.ResolveTCPAddr("tcp", n.config["address"].(string))
	if err != nil {
		n.Error("Failed to resolve TCP address", err)
		return err
	}

	l, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		n.Error("Failed to listen on TCP address", err)
		return err
	}
	n.Listener = l
	n.wg.Add(1)
	go n.Serve()
	n.isRunning.Store(true)
	return nil
}

func (n *p2pNode) Serve() {
	defer n.wg.Done()

	for {
		conn, err := n.Listener.Accept()
		if err != nil {
			select {
			case <-n.quit:
				n.Debug("Quitting")
				return
			default:
				n.Error("accept error", err)
			}
		} else {
			n.wg.Add(1)
			go func() {
				n.HandleConnection(DirectionInbound, conn, nil)
				n.wg.Done()
				n.Log("inbound connection closed")
			}()
		}
	}
}

func (n *p2pNode) Stop() {
	close(n.quit)
	n.Listener.Close()
	n.isRunning.Store(false)
	n.wg.Wait()
}

func (n *p2pNode) IsRunning() bool {
	return n.isRunning.Load()
}

func (n *p2pNode) HandleConnection(direction Direction, c net.Conn, signaler chan struct{}) {
	var iowg sync.WaitGroup

	p2pc := NewP2PConn(
		direction,
		c,
		n.config["readBufferSize"].(int),
		n.config["writeBufferSize"].(int),
	)

	n.Info("Handling a new connection")

	n.peers.Lock()
	n.peers.m[c.RemoteAddr().String()] = p2pc
	n.peers.Unlock()

	n.Info("New connection successfully pooled")

	// add handshake logic

	// kick off read loop
	iowg.Add(1)
	go func(p *p2pConn) {
		defer func() {
			p.Conn.Close()
			iowg.Done()
			n.Info("Read loop closing")
			p.cancel()
		}()

		if signaler != nil {
			signaler <- struct{}{}
		}

		for {
			select {
			case <-n.quit:
				n.Info("Read: Got it?")
				return
			case <-p.Context.Done():
				n.Info("Read: Got it?")
				return
			default:
				n.Debug("Read: blocking")
				if nbytes, err := io.ReadFull(p.reader, p.readBuffer[:WireCodecHeaderSize]); err != nil || nbytes == 0 {
					n.Debug("read error", err)
					return
				}

				_, nonce, isProto, size, err := n.decodeHeader(p.readBuffer[:WireCodecHeaderSize])
				if err != nil {
					n.Debug("header decoding", err)
					return
				}

				// TODO(derrandz): replace with configurable max value or keep it as is (i.e: max=chunk size) ??
				if size > uint32(len(p.readBuffer)-WireCodecHeaderSize) {
					n.Debug("invalid body size", err)
					return
				}

				if nbytes, err := io.ReadFull(p.reader, p.readBuffer[WireCodecHeaderSize:WireCodecHeaderSize+int(size)]); err != nil || nbytes == 0 {
					n.Debug("body read error", err)
					return
				}

				buff := append(make([]byte, 0), p.readBuffer[WireCodecHeaderSize:WireCodecHeaderSize+int(size)]...)

				packet := NewPacket(isProto, nonce, buff)

				if nonce != 0 {
					err := n.AnswerPendingRequest(*packet)
					if err != nil {
						n.Debug("pending request answering error", err)
						return
					}
					continue
				}

				n.Debug("Push to sink: BEGIN.")
				n.sink <- *packet
				n.Debug("Push to sink: OK.")
			}
		}
	}(p2pc)

	// kick off write loop
	iowg.Add(1)
	go func(p *p2pConn) {
		defer func() {
			p.Conn.Close()
			iowg.Done()
			n.Info("Write loop closing")
			p.cancel()
		}()

		if signaler != nil {
			signaler <- struct{}{}
		}

		for {
			select {
			case <-n.quit:
				n.Info("Write: Got it?")
				return
			case <-p.Context.Done():
				n.Info("Read: Got it?")
				return

			case <-p.signals():
				p.writeMutex.Lock()
				buff := append(make([]byte, 0), p.writeBuffer...)
				nonce := p.writeNonce
				isWrapped := p.writeIsWrapped
				p.writeBuffer = nil
				p.writeNonce = 0
				p.writeIsWrapped = false
				p.writeMutex.Unlock()

				buff = n.encode(false, nonce, buff, isWrapped)

				if _, err := p.writer.Write(buff); err != nil {
					n.Debug("write error", err)
					return
				}

				if err := p.writer.Flush(); err != nil {
					n.Debug("flush error", err)
					return
				}
			}

		}

	}(p2pc)

	iowg.Wait()

	n.Debug("Connection closed")
}

func (n *p2pNode) Handle() {
	select {
	case packet := <-n.sink:
		n.Log("Got a packet", packet)
		if packet.IsProto {
			msg := &types.P2PMessage{}
			err := proto.Unmarshal(packet.Data, msg)
			if err != nil {
				n.Error("Handle: failed to uunmarshal proto messsage, error", err)
			} else {
				n.Log("Handle: got a proto message")
				n.HandleMessage(packet.Nonce, msg)
			}
		} else {
			n.HandleRawBytes(packet.Nonce, packet.Data)
		}
	}
}

func (n *p2pNode) HandleMessage(nonce uint32, msg *types.P2PMessage) {
	if msg.Metadata.Broadcast {
		n.Log("HandleMessage: got a broadcast message")
		err := n.BroadcastMessage(msg, false, int(msg.Metadata.Level))
		if err != nil {
			n.Error("Handle: encountered error while handling broadcast message: %s", err)
		}
		n.Log("HandleMessage: broadcast message handled")
	}

	if n.IsMsgInHistory(msg.Metadata.Hash) {
		n.Log("HandleMessage: message already in history, not handling")
		return
	}

	n.Lock()
	for _, handle := range n.handlers {
		n.Log("HandleMessage: calling handler")
		handle(msg)
	}
	n.Unlock()
	n.Log("HandleMessage: message handled")
}

func (n *p2pNode) HandleRawBytes(nonce uint32, bytes []byte) {
	n.Warn("HandleRawBytes: received raw bytes, nonce: %d", nonce)
}

func (n *p2pNode) OnNewMessage(handler MessageHandler) {
	defer n.Unlock()
	n.Lock()
	n.handlers = append(n.handlers, handler)
}

func (n *p2pNode) Dial(address string) error {
	n.peers.Lock()
	if _, exists := n.peers.m[address]; exists {
		n.peers.Unlock()
		return nil
	}
	n.peers.Unlock()

	conn, err := n.Dialer.Dial("tcp", address)
	if err != nil {
		return err
	}

	ready := make(chan struct{}, 1)
	n.wg.Add(1)
	go func() {
		n.HandleConnection(DirectionOutbound, conn, ready)
		n.wg.Done()
		n.Log("outbound connection closed")
	}()

	<-ready

	return nil
}

func (n *p2pNode) Send(nonce uint32, address string, data []byte, isProto bool) error {
	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		n.Error("Send: failed to resolve address", err)
		return err
	}

	err = n.Dial(addr.String())
	if err != nil {
		return err
	}

	n.peers.Lock()
	peer := n.peers.m[addr.String()]
	n.peers.Unlock()
	peer.write(nonce, data, isProto)

	return nil
}

func (n *p2pNode) SendMessage(nonce uint32, address string, msg *types.P2PMessage) error {
	msg.Metadata.Destination = address
	msg.Metadata.Source = n.Address()

	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	return n.Send(nonce, address, data, true)
}

func (n *p2pNode) Request(ctx context.Context, address string, data []byte, isProto bool) ([]byte, error) {
	ch, nonce := n.NewRequest()

	err := n.Send(nonce, address, data, isProto)
	if err != nil {
		return nil, err
	}

	select {
	case buff := <-ch:
		return buff, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (n *p2pNode) RequestMessage(ctx context.Context, address string, msg *types.P2PMessage) (*types.P2PMessage, error) {
	msg.Metadata.Destination = address
	msg.Metadata.Source = n.Address()

	data, err := proto.Marshal(msg)
	if err != nil {
		return &types.P2PMessage{}, err
	}

	response, err := n.Request(ctx, address, data, true)
	if err != nil {
		return &types.P2PMessage{}, err
	}

	var responseMsg types.P2PMessage
	err = proto.Unmarshal(response, &responseMsg)
	if err != nil {
		return &types.P2PMessage{}, err
	}

	return &responseMsg, nil
}

func (n *p2pNode) AnswerPendingRequest(packet Packet) error {
	if _, exists := n.requests.m[packet.Nonce]; !exists {
		return errors.New("no pending request with nonce")
	}

	n.requests.m[packet.Nonce] <- packet.Data

	return nil
}

func (n *p2pNode) NewRequest() (chan []byte, uint32) {
	ch := make(chan []byte)
	nonce := n.requests.Next()
	n.requests.Lock()
	n.requests.m[nonce] = ch
	n.requests.Unlock()
	return ch, nonce
}

func (n *p2pNode) Ping(address string) error {
	ping := make([]byte, 4)
	binary.BigEndian.PutUint32(ping, PingNonce)
	// TODO(derrandz): make use of the context to achieve timeout behavior for the ping/pong sequence
	response, err := n.Request(context.Background(), address, ping, false)
	if err != nil {
		return err
	}

	if binary.BigEndian.Uint32(response) != PongNonce {
		return errors.New("invalid pong response")
	}

	return nil
}

func (n *p2pNode) Pong(nonce uint32, address string) error {
	pong := make([]byte, 4)
	binary.BigEndian.PutUint32(pong, n.config["pongNonce"].(uint32))
	return n.Send(nonce, address, pong, false)
}

func (n *p2pNode) Broadcast(msg []byte, isRoot bool, fromLevel int, isWrapped bool) error {
	peerTree := NewRainTree()

	peerTree.SetLeafs(n.peerList)
	peerTree.SetRoot(n.ID)
	err := peerTree.Traverse(
		isRoot,
		fromLevel,
		func(originatorId int, left, right peerInfo, currentLevel int) error {
			n.Log(fmt.Sprintf("Broadcast: originatorId: %d, left: %d, right: %d, currentLevel: %d", originatorId, left.ID, right.ID, currentLevel))
			go n.Send(0, right.address, msg, isWrapped)
			go n.Send(0, left.address, msg, isWrapped)
			return nil
		})

	return err
}

func (n *p2pNode) BroadcastMessage(msg *types.P2PMessage, isRoot bool, fromLevel int) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	err = n.Broadcast(data, isRoot, fromLevel, true)
	if err != nil {
		return err
	}

	// redundancy layer
	if isRoot && n.config["redundancy"].(bool) {
		for _, peer := range n.peerList {
			if peer.ID != n.ID {
				msg.Metadata.Broadcast = false
				go n.SendMessage(0, peer.address, msg)
			}
		}
	}

	return nil
}

func (n *p2pNode) Address() string {
	return n.config["address"].(string)
}

func (n *p2pNode) SetId(id int) {
	n.ID = id
}

func (n *p2pNode) AddMsgToHistory(msgHash int64) {
	n.historyLock.Lock()
	defer n.historyLock.Unlock()
	n.messagesHistory[msgHash] = true
}

func (n *p2pNode) IsMsgInHistory(id int64) bool {
	_, seen := n.messagesHistory[id]
	return seen
}

// p2pConn additional functionality

func (p *p2pConn) write(nonce uint32, data []byte, isWrapped bool) {
	defer p.writeMutex.Unlock()
	p.writeMutex.Lock()
	p.writeBuffer = append(p.writeBuffer, data...)
	p.writeNonce = nonce
	p.writeIsWrapped = isWrapped
	p.signalWrite()
}

func (p *p2pConn) signals() <-chan struct{} {
	return p.writeSignals
}

func (p *p2pConn) signalWrite() {
	p.writeSignals <- struct{}{}
}

// request map functionality

func (reqMap *requestMap) Next() uint32 {
	reqMap.Lock()
	defer reqMap.Unlock()
	reqMap.numRequests++
	return reqMap.numRequests
}

func (pMap *peerMap) IndexById() map[int]*p2pConn {
	pMap.Lock()
	defer pMap.Unlock()
	index := make(map[int]*p2pConn)
	for _, peer := range pMap.m {
		index[peer.ID] = peer
	}
	return index
}
