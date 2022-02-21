package p2p

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"pocket/shared/types"
	"strconv"
	"strings"

	"github.com/golang/protobuf/proto"
)

/*
 @ Domain Codec Interface
*/

type Topic string
type Action string

type message struct {
	nonce uint32

	source      string
	destination string

	level uint16

	topic   Topic
	action  Action
	payload interface{}
}

func (m *message) IsRequest() bool {
	return m.nonce != uint32(0)
}

func encode(m message) ([]byte, error) {
	src := parseipstring(m.source)
	dst := parseipstring(m.destination)

	tb := []byte(m.topic)
	ab := []byte(m.action)

	nb := make([]byte, 4) // nonce
	lb := make([]byte, 2) // level

	tl := make([]byte, 2) // topic length
	al := make([]byte, 2) // action length

	binary.BigEndian.PutUint32(nb, m.nonce)
	binary.BigEndian.PutUint16(lb, m.level)

	binary.BigEndian.PutUint16(tl, uint16(len(tb)))
	binary.BigEndian.PutUint16(al, uint16(len(ab)))

	buff := make([]byte, 0)

	buff = append(buff, src...)
	buff = append(buff, dst...)
	buff = append(buff, nb...)
	buff = append(buff, lb...)
	buff = append(buff, tl...)
	buff = append(buff, al...)
	buff = append(buff, tb...)
	buff = append(buff, ab...)

	return buff, nil
}

func decode(c interface {
	message(uint32, Action, uint16, ...string) message
}, data []byte) (message, error) {
	meta := data[:12]
	header := data[12 : 12+2+8] // 2 for the level, 8 for topic and action lengths
	body := data[12+2+8:]

	src := parseipbytes(meta[:6])
	dst := parseipbytes(meta[6:])

	nonce := binary.BigEndian.Uint32(header[:4])
	level := binary.BigEndian.Uint16(header[4:6])

	tl := binary.BigEndian.Uint16(header[6:8])
	al := binary.BigEndian.Uint16(header[8:10])

	t := body[:tl]
	a := body[tl : tl+al]

	m := c.message(nonce, Action(string(a)), level, src, dst)
	m.topic = Topic(t)

	return m, nil
}

/*
 @
 @ x.x.x.x:yyyyy
 @ each x is represented on 8bits (1 byte)
 @ the yyyyy port max len is 6 and is represented on 16bits (2 bytes)
 @
*/
func parseipstring(ip string) []byte {
	buff := bytes.NewBuffer(nil)
	parts := strings.FieldsFunc(ip, func(r rune) bool { return r == '.' || r == ':' })

	address := parts[:4]
	for _, p := range address {
		pi, _ := strconv.Atoi(p)
		binary.Write(buff, binary.BigEndian, uint8(pi))
	}

	port, _ := strconv.Atoi(parts[4])
	binary.Write(buff, binary.BigEndian, uint16(port))

	return buff.Bytes()
}

func parseipbytes(buff []byte) string {
	var port uint16
	ip := make([]uint8, 4)

	ipbuff := bytes.NewBuffer(buff[:4])
	portbuff := bytes.NewBuffer(buff[4:])

	binary.Read(ipbuff, binary.BigEndian, ip)
	binary.Read(portbuff, binary.BigEndian, &port)

	return fmt.Sprintf("%d.%d.%d.%d:%d", ip[0], ip[1], ip[2], ip[3], port)
}

/*
 @@@@@@@@@@@@@
 @ Each message has a domain to which it belong.
 @ that domain may have one or many topics
 @ Each domain is represented by an empty structure that implements:
 @ message() message
 @ encode(message) ([]byte, error)
 @ decode([]byte) (message, error)
 @
 @ This structure is named: Messenger
 @@@@@@@@@@@@@
*/

/*
 @
 @ Churn Management Messenger
 @
*/
var (
	Churn Topic = "churn"

	Ping  Action = "ping"
	Pong  Action = "pong"
	Join  Action = "join"
	Leave Action = "leave"
)

type churnmgmt struct{}

func (c *churnmgmt) message(nonce uint32, a Action, level uint16, srcndest ...string) message {
	m := message{
		nonce:       nonce,
		topic:       Churn,
		action:      a,
		payload:     nil,
		source:      "",
		destination: "",
		level:       level,
	}

	if len(srcndest) == 2 {
		m.source = srcndest[0]
		m.destination = srcndest[1]
	}

	return m
}

func (c *churnmgmt) encode(m message) ([]byte, error) {
	return encode(m)
}

func (c *churnmgmt) decode(payload []byte) (message, error) {
	return decode(c, payload)
}

/*
 @
 @ Gossip Messenger
 @
*/

var (
	Blockchain Topic = "blockchain"

	Gossip       Action = "gossip"
	GossipACK    Action = "gossip_ack"
	GossipRESEND Action = "gossip_resend"
)

type gossip struct{}

func (g *gossip) message(nonce uint32, a Action, level uint16, srcndest ...string) message {
	m := message{
		nonce:       nonce,
		topic:       Blockchain,
		action:      a,
		payload:     nil,
		source:      "",
		destination: "",
		level:       level,
	}

	if len(srcndest) == 2 {
		m.source = srcndest[0]
		m.destination = srcndest[1]
	}

	return m
}
func (g *gossip) encode(m message) ([]byte, error) {
	return encode(m)
}

func (g *gossip) decode(payload []byte) (message, error) {
	return decode(g, payload)
}

func GossipMessage(addr string) message {
	return (&gossip{}).message(0, Gossip, 4, addr, "0.0.0.0:02023")
}

/*
 @
 @ Protobuf messages
 @
*/

// protobuff domain codec, just a wrapper on top of protobuff
type pbuff struct{}

func (c *pbuff) message(nonce int32, level int32, topic types.PocketTopic, src, dest string) *types.NetworkMessage {
	return &types.NetworkMessage{
		Level:       level,
		Nonce:       nonce,
		Topic:       topic,
		Source:      src,
		Destination: dest,
	}
}

func (c *pbuff) encode(m types.NetworkMessage) ([]byte, error) {
	data, err := proto.Marshal(&m)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (c *pbuff) decode(data []byte) (types.NetworkMessage, error) {
	msg := &types.NetworkMessage{}
	err := proto.Unmarshal(data, msg)
	if err != nil {
		return types.NetworkMessage{Nonce: -1, Level: -1}, err
	}
	return *msg, nil
}
