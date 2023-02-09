package stdnetwork

import (
	"context"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"

	"github.com/pokt-network/pocket/p2p/common"
	"github.com/pokt-network/pocket/p2p/types"
	poktCrypto "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
)

type libp2pNetwork struct {
	bus   modules.Bus
	host  host.Host
	topic *pubsub.Topic
}

var (
	ErrNetwork = common.NewErrFactory("LibP2P network error")
)

func NewLibp2pNetwork(bus modules.Bus, host host.Host, topic *pubsub.Topic) (types.Network, error) {
	return &libp2pNetwork{
		// TODO: is it unconventional to set bus here?
		bus:   bus,
		host:  host,
		topic: topic,
	}, nil
}

func (p2pNet *libp2pNetwork) NetworkBroadcast(data []byte) error {
	// TODO: receive context in interface methods?
	ctx := context.Background()

	// Routed send using pubsub
	if err := p2pNet.topic.Publish(ctx, data); err != nil {
		return ErrNetwork("unable to publish to topic", err)
	}
	return nil
}

func (p2pNet *libp2pNetwork) NetworkSend(data []byte, poktAddr poktCrypto.Address) error {
	// Direct send using host
	//TODO: ...
}

// This function was added to specifically support the RainTree implementation.
// Handles the raw data received from the network and returns the data to be processed
// by the application layer.
func (p2pNet *libp2pNetwork) HandleNetworkData(data []byte) ([]byte, error) {
	return data, nil
}

func (p2pNet *libp2pNetwork) GetAddrBook() types.AddrBook {
	return nil
}

func (p2pNet *libp2pNetwork) AddPeerToAddrBook(peer *types.NetworkPeer) error {
	return nil
}

// TODO: implement (?)
// TODO: fix typo in interface (?)
func (p2pNet *libp2pNetwork) RemovePeerToAddrBook(peer *types.NetworkPeer) error {
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
