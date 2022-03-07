package p2p

import (
	"strings"

	"github.com/pokt-network/pocket/p2p/types"
)

func (m *p2pModule) configure(protocol, address, external string, peers []string) {
	m.protocol = protocol
	m.address = address
	m.externaladdr = external
	m.peerlist = types.NewPeerlist()

	// TODO(derrandz): this is a hack to get going no more no less
	// This hack tries to achieve addressbook injection behavior
	// it basically takes a slice of ips and turns them into peers
	// and generating consecutive uints for them from i to length(peers)
	// while also self-filtering.
	// this is filler code until the expected behavior is well spec'd
	for i, p := range peers {
		peer := types.NewPeer(uint64(i+1), p)

		peerAddrParts := strings.Split(peer.Addr(), ":")
		myAddrParts := strings.Split(m.address, ":")

		peerPort, myPort := peerAddrParts[1], myAddrParts[1]
		if peerPort == myPort {
			m.id = peer.Id()
		}
		m.peerlist.Add(*peer)
	}
}

func (m *p2pModule) init() error {
	p2pmsg := types.P2PMessage{}
	_, err := m.c.Register(p2pmsg, types.Encode, types.Decode)
	if err != nil {
		return err
	}

	return nil
}

func (m *p2pModule) isReady() <-chan uint {
	return m.ready
}

func (m *p2pModule) close() {
	m.done <- 1
	m.closed <- 1
	m.isListening.Store(false)
	m.listener.Close()
	close(m.done)
}

func (m *p2pModule) finished() <-chan uint {
	return m.closed
}

func (m *p2pModule) error(err error) {
	defer m.err.Unlock()
	m.err.Lock()

	if m.err.error != nil {
		m.err.error = err
	}

	m.errored <- 1
}
