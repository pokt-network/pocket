package p2p

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pokt-network/pocket/p2p/types"
	commonTypes "github.com/pokt-network/pocket/shared/types"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestP2PNode_StartStop(t *testing.T) {
	config := map[string]interface{}{
		"address":         "127.0.0.1:10001",
		"readBufferSize":  1024,
		"writeBufferSize": 1024,
		"redundancy":      false,
		"id":              1,
	}
	node := NewP2PNode(config)
	err := node.Start()

	assert.Nil(
		t,
		err,
		"P2PNode.Start() should not return an error",
	)

	assert.NotNil(
		t,
		node.Listener,
		"P2PNode.Start() should set the tcp listener",
	)

	node.Stop()

	assert.False(
		t,
		node.IsRunning(),
		"P2PNode.Stop() should set the running flag to false",
	)
}

// This test case setup does the following:
// 1- start tcp server
// 2- establish a connection to the tcp server
// 3- return the connection and server
func Setup_TestP2PNode_HandleConnection() *p2pNode {
	config := map[string]interface{}{
		"address":         "127.0.0.1:10001",
		"readBufferSize":  1024,
		"writeBufferSize": 1024,
		"redundancy":      false,
		"id":              1,
	}

	node := NewP2PNode(config)

	err := node.Start()
	if err != nil {
		log.Fatal("Setup_TestP2PNode_HandleConnection() failed to start node")
	}

	return node
}

func TestP2PNode_HandleConnection(t *testing.T) {
	node := Setup_TestP2PNode_HandleConnection()

	t.Run("HandleConnection handles inbound connections properly", func(t *testing.T) {

		conn, err := net.Dial("tcp", node.Address())
		if err != nil {
			log.Fatal("TestP2PNode_HandleConnection() failed to dial the p2p node")
		}

		fullmsg := node.encode(false, 0, []byte("hello"), false)

		go func() {
			_, err := conn.Write(fullmsg)
			if err != nil {
				log.Fatal("Failed to write to inbound connection")
			}
		}()

		<-time.After(time.Millisecond * 10) // give the node the time to handle the new inbound connection

		node.peers.Lock()
		pconn, exists := node.peers.m[conn.LocalAddr().String()]
		node.peers.Unlock()

		assert.True(
			t,
			exists,
			"P2PNode.HandleConnection() should add the connection as a peer to the peers map",
		)

		assert.NotNil(
			t,
			pconn.Conn,
			"P2PNode.HandleConnection() should set the connection for the p2p connection",
		)
		packet := <-node.sink

		assert.Equal(
			t,
			packet.Data,
			[]byte("hello"),
			"P2PNode.HandleConnection() should push incoming connection data to the node's sink",
		)

		go pconn.write(0, []byte("hello back"), false)

		data := make([]byte, 1024)
		conn.Read(data)

		_, decodedData, _, _ := node.decode(data)

		assert.Equal(
			t,
			decodedData,
			[]byte("hello back"),
			"P2PNode.HandleConnection() should push incoming connection data to the node's sink",
		)

		conn.Close()
	})

	t.Run("HandleConnection handles outbound connections IO properly", func(t *testing.T) {

		l, err := net.Listen("tcp", "127.0.0.1:10003")
		if err != nil {
			log.Fatal("TestP2PNode_HandleConnection() failed to listen for connections coming from the p2p node")
		}

		go func() {
			<-time.After(time.Millisecond * 5)
			err := node.Dial("127.0.0.1:10003")
			if err != nil {
				log.Fatal("TestP2PNode_HandleConnection() p2p node failed to dial the mock server")
			}
		}()

		conn, err := l.Accept()
		if err != nil {
			log.Fatal("TestP2PNode_HandleConnection() failed to accept a connection")
		}

		<-time.After(time.Millisecond * 10) // give the node the time to handle the new inbound connection
		node.peers.Lock()
		pconn, exists := node.peers.m[conn.LocalAddr().String()]
		node.peers.Unlock()

		assert.True(
			t,
			exists,
			"P2PNode.HandleConnection() should add the connection as a peer to the peers map",
		)

		assert.NotNil(
			t,
			pconn.Conn,
			"P2PNode.HandleConnection() should set the connection for the p2p connection",
		)

		fullmsg := node.encode(false, 0, []byte("hello"), false)

		go func() {
			_, err := conn.Write(fullmsg)
			if err != nil {
				log.Fatal("Failed to write to inbound connection")
			}
		}()

		<-time.After(time.Millisecond * 10) // give the node the time to handle the new inbound connection

		packet := <-node.sink

		assert.Equal(
			t,
			packet.Data,
			[]byte("hello"),
			"P2PNode.HandleConnection() should push incoming connection data to the node's sink",
		)

		go pconn.write(0, []byte("hello back"), false)

		data := make([]byte, 1024)
		conn.Read(data)

		_, decodedData, _, _ := node.decode(data)

		assert.Equal(
			t,
			decodedData,
			[]byte("hello back"),
			"P2PNode.HandleConnection() should push incoming connection data to the node's sink",
		)

		conn.Close()
		l.Close()
	})

	Teardown_TestP2PNode_HandleConnection(node)
}

