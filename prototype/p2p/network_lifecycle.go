package p2p

import (
	"pocket/p2p/types"
	"strings"
)

func (m *networkModule) configure(protocol, address, external string, peers []string) {
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

func (m *networkModule) init() error {
	msg := Message(int32(0), 1, types.PocketTopic_P2P_PING, "", "")
	_, err := m.c.register(*msg, Encode, Decode)
	if err != nil {
		return err
	}

	return nil
}

func (m *networkModule) isReady() <-chan uint {
	return m.ready
}

func (m *networkModule) close() {
	m.done <- 1
	m.closed <- 1
	m.isListening.Store(false)
	m.listener.Close()
	close(m.done)
}

func (m *networkModule) finished() <-chan uint {
	return m.closed
}

func (m *networkModule) error(err error) {
	defer m.err.Unlock()
	m.err.Lock()

	if m.err.error != nil {
		m.err.error = err
	}

	m.errored <- 1
}
