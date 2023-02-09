package stdnetwork

import (
	"context"
	"fmt"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"

	"github.com/pokt-network/pocket/p2p/common"
	"github.com/pokt-network/pocket/p2p/libp2p/identity"
	"github.com/pokt-network/pocket/p2p/types"
	poktCrypto "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
)

type libp2pNetwork struct {
	bus         modules.Bus
	host        host.Host
	topic       *pubsub.Topic
	addrBookMap types.AddrBookMap
}

var (
	ErrNetwork = common.NewErrFactory("LibP2P network error")
)

func NewLibp2pNetwork(bus modules.Bus, host host.Host, topic *pubsub.Topic) (types.Network, error) {
	return &libp2pNetwork{
		// TODO: is it unconventional to set bus here?
		bus:         bus,
		host:        host,
		topic:       topic,
		addrBookMap: make(types.AddrBookMap),
	}, nil
}

// NetworkBroadcast uses the configured pubsub router to broadcast data to peers.
func (p2pNet *libp2pNetwork) NetworkBroadcast(data []byte) error {
	// TODO: receive context in interface methods?
	ctx := context.Background()

	// NB: Routed send using pubsub
	if err := p2pNet.topic.Publish(ctx, data); err != nil {
		return ErrNetwork("unable to publish to topic", err)
	}
	return nil
}

// NetworkSend connects sends data directly to the specified peer.
func (p2pNet *libp2pNetwork) NetworkSend(data []byte, poktAddr poktCrypto.Address) error {
	poktPeer, ok := p2pNet.addrBookMap[poktAddr.String()]
	if !ok {
		// NB: this should not happen.
		return ErrNetwork("peer not found in address book", fmt.Errorf(
			"peer address: %s", poktAddr,
		))
	}

	peerAddrInfo, err := identity.PeerAddrInfoFromPoktPeer(poktPeer)
	if err != nil {
		return ErrNetwork("unable to parse peer multiaddr", err)
	}

	// TODO: add context to interface methods.
	ctx := context.Background()
	stream, err := p2pNet.host.NewStream(ctx, peerAddrInfo.ID)
	if err != nil {
		return ErrNetwork(fmt.Sprintf(
			"unable to open a stream to peer with pokt address %q", poktAddr,
		), err)
	}

	if _, err := stream.Write(data); err != nil {
		return ErrNetwork(fmt.Sprintf(
			"unable to write to stream (peer address: %s)", poktAddr,
		), err)
	}

	// TODO: find a more conventional way to send messages directly
	// so that we don't have to close and re-open per direct message.
	// NB: close the stream so that peer receives EOF.
	if err := stream.Close(); err != nil {
		return ErrNetwork(fmt.Sprintf(
			"unable to close stream with peer with address %q", poktAddr,
		), err)
	}
	return nil
}

// This function was added to specifically support the RainTree implementation.
// Handles the raw data received from the network and returns the data to be processed
// by the application layer.
func (p2pNet *libp2pNetwork) HandleNetworkData(data []byte) ([]byte, error) {
	return data, nil
}

func (p2pNet *libp2pNetwork) GetAddrBook() types.AddrBook {
	addrBook := make(types.AddrBook, 0)
	for _, p := range p2pNet.addrBookMap {
		addrBook = append(addrBook, p)
	}
	return addrBook
}

func (p2pNet *libp2pNetwork) AddPeerToAddrBook(peer *types.NetworkPeer) error {
	p2pNet.addrBookMap[peer.Address.String()] = peer
	return nil
}

// TODO: fix typo in interface (?)
func (p2pNet *libp2pNetwork) RemovePeerToAddrBook(peer *types.NetworkPeer) error {
	delete(p2pNet.addrBookMap, peer.Address.String())
	return nil
}

func (p2pNet *libp2pNetwork) GetBus() modules.Bus {
	return p2pNet.bus
}

func (p2pNet *libp2pNetwork) SetBus(bus modules.Bus) {
	p2pNet.bus = bus
}

func (p2pNet *libp2pNetwork) Close() error {
	return p2pNet.host.Close()
}