func Teardown_TestP2PNode_HandleConnection(node *p2pNode) {
	node.Stop()
}

func Setup_TestP2PNode_Dial() (*p2pNode, net.Listener) {
	l, err := net.Listen("tcp", "127.0.0.1:10003")
	if err != nil {
		log.Fatal("TestP2PNode_Dial() failed to start mock server to receive connections coming from the p2p node")
	}

	config := map[string]interface{}{
		"address":         "127.0.0.1:10001",
		"readBufferSize":  1024,
		"writeBufferSize": 1024,
		"redundancy":      false,
		"id":              1,
	}

	node := NewP2PNode(config)

	err = node.Start()
	if err != nil {
		log.Fatal("Setup_TestP2PNode_HandleConnection() failed to start node")
	}

	return node, l
}

func TestP2PNode_Dial(t *testing.T) {
	node, l := Setup_TestP2PNode_Dial()

	node.peers.Lock()
	assert.Equal(
		t,
		0,
		len(node.peers.m),
		"P2PNode.Dial() should not have any peers before none were dialed.",
	)
	node.peers.Unlock()

	go func() {
		err := node.Dial(l.Addr().String())

		assert.Nilf(
			t,
			err,
			"P2PNode.Dial() failed to dial the mock server. err %s", err,
		)
	}()

	conn, err := l.Accept()
	if err != nil {
		log.Fatal("TestP2PNode_Dial() mock server failed to accept a connection")
	}

	t.Log("Connected to mock server")

	<-time.After(time.Millisecond * 10)
	node.peers.Lock()
	pconn, exists := node.peers.m[conn.LocalAddr().String()]
	node.peers.Unlock()

	assert.True(
		t,
		exists,
		"P2PNode.Dial() should add the connection as a peer to the peers map",
	)

	assert.NotNil(
		t,
		pconn.Conn,
		"P2PNode.Dial() should set the connection for the p2p connection",
	)
	t.Log("Okay")

	go func() {
		<-time.After(time.Millisecond * 10)
		conn.Close()
		l.Close()
	}()
	Teardown_TestP2PNode_Dial(node)
}

func Teardown_TestP2PNode_Dial(node *p2pNode) {
	node.Stop()
}

func Setup_TestP2PNode_Send() *p2pNode {
	config := map[string]interface{}{
		"address":         "127.0.0.1:10001",
		"readBufferSize":  1024,
		"writeBufferSize": 1024,
		"redundancy":      false,
		"id":              1,
	}

	node := NewP2PNode(config)

	err := node.Start()
	if err != nil {
		log.Fatal("Setup_TestP2PNode_HandleConnection() failed to start node")
	}

	return node
}

func TestP2PNode_Send(t *testing.T) {
	node := Setup_TestP2PNode_Send()

	t.Run("Send sends to an inbound connection successfully", func(t *testing.T) {
		var wg sync.WaitGroup

		conn, err := net.Dial("tcp", node.Address())
		if err != nil {
			log.Fatal("TestP2PNode_HandleConnection() failed to dial the p2p node")
		}

		fullmsg := node.encode(false, 0, []byte("hello"), false)
		wg.Add(1)
		go func() {
			<-time.After(time.Millisecond * 10)
			err := node.Send(0, conn.LocalAddr().String(), fullmsg, false)

			assert.Nil(
				t,
				err,
				"P2PNode.Send() should not return an error when sending to an inbound connection that is active",
			)
			wg.Done()
		}()

		data := make([]byte, 1024)
		conn.Read(data)

		wg.Wait()

		_, decodedData, _, _ := node.decode(data)

		assert.Equal(
			t,
			fullmsg,
			decodedData,
			"P2PNode.Send() should send the data to the inbound connection uncorrupted",
		)

		conn.Close()
	})

	t.Run("Send sends to an outbound connection successfully", func(t *testing.T) {
		var wg sync.WaitGroup
		l, err := net.Listen("tcp", "127.0.0.1:10003")
		if err != nil {
			log.Fatal("TestP2PNode_HandleConnection() failed to listen for connections coming from the p2p node")
		}

		fullmsg := []byte("hello")

		wg.Add(1)
		go func() {
			err := node.Send(0, l.Addr().String(), fullmsg, false)
			assert.Nil(
				t,
				err,
				"P2PNode.Send() should not return an error when sending to an inbound connection that is active",
			)
			wg.Done()
		}()

		conn, err := l.Accept()
		if err != nil {
			log.Fatal("TestP2PNode_Send() mock server failed to accept a connection")
		}

		wg.Wait()

		data := make([]byte, 1024)
		conn.Read(data)

		wg.Wait()

		_, decodedData, _, _ := node.decode(data)

		assert.Equal(
			t,
			fullmsg,
			decodedData,
			"P2PNode.Send() should send the data to the inbound connection uncorrupted",
		)

		conn.Close()
		l.Close()
	})

	Teardown_TestP2PNode_Send(node)
}
func Teardown_TestP2PNode_Send(node *p2pNode) {
	node.Stop()
}

