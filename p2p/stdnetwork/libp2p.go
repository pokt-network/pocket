package stdnetwork

import (
	"context"
	"fmt"
	"log"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"

	"github.com/pokt-network/pocket/p2p/common"
	"github.com/pokt-network/pocket/p2p/libp2p/identity"
	"github.com/pokt-network/pocket/p2p/libp2p/protocol"
	"github.com/pokt-network/pocket/p2p/providers"
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
	Year       = time.Hour * 24 * 365
	// TODO: consider more carefully and parameterize.
	DefaultPeerTTL = 2 * Year
)

// TODO: factor out args which are common to network
// implementations to an options or config struct.
func NewLibp2pNetwork(
	bus modules.Bus,
	addrBookProvider providers.AddrBookProvider,
	currentHeightProvider providers.CurrentHeightProvider,
	host host.Host,
	topic *pubsub.Topic,
) (types.Network, error) {
	// TODO: receive ctx in interface methods (?)
	//ctx := context.Background()

	addrBook, err := addrBookProvider.GetStakedAddrBookAtHeight(currentHeightProvider.CurrentHeight())
	if err != nil {
		log.Fatalf("%s", ErrNetwork("getting staked address book", err))
	}

	addrBookMap := make(types.AddrBookMap)
	for _, poktPeer := range addrBook {
		// TODO: (?)
		// NB: don't add self to address book.
		//if peer.ID == host.ID() {
		//	continue
		//}

		addrBookMap[poktPeer.Address.String()] = poktPeer
		pubKey, err := identity.PubKeyFromPoktPeer(poktPeer)
		if err != nil {
			return nil, ErrNetwork("converting peer public key", err)
		}
		peer, err := identity.PeerAddrInfoFromPoktPeer(poktPeer)
		if err != nil {
			return nil, ErrNetwork("converting peer info", err)
		}

		host.Peerstore().AddAddrs(peer.ID, peer.Addrs, DefaultPeerTTL)
		if err := host.Peerstore().AddPubKey(peer.ID, pubKey); err != nil {
			return nil, ErrNetwork("", err)
		}

		//log.Printf("stdnetwork/libp2p.go:52 | peer: %s", peer)
	}

	return &libp2pNetwork{
		// TODO: is it unconventional to set bus here?
		bus:         bus,
		host:        host,
		topic:       topic,
		addrBookMap: addrBookMap,
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
	selfPoktAddr, err := p2pNet.GetBus().GetP2PModule().GetAddress()
	if err != nil {
		return ErrNetwork(fmt.Sprintf(
			"sending to poktAddr: %s", poktAddr,
		), err)
	}

	// NB: don't send to self.
	if selfPoktAddr.Equals(poktAddr) {
		return nil
	}

	poktPeer, ok := p2pNet.addrBookMap[poktAddr.String()]
	if !ok {
		// NB: this should not happen.
		return ErrNetwork("peer not found in address book", fmt.Errorf(
			"peer address: %s", poktAddr,
		))
	}

	peerAddrInfo, err := identity.PeerAddrInfoFromPoktPeer(poktPeer)
	if err != nil {
		return ErrNetwork("parsing peer multiaddr", err)
	}

	// TODO: add context to interface methods.
	ctx := context.Background()
	//log.Printf("stdnetwork/libp2p.go:99 | peer: %s", peerAddrInfo)
	stream, err := p2pNet.host.NewStream(ctx, peerAddrInfo.ID, protocol.PoktProtocolID)
	//log.Printf("stdnetwork/libp2p.go:103 | peerAddrInfo: %s", peerAddrInfo)
	if err != nil {
		return ErrNetwork(fmt.Sprintf(
			"opening a stream (peer address %s)", poktAddr,
		), err)
	}

	if _, err := stream.Write(data); err != nil {
		return ErrNetwork(fmt.Sprintf(
			"writing to stream (peer address: %s)", poktAddr,
		), err)
	}
	defer func() {
		if err := stream.Close(); err != nil {
			// TODO: conventional error handling.
			log.Printf("%v", ErrNetwork("closing stream", err))
		}
	}()

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
	for _, poktPeer := range p2pNet.addrBookMap {
		addrBook = append(addrBook, poktPeer)
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
