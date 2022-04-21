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
		msg.Metadata.Broadcast = true // forcing function

		return m.dichotomicAckSend(l, r, msg)
	}

	fmt.Println("id", m.id)
	fmt.Println("list", list)
	fmt.Println("current level:", currentLevel)
	// panic(nil)
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

// handles the received broadcast message
// TODO(derrandz): he need for the sourceAddr (even though the p2pmsg has a source field) stems from the inbound/outbound differentiation in address
// this will be resolved with nodeIds used to index the pooled connection as opposed to temporary ip addresses
func (m *p2pModule) handleBroadcast(nonce uint32, sourceAddr string, msg *types.P2PMessage) error {
	// todo send ack
	ackMsg := &types.P2PAckMessage{
		Acker: m.address,
		Ackee: sourceAddr,
	}
	encodedAck, encErr := m.c.Marshal(ackMsg)
	if encErr != nil {
		return ErrFailedToAckBroadcast(encErr)
	}

	err := m.respond(nonce, false, sourceAddr, encodedAck, true)
	if err != nil {
		return ErrFailedToAckBroadcast(err)
	}

	return m.broadcast(msg, false)
}