func TestP2PNode_SendMessage(t *testing.T) {
	node := Setup_TestP2PNode_Send()

	t.Run("Send sends to an inbound connection successfully", func(t *testing.T) {
		var wg sync.WaitGroup

		conn, err := net.Dial("tcp", node.Address())
		if err != nil {
			log.Fatal("TestP2PNode_HandleConnection() failed to dial the p2p node")
		}

		fullmsg := types.NewP2PMessage(0, 0, conn.LocalAddr().String(), node.config["address"].(string), &commonTypes.PocketEvent{
			Topic: commonTypes.PocketTopic_CONSENSUS_MESSAGE_TOPIC,
		})

		wg.Add(1)
		go func() {
			<-time.After(time.Millisecond * 10)
			err := node.SendMessage(0, conn.LocalAddr().String(), fullmsg)

			assert.Nil(
				t,
				err,
				"P2PNode.SendMessage() should not return an error when sending to an inbound connection that is active",
			)
			wg.Done()
		}()

		data := make([]byte, 1024)
		conn.Read(data)

		wg.Wait()

		_, decodedData, _, _ := node.decode(data)
		receivedMessage := &types.P2PMessage{}
		err = proto.Unmarshal(decodedData, receivedMessage)

		assert.Nil(
			t,
			err,
			"P2PNode.SendMessage() should send the data to the inbound connection uncorrupted",
		)
		assert.True(
			t,
			AreProtoMessagesEqual(fullmsg, receivedMessage),
			"P2PNode.SendMessage() should send the data to the inbound connection uncorrupted",
		)

		conn.Close()
	})

	t.Run("Send sends to an outbound connection successfully", func(t *testing.T) {
		var wg sync.WaitGroup
		l, err := net.Listen("tcp", "127.0.0.1:10003")
		if err != nil {
			log.Fatal("TestP2PNode_SendMessage() failed to listen for connections coming from the p2p node")
		}

		fullmsg := types.NewP2PMessage(0, 0, l.Addr().String(), node.config["address"].(string), &commonTypes.PocketEvent{
			Topic: commonTypes.PocketTopic_CONSENSUS_MESSAGE_TOPIC,
		})

		wg.Add(1)
		go func() {
			err := node.SendMessage(0, l.Addr().String(), fullmsg)
			assert.Nil(
				t,
				err,
				"P2PNode.SendMessage() should not return an error when sending to an inbound connection that is active",
			)
			wg.Done()
		}()

		conn, err := l.Accept()
		if err != nil {
			log.Fatal("TestP2PNode_SendMessag() mock server failed to accept a connection")
		}

		wg.Wait()

		data := make([]byte, 1024)
		conn.Read(data)

		wg.Wait()

		_, decodedData, _, _ := node.decode(data)
		receivedMessage := &types.P2PMessage{}
		err = proto.Unmarshal(decodedData, receivedMessage)

		assert.Nil(
			t,
			err,
			"P2PNode.SendMessage() should send the data to the inbound connection uncorrupted",
		)

		assert.True(
			t,
			AreProtoMessagesEqual(fullmsg, receivedMessage),
			"P2PNode.SendMessage() should send the data to the inbound connection uncorrupted",
		)

		conn.Close()
		l.Close()
	})

	Teardown_TestP2PNode_Send(node)
}

func Setup_TestP2PNode_Request() *p2pNode {
	config := map[string]interface{}{
		"address":         "127.0.0.1:10001",
		"readBufferSize":  1024,
		"writeBufferSize": 1024,
		"redundancy":      false,
		"id":              1,
	}

	node := NewP2PNode(config)

	err := node.Start()
	if err != nil {
		log.Fatal("Setup_TestP2PNode_Request() failed to start node")
	}

	return node
}

