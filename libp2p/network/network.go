package network

import (
	"context"
	"fmt"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	libp2pHost "github.com/libp2p/go-libp2p/core/host"

	"github.com/pokt-network/pocket/libp2p/identity"
	"github.com/pokt-network/pocket/libp2p/protocol"
	typesLibp2p "github.com/pokt-network/pocket/libp2p/types"
	"github.com/pokt-network/pocket/p2p/providers"
	"github.com/pokt-network/pocket/p2p/providers/addrbook_provider"
	persABP "github.com/pokt-network/pocket/p2p/providers/addrbook_provider/persistence"
	"github.com/pokt-network/pocket/p2p/providers/current_height_provider"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
)

type libp2pNetwork struct {
	base_modules.IntegratableModule

	logger *modules.Logger
	//nolint:unused // bus is used by embedded base module(s)
	bus         modules.Bus
	host        libp2pHost.Host
	topic       *pubsub.Topic
	addrBookMap typesP2P.AddrBookMap
}

const (
	year = time.Hour * 24 * 365
	// TECHDEBT: consider more carefully and parameterize.
	defaultPeerTTL = 2 * year
)

var (
	// ErrNetwork wraps errors which occur within the libp2pNetwork implementation
	// Exported for testing purposes.
	ErrNetwork = typesLibp2p.NewErrFactory("libp2p network error")
)

// TECHDEBT: factor out args which are common to network
// implementations to an options or config struct.
func NewLibp2pNetwork(
	bus modules.Bus,
	logger *modules.Logger,
	host libp2pHost.Host,
	topic *pubsub.Topic,
) (typesP2P.Network, error) {
	p2pNet := &libp2pNetwork{
		logger: logger,
		host:   host,
		topic:  topic,
	}

	p2pNet.SetBus(bus)
	if err := p2pNet.setupAddrBookMap(); err != nil {
		return nil, err
	}

	return p2pNet, nil
}

// NetworkBroadcast uses the configured pubsub router to broadcast data to peers.
func (p2pNet *libp2pNetwork) NetworkBroadcast(data []byte) error {
	// IMPROVE: receive context in interface methods?
	ctx := context.Background()

	// NB: Routed send using pubsub
	if err := p2pNet.topic.Publish(ctx, data); err != nil {
		return ErrNetwork("unable to publish to topic", err)
	}
	return nil
}

// NetworkSend connects sends data directly to the specified peer.
func (p2pNet *libp2pNetwork) NetworkSend(data []byte, poktAddr crypto.Address) error {
	// IMPROVE: add context to interface methods.
	ctx := context.Background()

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

	peer, ok := p2pNet.addrBookMap[poktAddr.String()]
	if !ok {
		// NB: this should not happen.
		return ErrNetwork("", fmt.Errorf(
			"peer not found in address book, pokt address: %s", poktAddr,
		))
	}

	peerAddrInfo, err := identity.Libp2pAddrInfoFromPeer(peer)
	if err != nil {
		return ErrNetwork("parsing peer multiaddr", err)
	}

	stream, err := p2pNet.host.NewStream(ctx, peerAddrInfo.ID, protocol.PoktProtocolID)
	if err != nil {
		return ErrNetwork(fmt.Sprintf(
			"opening peer stream, pokt address: %s", poktAddr,
		), err)
	}

	if _, err := stream.Write(data); err != nil {
		return ErrNetwork(fmt.Sprintf(
			"writing to stream (peer address: %s)", poktAddr,
		), err)
	}
	defer func() {
		// NB: close the stream so that peer receives EOF.
		if err := stream.Close(); err != nil {
			p2pNet.logger.Error().Err(err).Msg(fmt.Sprintf(
				"closing peer stream, pokt address: %s", poktAddr,
			))
		}
	}()
	return nil
}

// This function was added to specifically support the RainTree implementation.
// Handles the raw data received from the network and returns the data to be processed
// by the application layer.
func (p2pNet *libp2pNetwork) HandleNetworkData(data []byte) ([]byte, error) {
	return data, nil
}

