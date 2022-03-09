package p2p

import (
	"fmt"
	"strings"
	"sync"

	"github.com/pokt-network/pocket/p2p/types"
)

func (m *p2pModule) encodeTargetedMessage(target *types.Peer, msg *types.P2PMessage) ([]byte, error) {
	tmsg := msg.Copy()
	tmsg.Metadata.Destination = target.Addr()
	msgEnc, lerr := m.c.Marshal(tmsg)
	return msgEnc, lerr
}

func (m *p2pModule) ackSend(addr string, data []byte) error {
	var sourceAddr string = m.externaladdr
	ack, rerr := m.request(addr, data, true)

	if rerr != nil {
		return rerr
	}

	ackMsg := &types.P2PAckMessage{}
	derr := m.c.Unmarshal(ack, ackMsg)
	if derr != nil {
		return derr
	}

	if strings.Compare(ackMsg.Acker, addr) != 0 || strings.Compare(ackMsg.Ackee, sourceAddr) != 0 {
		return fmt.Errorf("Unrecognized ACK from %s", addr) // TODO(derrandz): put in errors.go
	}

	return nil
}

func (m *p2pModule) targetedAckSend(t *types.Peer, msg *types.P2PMessage, done func(error)) {
	var err error
	var enc []byte

	enc, err = m.encodeTargetedMessage(t, msg)

	if err == nil {
		err = m.ackSend(t.Addr(), enc)
	}

	done(err)
}

func (m *p2pModule) dichotomicAckSend(l, r *types.Peer, msg *types.P2PMessage) error {
	var wg sync.WaitGroup
	var rwx sync.RWMutex

	var targets []*types.Peer = []*types.Peer{l, r}
	var errors []error = []error{}

	done := func(err error) {
		defer func() {
			rwx.Unlock()
			wg.Done()
		}()

		rwx.Lock()
		errors = append(errors, err)
	}

	for _, t := range targets {
		wg.Add(1)
		go m.targetedAckSend(t, msg, done)
	}

	wg.Wait()

	for _, err := range errors {
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *p2pModule) broadcast(msg *types.P2PMessage, isRoot bool) error {
	var list *types.Peerlist = m.peerlist

	var topLevel int = int(getTopLevel(list))
	var currentLevel int = topLevel - 1

	var sourceAddr string = m.externaladdr

	// TODO(Derrandz): m.config.Address
	if msg.Metadata.Level == int32(topLevel) && msg.Metadata.Source == "" {
		isRoot = true
	} else {
		currentLevel = int(msg.Metadata.Level)
	}

	// Sends to left and right peers and expects to receive an ACK or fails otherwise
	onRainDrop := func(id uint64, l, r *types.Peer, itCurrentLevel int) error {
		msg.Metadata.Level = int32(itCurrentLevel)
		msg.Metadata.Source = sourceAddr

		return m.dichotomicAckSend(l, r, msg)
	}

	err := rain(
		m.id,
		list,
		onRainDrop,
		isRoot,
		currentLevel,
	)

	if err != nil {
		return err
	}

	for _, handler := range m.handlers[types.BroadcastDoneEvent] {
		handler(m)
	}

	return nil
}

func (m *p2pModule) handle() {
	// var msg *types.P2PMessage
	// var mx sync.Mutex

	// m.log("Handling...")
	// for w := range m.sink {
	// 	nonce, data, srcaddr, encoded := (&w).Implode()

	// 	if encoded {
	// 		mx.Lock()
	// 		var msg
	// 		err := m.c.Unmarshal(data)
	// 		if err != nil {
	// 			m.log("Error decoding data", err.Error())
	// 			continue
	// 		}
	// 		msgi := decoded.(*types.P2PMessage)
	// 		msg = msgi
	// 		msg.Metadata.Nonce = int32(nonce)
	// 		mx.Unlock()
	// 	} else {
	// 		msg.Payload.Data = &anypb.Any{}
	// 		msg.Metadata.Nonce = int32(nonce)
	// 		msg.Metadata.Source = srcaddr
	// 	}

	// 	switch msg.Payload.Topic {

	// 	case shared.PocketTopic_CONSENSUS_MESSAGE_TOPIC:
	// 		mx.Lock()
	// 		md := &types.Metadata{
	// 			Nonce:       msg.Metadata.Nonce,
	// 			Level:       msg.Metadata.Level,
	// 			Source:      m.externaladdr,
	// 			Destination: msg.Metadata.Source,
	// 		}
	// 		pl := &shared.PocketEvent{
	// 			Topic: shared.PocketTopic_CONSENSUS_MESSAGE_TOPIC,
	// 		}
	// 		ack := &types.P2PMessage{Metadata: md, Payload: pl}
	// 		encoded, err := m.c.Marshal(ack)
	// 		if err != nil {
	// 			m.log("Error encoding m for gossipaCK", err.Error())
	// 		}

	// 		err = m.respond(uint32(msg.Metadata.Nonce), false, srcaddr, encoded, true)
	// 		if err != nil {
	// 			m.log("Error encoding msg for gossipaCK", err.Error())
	// 		}

	// 		mx.Unlock()

	// 		m.log("Acked to", ack.Metadata.Destination)

	// 		go m.broadcast(msg, false)

	// 	default:
	// 		m.log("Unrecognized message topic", msg.Payload.Topic, "from", msg.Metadata.Source, "to", msg.Metadata.Destination, "@node", m.address)
	// 	}
	// }
}