func TestP2PNode_Request(t *testing.T) {
	node := Setup_TestP2PNode_Request()

	t.Run("Request sends to an inbound connection and receives response back succesfully.", func(t *testing.T) {
		var wg sync.WaitGroup

		conn, err := net.Dial("tcp", node.Address())
		if err != nil {
			log.Fatal("TestP2PNode_Request() failed to dial the p2p node")
		}

		requestMsg := []byte("What time is it?")
		responseMsg := []byte("Time O'Clock")

		wg.Add(1)
		go func() {
			<-time.After(time.Millisecond * 10)
			response, err := node.Request(context.Background(), conn.LocalAddr().String(), requestMsg, false)

			assert.Nil(
				t,
				err,
				"P2PNode.Request() should not return an error when sending to an inbound connection that is active",
			)

			assert.Equal(
				t,
				responseMsg,
				response,
				"P2PNode.Request() should receive the response from the inbound connection",
			)
			wg.Done()
		}()

		data := make([]byte, 1024)
		conn.Read(data)

		nonce, decodedData, _, _ := node.decode(data)

		assert.Equal(
			t,
			requestMsg,
			decodedData,
			"P2PNode.Request(): the mock inbound should receive the request data uncorrupted",
		)

		responseData := node.encode(false, nonce, responseMsg, false)

		go conn.Write(responseData)

		wg.Wait()

		conn.Close()
	})

	t.Run("Request sends to an outbound connection and receives response back succesfully.", func(t *testing.T) {
		var wg sync.WaitGroup
		l, err := net.Listen("tcp", "127.0.0.1:10003")
		if err != nil {
			log.Fatal("TestP2PNode_Request() failed to listen for connections coming from the p2p node")
		}

		requestMsg := []byte("What time is it?")
		responseMsg := []byte("Time O'Clock")

		wg.Add(1)
		go func() {
			<-time.After(time.Millisecond * 10)
			response, err := node.Request(context.Background(), l.Addr().String(), requestMsg, false)

			assert.Nil(
				t,
				err,
				"P2PNode.Request() should not return an error when sending to an inbound connection that is active",
			)

			assert.Equal(
				t,
				responseMsg,
				response,
				"P2PNode.Request() should receive the response from the inbound connection",
			)
			wg.Done()
		}()

		conn, err := l.Accept()
		if err != nil {
			log.Fatal("TestP2PNode_Request() mock server failed to accept a connection")
		}

		data := make([]byte, 1024)
		conn.Read(data)

		nonce, decodedData, _, _ := node.decode(data)

		assert.Equal(
			t,
			requestMsg,
			decodedData,
			"P2PNode.Request(): the mock inbound should receive the request data uncorrupted",
		)

		responseData := node.encode(false, nonce, responseMsg, false)

		go conn.Write(responseData)

		wg.Wait()

		conn.Close()
		l.Close()
	})

	Teardown_TestP2PNode_Request(node)
}

func Teardown_TestP2PNode_Request(node *p2pNode) {
	node.Stop()
}

