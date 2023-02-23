package network

import (
	"context"
	"fmt"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	libp2pHost "github.com/libp2p/go-libp2p/core/host"

	"github.com/pokt-network/pocket/libp2p/protocol"
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
	if err := p2pNet.setup(); err != nil {
		return nil, err
	}

	return p2pNet, nil
}

// NetworkBroadcast uses the configured pubsub router to broadcast data to peers.
func (p2pNet *libp2pNetwork) NetworkBroadcast(data []byte) error {
	// IMPROVE: receive context in interface methods.
	ctx := context.Background()

	// NB: Routed send using pubsub
	if err := p2pNet.topic.Publish(ctx, data); err != nil {
		return fmt.Errorf("unable to publish to topic: %w", err)
	}
	return nil
}

// NetworkSend connects sends data directly to the specified peer.
func (p2pNet *libp2pNetwork) NetworkSend(data []byte, poktAddr crypto.Address) error {
	// IMPROVE: receive context in interface methods.
	ctx := context.Background()

	selfPoktAddr, err := p2pNet.GetBus().GetP2PModule().GetAddress()
	if err != nil {
		return fmt.Errorf(
			"sending to poktAddr: %s: %w",
			poktAddr,
			err,
		)
	}

	// Don't send to self.
	if selfPoktAddr.Equals(poktAddr) {
		return nil
	}

	peer, ok := p2pNet.addrBookMap[poktAddr.String()]
	if !ok {
		// This should not happen.
		return fmt.Errorf(
			"peer not found in address book, pokt address: %s: %w",
			poktAddr,
			err,
		)
	}

	peerAddrInfo, err := Libp2pAddrInfoFromPeer(peer)
	if err != nil {
		return fmt.Errorf("parsing peer multiaddr: %w", err)
	}

	stream, err := p2pNet.host.NewStream(ctx, peerAddrInfo.ID, protocol.PoktProtocolID)
	if err != nil {
		return fmt.Errorf(
			"opening peer stream, pokt address: %s: %w",
			poktAddr,
			err,
		)
	}

	if _, err := stream.Write(data); err != nil {
		return fmt.Errorf(
			"writing to stream, peer address: %s: %w",
			poktAddr,
			err,
		)
	}
	defer func() {
		// Close the stream so that peer receives EOF.
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

	pubKey, err := Libp2pPublicKeyFromPeer(peer)
	if err != nil {
		return fmt.Errorf(
			"converting peer public key, pokt address: %s: %w",
			peer.Address,
			err,
		)
	}
	libp2pPeer, err := Libp2pAddrInfoFromPeer(peer)
	if err != nil {
		return fmt.Errorf(
			"converting peer info, pokt address: %s: %w",
			peer.Address,
			err,
		)
	}

	p2pNet.host.Peerstore().AddAddrs(libp2pPeer.ID, libp2pPeer.Addrs, defaultPeerTTL)
	if err := p2pNet.host.Peerstore().AddPubKey(libp2pPeer.ID, pubKey); err != nil {
		return fmt.Errorf(
			"adding peer public key, pokt address: %s: %w",
			peer.Address,
			err,
		)
	}
	return nil
}

func (p2pNet *libp2pNetwork) RemovePeerFromAddrBook(peer *typesP2P.NetworkPeer) error {
	delete(p2pNet.addrBookMap, peer.Address.String())

	libp2pPeer, err := Libp2pAddrInfoFromPeer(peer)
	if err != nil {
		return fmt.Errorf(
			"converting peer info, pokt address: %s: %w",
			peer.Address,
			err,
		)
	}

	p2pNet.host.Peerstore().RemovePeer(libp2pPeer.ID)
	return nil
}

func (p2pNet *libp2pNetwork) Close() error {
	return p2pNet.host.Close()
}

// setupAddrBookProvider attempts to retrieve the address book provider from the
// bus, if one is registered, otherwise returns a new `persistenceAddrBookProvider`.
func (p2pNet *libp2pNetwork) setupAddrBookProvider() providers.AddrBookProvider {
	addrBookProviderModule, err := p2pNet.GetBus().GetModulesRegistry().GetModule(addrbook_provider.ModuleName)
	if err != nil {
		addrBookProviderModule = persABP.NewPersistenceAddrBookProvider(p2pNet.GetBus())
	}
	return addrBookProviderModule.(providers.AddrBookProvider)
}

// setupCurrentHeightProvider attempts to retrieve the current height provider
// from the bus registry, falls back to the consensus module if none is registered.
func (p2pNet *libp2pNetwork) setupCurrentHeightProvider() providers.CurrentHeightProvider {
	currentHeightProviderModule, err := p2pNet.GetBus().GetModulesRegistry().GetModule(current_height_provider.ModuleName)
	if err != nil {
		currentHeightProviderModule = p2pNet.GetBus().GetConsensusModule()
	}
	return currentHeightProviderModule.(providers.CurrentHeightProvider)
}

// setupAddrBookMap iterates through `addrBook`, storing each entry into the
// returned `AddrBookMap` but also converting each entry to its respective
// libp2p representation and adding it to the libp2p peerstore.
func (p2pNet *libp2pNetwork) setupAddrBookMap(addrBook typesP2P.AddrBook) (typesP2P.AddrBookMap, error) {
	addrBookMap := make(typesP2P.AddrBookMap)

	for _, peer := range addrBook {
		addrBookMap[peer.Address.String()] = peer
		pubKey, err := Libp2pPublicKeyFromPeer(peer)
		if err != nil {
			return nil, fmt.Errorf(
				"converting peer public key, pokt address: %s: %w",
				peer.Address,
				err,
			)
		}
		libp2pPeer, err := Libp2pAddrInfoFromPeer(peer)
		if err != nil {
			return nil, fmt.Errorf(
				"converting peer info, pokt address: %s: %w",
				peer.Address,
				err,
			)
		}

		p2pNet.host.Peerstore().AddAddrs(libp2pPeer.ID, libp2pPeer.Addrs, defaultPeerTTL)
		if err := p2pNet.host.Peerstore().AddPubKey(libp2pPeer.ID, pubKey); err != nil {
			return nil, fmt.Errorf(
				"adding peer public key, pokt address: %s: %w",
				peer.Address,
				err,
			)
		}
	}
	return addrBookMap, nil
}

// setup initializes p2pNet.addrBookMap using the addrBookProvider
// and currentHeightProvider registered on the bus, if preseent.
func (p2pNet *libp2pNetwork) setup() error {
	addrBookProvider := p2pNet.setupAddrBookProvider()
	currentHeightProvider := p2pNet.setupCurrentHeightProvider()

	addrBook, err := addrBookProvider.GetStakedAddrBookAtHeight(currentHeightProvider.CurrentHeight())
	if err != nil {
		return fmt.Errorf("getting staked address book: %w", err)
	}

	addrBookMap, err := p2pNet.setupAddrBookMap(addrBook)
	if err != nil {
		return err
	}

	p2pNet.addrBookMap = addrBookMap
	return nil
}
