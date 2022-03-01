package p2p

import (
	"pocket/p2p/types"
	"strings"
)

func (g *networkModule) configure(protocol, address, external string, peers []string) {
	g.protocol = protocol
	g.address = address
	g.externaladdr = external
	g.peerlist = types.NewPeerlist()

	// this is a hack to get going no more no less
	for i, p := range peers {
		peer := types.NewPeer(uint64(i+1), p)
		port := strings.Split(peer.Addr(), ":")
		myport := strings.Split(g.address, ":")
		if port[1] == myport[1] {
			g.id = peer.Id()
		}
		g.peerlist.Add(*peer)
	}
}

func (g *networkModule) init() error {
	msg := Message(int32(0), 1, types.PocketTopic_P2P_PING, "", "")
	_, err := g.c.register(*msg, Encode, Decode)
	if err != nil {
		return err
	}

	return nil
}

func (g *networkModule) isReady() <-chan uint {
	return g.ready
}

func (g *networkModule) close() {
	g.done <- 1
	g.closed <- 1
	g.listening.Store(false)
	g.listener.Close()
	close(g.done)
}

func (g *networkModule) finished() <-chan uint {
	return g.closed
}

func (g *networkModule) error(err error) {
	defer g.err.Unlock()
	g.err.Lock()

	if g.err.error != nil {
		g.err.error = err
	}

	g.errored <- 1
}