func TestP2PNode_RequestMessage(t *testing.T) {
	node := Setup_TestP2PNode_Request()

	t.Run("Request sends to an inbound connection and receives response back succesfully.", func(t *testing.T) {
		var wg sync.WaitGroup

		conn, err := net.Dial("tcp", node.Address())
		if err != nil {
			log.Fatal("TestP2PNode_Request() failed to dial the p2p node")
		}

		requestMsg := types.NewP2PMessage(0, 0, conn.LocalAddr().String(), node.config["address"].(string), &commonTypes.PocketEvent{
			Topic: commonTypes.PocketTopic_CONSENSUS_MESSAGE_TOPIC,
		})
		responseMsg := types.NewP2PMessage(0, 0, conn.LocalAddr().String(), node.config["address"].(string), &commonTypes.PocketEvent{
			Topic: commonTypes.PocketTopic_CONSENSUS_MESSAGE_TOPIC,
		})

		wg.Add(1)
		go func() {
			<-time.After(time.Millisecond * 10)
			response, err := node.RequestMessage(context.Background(), conn.LocalAddr().String(), requestMsg)

			assert.Nil(
				t,
				err,
				"P2PNode.RequestMessage() should not return an error when sending to an inbound connection that is active",
			)

			assert.True(
				t,
				AreProtoMessagesEqual(responseMsg, response),
				"P2PNode.RequestMessage() should receive the response from the inbound connection",
			)
			wg.Done()
		}()

		data := make([]byte, 1024)
		conn.Read(data)

		nonce, decodedData, _, _ := node.decode(data)

		receivedMsg := &types.P2PMessage{}
		err = proto.Unmarshal(decodedData, receivedMsg)

		assert.Nil(
			t,
			err,
			"P2PNode.RequestMessage(): the mock inbound should receive the request data uncorrupted",
		)

		assert.True(
			t,
			AreProtoMessagesEqual(requestMsg, receivedMsg),
			"P2PNode.RequestMessage(): the mock inbound should receive the request data uncorrupted",
		)

		responseMsgBytes, err := proto.Marshal(responseMsg)
		responseData := node.encode(false, nonce, responseMsgBytes, false)

		go conn.Write(responseData)

		wg.Wait()
		conn.Close()
	})

	t.Run("Request sends to an outbound connection and receives response back succesfully.", func(t *testing.T) {
		var wg sync.WaitGroup
		l, err := net.Listen("tcp", "127.0.0.1:10003")
		if err != nil {
			log.Fatal("TestP2PNode_Request() failed to listen for connections coming from the p2p node")
		}

		requestMsg := types.NewP2PMessage(0, 0, l.Addr().String(), node.config["address"].(string), &commonTypes.PocketEvent{
			Topic: commonTypes.PocketTopic_CONSENSUS_MESSAGE_TOPIC,
		})
		responseMsg := types.NewP2PMessage(0, 0, node.config["address"].(string), l.Addr().String(), &commonTypes.PocketEvent{
			Topic: commonTypes.PocketTopic_CONSENSUS_MESSAGE_TOPIC,
		})

		wg.Add(1)
		go func() {
			<-time.After(time.Millisecond * 10)
			response, err := node.RequestMessage(context.Background(), l.Addr().String(), requestMsg)

			assert.Nil(
				t,
				err,
				"P2PNode.RequestMessage() should not return an error when sending to an inbound connection that is active",
			)

			assert.True(
				t,
				AreProtoMessagesEqual(responseMsg, response),
				"P2PNode.RequestMessage() should receive the response from the inbound connection",
			)
			wg.Done()
		}()

		conn, err := l.Accept()
		if err != nil {
			log.Fatal("TestP2PNode_Request() mock server failed to accept a connection")
		}

		data := make([]byte, 1024)
		conn.Read(data)

		nonce, decodedData, _, _ := node.decode(data)
		receivedMsg := &types.P2PMessage{}
		err = proto.Unmarshal(decodedData, receivedMsg)

		assert.Nil(
			t,
			err,
			"P2PNode.RequestMessage(): the mock inbound should receive the request data uncorrupted",
		)

		assert.True(
			t,
			AreProtoMessagesEqual(requestMsg, receivedMsg),
			"P2PNode.Request(): the mock inbound should receive the request data uncorrupted",
		)

		responseMsgBytes, err := proto.Marshal(responseMsg)
		responseData := node.encode(false, nonce, responseMsgBytes, false)

		go conn.Write(responseData)

		wg.Wait()

		conn.Close()
		l.Close()
	})

	Teardown_TestP2PNode_Request(node)
}

func TestP2PNode_Ping(t *testing.T) {
	t.Skip("Not Implemeneted")
}

func TestP2PNode_Pong(t *testing.T) {
	t.Skip("Not Implemented")
}

func Setup_TestP2PNode_Broadcast() (*p2pNode, []net.Conn, []net.Listener) {
	peers := []string{}
	for i := 0; i < 27; i++ {
		peers = append(peers, fmt.Sprintf("%d@127.0.0.1:%d", i+1, i+1+10000))
	}

	// create p2p node
	config := map[string]interface{}{
		"address":         "127.0.0.1:10001",
		"readBufferSize":  1024,
		"writeBufferSize": 1024,
		"peers":           peers,
		"redundancy":      false,
		"id":              1,
	}

	node := NewP2PNode(config)

	node.ID = 1

	node.peers.Lock()
	nodeRepresentation := NewP2PConn(DirectionOutbound, nil, 0, 0)
	nodeRepresentation.ID = 1
	nodeRepresentation.address = "127.0.0.1:10001"
	node.peers.m["127.0.0.1:10001"] = nodeRepresentation
	node.peers.Unlock()

	err := node.Start()
	if err != nil {
		log.Fatal("Setup_TestP2PNode_Broadcast() failed to start node")
	}

	// create 26 peers
	listeners := make([]net.Listener, 26)
	for i := 0; i < 26; i++ {
		listeners[i], err = net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", 10000+i+2))
		if err != nil {
			log.Fatal("Setup_TestP2PNode_Broadcast() failed to listen for connections coming from the p2p node")
		}
	}

	// connect the p2p peer to the other 26
	for i := 0; i < 26; i++ {
		err := node.Dial(listeners[i].Addr().String())
		if err != nil {
			log.Fatal("Setup_TestP2PNode_Broadcast() p2p node failed to dial the mock peers")
		}
	}

	connections := make([]net.Conn, 26)
	for i := 0; i < 26; i++ {
		connections[i], err = listeners[i].Accept()
		if err != nil {
			log.Fatal("Setup_TestP2PNode_Broadcast() mock server(s) failed to accept the p2p node's connection")
		}
	}

	// modify the peer list to give them proper ids
	for k := range node.peers.m {
		peer := node.peers.m[k]
		if peer.Conn != nil {
			peerId := peer.Conn.RemoteAddr().(*net.TCPAddr).Port - 10000
			node.peers.m[k].ID = peerId
		}
	}

	// return
	return node, connections, listeners
}