func (p2pNet *libp2pNetwork) GetAddrBook() typesP2P.AddrBook {
	addrBook := make(typesP2P.AddrBook, 0)
	for _, peer := range p2pNet.addrBookMap {
		addrBook = append(addrBook, peer)
	}
	return addrBook
}

func (p2pNet *libp2pNetwork) AddPeerToAddrBook(peer *typesP2P.NetworkPeer) error {
	p2pNet.addrBookMap[peer.Address.String()] = peer

	pubKey, err := identity.Libp2pPublicKeyFromPeer(peer)
	if err != nil {
		return ErrNetwork(fmt.Sprintf(
			"converting peer public key, pokt address: %s", peer.Address,
		), err)
	}
	libp2pPeer, err := identity.Libp2pAddrInfoFromPeer(peer)
	if err != nil {
		return ErrNetwork(fmt.Sprintf(
			"converting peer info, pokt address: %s", peer.Address,
		), err)
	}

	p2pNet.host.Peerstore().AddAddrs(libp2pPeer.ID, libp2pPeer.Addrs, defaultPeerTTL)
	if err := p2pNet.host.Peerstore().AddPubKey(libp2pPeer.ID, pubKey); err != nil {
		return ErrNetwork(fmt.Sprintf(
			"adding peer public key, pokt address: %s", peer.Address,
		), err)
	}
	return nil
}

func (p2pNet *libp2pNetwork) RemovePeerFromAddrBook(peer *typesP2P.NetworkPeer) error {
	delete(p2pNet.addrBookMap, peer.Address.String())

	libp2pPeer, err := identity.Libp2pAddrInfoFromPeer(peer)
	if err != nil {
		return ErrNetwork(fmt.Sprintf(
			"converting peer info, pokt address: %s", peer.Address,
		), err)
	}

	p2pNet.host.Peerstore().RemovePeer(libp2pPeer.ID)
	return nil
}

func (p2pNet *libp2pNetwork) Close() error {
	return p2pNet.host.Close()
}

// setupAddrBookMap initializes p2pNet.addrBookMap using the
// addrBookProvider and currentHeightProvider registered on the bus.
func (p2pNet *libp2pNetwork) setupAddrBookMap() error {
	addrBookProviderModule, err := p2pNet.GetBus().GetModulesRegistry().GetModule(addrbook_provider.ModuleName)
	if err != nil {
		addrBookProviderModule = persABP.NewPersistenceAddrBookProvider(p2pNet.GetBus())
	}
	addrBookProvider := addrBookProviderModule.(providers.AddrBookProvider)

	currentHeightProviderModule, err := p2pNet.GetBus().GetModulesRegistry().GetModule(current_height_provider.ModuleName)
	if err != nil {
		currentHeightProviderModule = p2pNet.GetBus().GetConsensusModule()
	}
	currentHeightProvider := currentHeightProviderModule.(providers.CurrentHeightProvider)

	addrBook, err := addrBookProvider.GetStakedAddrBookAtHeight(currentHeightProvider.CurrentHeight())
	if err != nil {
		p2pNet.logger.Fatal().Err(err).Msg("getting staked address book")
		return err
	}

	addrBookMap := make(typesP2P.AddrBookMap)
	for _, peer := range addrBook {
		addrBookMap[peer.Address.String()] = peer
		pubKey, err := identity.Libp2pPublicKeyFromPeer(peer)
		if err != nil {
			return ErrNetwork(fmt.Sprintf(
				"converting peer public key, pokt address: %s", peer.Address,
			), err)
		}
		libp2pPeer, err := identity.Libp2pAddrInfoFromPeer(peer)
		if err != nil {
			return ErrNetwork(fmt.Sprintf(
				"converting peer info, pokt address: %s", peer.Address,
			), err)
		}

		p2pNet.host.Peerstore().AddAddrs(libp2pPeer.ID, libp2pPeer.Addrs, defaultPeerTTL)
		if err := p2pNet.host.Peerstore().AddPubKey(libp2pPeer.ID, pubKey); err != nil {
			return ErrNetwork(fmt.Sprintf(
				"adding peer public key, pokt address: %s", peer.Address,
			), err)
		}
	}
	p2pNet.addrBookMap = addrBookMap

	return nil
}