func TestP2PNode_Broadcast(t *testing.T) {
	// setup
	node, peers, listeners := Setup_TestP2PNode_Broadcast()

	// track impact for connections
	wg := sync.WaitGroup{}
	rw := sync.RWMutex{}
	impact := map[int]bool{}
	for _, peer := range peers {
		wg.Add(1)
		go func(c net.Conn, m map[int]bool) {
			data := make([]byte, 1024)
			c.SetReadDeadline(time.Now().Add(time.Millisecond * 2000))
			_, err := c.Read(data)
			if err == nil {
				ID := c.LocalAddr().(*net.TCPAddr).Port - 10000
				rw.Lock()
				m[ID] = true
				rw.Unlock()
			}
			wg.Done()
		}(peer, impact)
	}

	// broadcast
	go node.Broadcast([]byte("Hello World"), true, 0, false)

	wg.Wait()

	expectedImpact := GetRainTreeExpectedImpact(1, 27)

	// verify impact on the subset of peers
	assert.Equal(
		t,
		expectedImpact,
		impact,
		"P2PNode.Broadcast() should send the message to all of the expected peers",
	)

	Teardown_TestP2PNode_Broadcast(node, peers, listeners)
}

func TestP2PNode_BroadcastMessage(t *testing.T) {
	// setup
	node, peers, listeners := Setup_TestP2PNode_Broadcast()

	// track impact for connections
	wg := sync.WaitGroup{}
	rw := sync.RWMutex{}
	impact := map[int]bool{}
	data := map[int][]byte{}
	for _, peer := range peers {
		wg.Add(1)
		go func(c net.Conn, m map[int]bool, d map[int][]byte) {
			data := make([]byte, 1024)
			c.SetReadDeadline(time.Now().Add(time.Millisecond * 2000))
			n, err := c.Read(data)
			if err == nil {
				ID := c.LocalAddr().(*net.TCPAddr).Port - 10000
				rw.Lock()
				m[ID] = true
				d[ID] = data[:n]
				rw.Unlock()
			}
			wg.Done()
		}(peer, impact, data)
	}

	// broadcast
	msg := types.NewP2PMessage(0, 0, "", node.config["address"].(string), &commonTypes.PocketEvent{
		Topic: commonTypes.PocketTopic_CONSENSUS_MESSAGE_TOPIC,
	})
	go node.BroadcastMessage(msg, true, 0)

	wg.Wait()

	expectedImpact := GetRainTreeExpectedImpact(1, 27)

	// verify impact on the subset of peers
	assert.Equal(
		t,
		expectedImpact,
		impact,
		"P2PNode.BroadcastMessage() should send the message to all of the expected peers",
	)

	for k, _ := range data {
		_, decodedData, _, err := node.decode(data[k])

		assert.Nil(
			t,
			err,
			"P2PNode.BroadcastMessage() mock peers should not fail to decode the received broadcast message",
		)

		protoMsg := &types.P2PMessage{}
		err = proto.Unmarshal(decodedData, protoMsg)

		assert.Nil(
			t,
			err,
			"P2PNode.BroadcastMessage() mock peers should not fail to unmarshal the received broadcast message",
		)

		assert.True(
			t,
			AreProtoMessagesEqual(msg, protoMsg),
			"P2PNode.BroadcastMessage() mock peers should receive the broadcast message uncorrupted",
		)
	}

	Teardown_TestP2PNode_Broadcast(node, peers, listeners)
}

// this test simulates the scenario where a node receives a broadcast message from a given level (say=2)
// and asserts that this node has successfully re-broadcasted the message.
func TestP2PNode_HandleBroadcastMessage(t *testing.T) {
	// setup
	node, peers, listeners := Setup_TestP2PNode_Broadcast()

	go node.Handle()

	// track impact for connections
	wg := sync.WaitGroup{}
	rw := sync.RWMutex{}
	impact := map[int]bool{}
	data := map[int][]byte{}
	for _, peer := range peers {
		wg.Add(1)
		go func(c net.Conn, m map[int]bool, d map[int][]byte) {
			data := make([]byte, 1024)
			c.SetReadDeadline(time.Now().Add(time.Millisecond * 2000))
			n, err := c.Read(data)
			if err == nil {
				ID := c.LocalAddr().(*net.TCPAddr).Port - 10000
				rw.Lock()
				m[ID] = true
				d[ID] = data[:n]
				rw.Unlock()
			}
			wg.Done()
		}(peer, impact, data)
	}

	// broadcast
	msg := types.NewP2PMessage(0, 2, "", node.config["address"].(string), &commonTypes.PocketEvent{
		Topic: commonTypes.PocketTopic_CONSENSUS_MESSAGE_TOPIC,
	})
	msg.MarkAsBroadcastMessage()

	msgBytes, err := proto.Marshal(msg)
	if err != nil {
		log.Fatal("TestP2PNode_HandleBroadcastMessage: mock peer failed to marshal the broadcast message before send")
	}
	encodedMsg := node.encode(false, 0, msgBytes, true)

	go peers[5].Write(encodedMsg) // these peers are already connected to the p2p node

	wg.Wait()

	expectedImpact := GetRainTreeExpectedImpactAtLevel(1, 27, 2)

	// verify impact on the subset of peers
	assert.Equal(
		t,
		expectedImpact,
		impact,
		"P2PNode.BroadcastMessage() should send the message to all of the expected peers",
	)

	for k, _ := range data {
		_, decodedData, _, err := node.decode(data[k])

		assert.Nil(
			t,
			err,
			"P2PNode.BroadcastMessage() mock peers should not fail to decode the received broadcast message",
		)

		protoMsg := &types.P2PMessage{}
		err = proto.Unmarshal(decodedData, protoMsg)

		assert.Nil(
			t,
			err,
			"P2PNode.BroadcastMessage() mock peers should not fail to unmarshal the received broadcast message",
		)

		assert.True(
			t,
			AreProtoMessagesEqual(msg, protoMsg),
			"P2PNode.BroadcastMessage() mock peers should receive the broadcast message uncorrupted",
		)
	}

	Teardown_TestP2PNode_Broadcast(node, peers, listeners)
}

func Teardown_TestP2PNode_Broadcast(node *p2pNode, peers []net.Conn, listeners []net.Listener) {
	// close all connections
	for _, peer := range peers {
		peer.Close()
	}

	// close all listeners
	for _, listener := range listeners {
		listener.Close()
	}

	// stop the node
	node.Stop()
}

func Setup_TestP2PNode_BroadcastMessage_Integration() []*p2pNode {
	// generate the peer list
	peers := []string{}
	for i := 0; i < 27; i++ {
		peers = append(peers, fmt.Sprintf("%d@127.0.0.1:%d", i+1, i+1+10000))
	}

	// create p2p nodes
	nodes := make([]*p2pNode, 0)
	for i := 0; i < 27; i++ {
		config := map[string]interface{}{
			"address":         fmt.Sprintf("127.0.0.1:%d", i+1+10000),
			"readBufferSize":  1024,
			"writeBufferSize": 1024,
			"peers":           peers,
			"redundancy":      false,
			"id":              i + 1,
		}

		node := NewP2PNode(config)

		node.ID = i + 1
		nodes = append(nodes, node)
	}

	// launch the p2p nodes
	for i, node := range nodes {
		err := node.Start()
		if err != nil {
			log.Fatalf("Setup_TestP2PNode_Broadcast() failed to start node #%d: reason=%s", i, err)
		}
		go node.Handle()
	}

	// connect nodes in between each other
	for i := 0; i < 27; i++ {
		for j := 0; j < 27; j++ {
			if i != j {
				nodes[i].Dial(nodes[j].Address())
			}
		}
	}

	return nodes
}
func TestP2PNode_BroadcastMessage_Integration(t *testing.T) {
	// setup
	nodes := Setup_TestP2PNode_BroadcastMessage_Integration()

	// track broadcast impact
	wg := sync.WaitGroup{}
	rw := sync.RWMutex{}
	impact := map[int][]bool{}
	data := map[int]*types.P2PMessage{}

	// setup the expectation
	expectedImpact := GetRainTreeExpectedImpactAtAllLevels(1, 27)

	// setup the waitgroup to wait for the expected nodes to receive the broadcast message
	// not all nodes in the list will receive the broadcast message (no cleanup or redundancy layer is implemented yet)
	for _, node := range nodes {
		if _, exists := expectedImpact[node.ID]; exists {
			for i := 0; i < len(expectedImpact[node.ID]); i++ {
				wg.Add(1)
				go func(node *p2pNode) {
					node.OnNewMessage(func(msg *types.P2PMessage) {
						rw.Lock()
						if _, ok := impact[node.ID]; !ok {
							impact[node.ID] = make([]bool, 0)
						}
						impact[node.ID] = append(impact[node.ID], true)
						data[node.ID] = msg
						rw.Unlock()
						wg.Done()
					})
				}(node)
			}
		}
	}

	// suppress logs
	for _, node := range nodes {
		node.Suppress(true)
	}

	// pick an originator
	originId := 0
	originNode := nodes[originId]

	// broadcast
	msg := types.NewP2PMessage(0, 4, originNode.config["address"].(string), "", &commonTypes.PocketEvent{
		Topic: commonTypes.PocketTopic_CONSENSUS_MESSAGE_TOPIC,
	})
	msg.MarkAsBroadcastMessage()
	go originNode.BroadcastMessage(msg, true, 0)

	wg.Wait()

	assert.Equal(
		t,
		expectedImpact,
		impact,
		"P2PNode.BroadcastMessage() should send the message to all of the expected peers",
	)

	// assert that all nodes have received the broadcast message
	for k := range expectedImpact {
		assert.Truef(
			t,
			AreProtoMessagesEqual(msg, data[k]),
			"P2PNode.BroadcastMessage() should have received a broadcast message from node %d", k,
		)
	}

	Teardown_TestP2PNode_BroadcastMessage_Integration(nodes)
}

func Teardown_TestP2PNode_BroadcastMessage_Integration(nodes []*p2pNode) {
	// close all connections
	for _, node := range nodes {
		node.Stop()
	}
}

func GetRainTreeExpectedImpact(root int, listSize int) map[int]bool {
	list := make([]peerInfo, 0)
	for i := 0; i < listSize; i++ {
		list = append(list, peerInfo{
			ID:      i + 1,
			address: "127.0.0.1:" + strconv.Itoa(10000+i+1),
		})
	}

	tree := NewRainTree()
	tree.SetLeafs(list)
	tree.SetRoot(root)

	impact := map[int]bool{}
	tree.Traverse(
		true,
		0,
		func(originatorId int, left, right peerInfo, currentLevel int) error {
			impact[left.ID] = true
			impact[right.ID] = true
			return nil
		})

	return impact
}

func GetRainTreeExpectedImpactAtLevel(root, listSize, level int) map[int]bool {
	list := make([]peerInfo, 0)
	for i := 0; i < listSize; i++ {
		list = append(list, peerInfo{
			ID:      i + 1,
			address: "127.0.0.1:" + strconv.Itoa(10000+i+1),
		})
	}

	tree := NewRainTree()
	tree.SetLeafs(list)
	tree.SetRoot(root)

	impact := map[int]bool{}
	tree.Traverse(
		false,
		level,
		func(originatorId int, left, right peerInfo, currentLevel int) error {
			impact[left.ID] = true
			impact[right.ID] = true
			return nil
		})

	return impact
}

func GetRainTreeExpectedImpactAtAllLevels(root int, listSize int) map[int][]bool {
	list := make([]peerInfo, 0)
	for i := 0; i < listSize; i++ {
		list = append(list, peerInfo{
			ID:      i + 1,
			address: "127.0.0.1:" + strconv.Itoa(10000+i+1),
		})
	}

	tree := NewRainTree()
	tree.SetLeafs(list)
	tree.SetRoot(root)
	var rw sync.RWMutex

	peermap := map[int][]struct {
		l int
		r int
	}{}

	for _, id := range tree.GetSortedList() {
		peermap[id] = make([]struct {
			l int
			r int
		}, 0)
	}

	queue := make([]struct {
		id          int
		level       int
		root        bool
		contactedby int
	}, 0)

	addtopeermap := func(id, l, r int) {
		peermap[id] = append(peermap[id], struct {
			l int
			r int
		}{l, r})
	}
	queuein := func(id int, level int, root bool, contactedby int) {
		queue = append(queue, struct {
			id          int
			level       int
			root        bool
			contactedby int
		}{id, level, root, contactedby})
	}
	queuepop := func() struct {
		id          int
		level       int
		root        bool
		contactedby int
	} {
		popped := queue[0]
		queue = queue[1:]
		return popped
	}

	impact := map[int][]bool{}
	act := func(originator int, l, r peerInfo, currentlevel int) error {
		defer rw.Unlock()
		rw.Lock()

		lid := l.ID
		rid := r.ID

		addtopeermap(originator, lid, rid)
		queuein(lid, currentlevel, false, originator)
		queuein(rid, currentlevel, false, originator)

		if _, ok := impact[lid]; !ok {
			impact[lid] = make([]bool, 0)
		}

		if _, ok := impact[rid]; !ok {
			impact[rid] = make([]bool, 0)
		}

		impact[lid] = append(impact[lid], true)
		impact[rid] = append(impact[rid], true)

		return nil
	}

	queuein(int(root), 3, true, 0)

	for {
		currentpeer := queuepop()
		tree.SetRoot(currentpeer.id)
		tree.Traverse(currentpeer.root, currentpeer.level, act)
		if len(queue) == 0 {
			break
		}
	}

	return impact
}

func AreProtoMessagesEqual(msgA *types.P2PMessage, msgB *types.P2PMessage) bool {
	if msgA.Metadata.Source != msgB.Metadata.Source {
		return false
	}

	if msgA.Metadata.Destination != msgB.Metadata.Destination {
		return false
	}

	if msgA.Metadata.Level != msgB.Metadata.Level {
		return false
	}

	if msgA.Payload.Topic != msgB.Payload.Topic {
		return false
	}

	if strings.Compare(msgA.Payload.Data.String(), msgB.Payload.Data.String()) != 0 {
		return false
	}

	return true
}
