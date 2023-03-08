package network

import (
	"context"
	"fmt"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	libp2pHost "github.com/libp2p/go-libp2p/core/host"

	"github.com/pokt-network/pocket/libp2p/protocol"
	"github.com/pokt-network/pocket/p2p/providers"
	"github.com/pokt-network/pocket/p2p/providers/current_height_provider"
	"github.com/pokt-network/pocket/p2p/providers/peerstore_provider"
	persABP "github.com/pokt-network/pocket/p2p/providers/peerstore_provider/persistence"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
	sharedP2P "github.com/pokt-network/pocket/shared/p2p"
)

const (
	week = time.Hour * 24 * 7
	// TECHDEBT: consider more carefully and parameterize.
	defaultPeerTTL = 2 * week
)

var (
	_ typesP2P.Network = &libp2pNetwork{}
)

type libp2pNetwork struct {
	base_modules.IntegratableModule

	logger *modules.Logger
	host   libp2pHost.Host
	topic  *pubsub.Topic
	pstore sharedP2P.Peerstore
}

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

	peer := p2pNet.pstore.GetPeer(poktAddr)
	if peer == nil {
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

func (p2pNet *libp2pNetwork) GetPeerstore() sharedP2P.Peerstore {
	return p2pNet.pstore
}

func (p2pNet *libp2pNetwork) AddPeer(peer sharedP2P.Peer) error {
	if err := p2pNet.pstore.AddPeer(peer); err != nil {
		return fmt.Errorf(
			"adding peer, pokt address %s: %w",
			peer.GetAddress(),
			err,
		)
	}

	pubKey, err := Libp2pPublicKeyFromPeer(peer)
	if err != nil {
		return fmt.Errorf(
			"converting peer public key, pokt address: %s: %w",
			peer.GetAddress(),
			err,
		)
	}
	libp2pPeer, err := Libp2pAddrInfoFromPeer(peer)
	if err != nil {
		return fmt.Errorf(
			"converting peer info, pokt address: %s: %w",
			peer.GetAddress(),
			err,
		)
	}

	p2pNet.host.Peerstore().AddAddrs(libp2pPeer.ID, libp2pPeer.Addrs, defaultPeerTTL)
	if err := p2pNet.host.Peerstore().AddPubKey(libp2pPeer.ID, pubKey); err != nil {
		return fmt.Errorf(
			"adding peer public key, pokt address: %s: %w",
			peer.GetAddress(),
			err,
		)
	}
	return nil
}

func (p2pNet *libp2pNetwork) RemovePeer(peer sharedP2P.Peer) error {
	if err := p2pNet.pstore.RemovePeer(peer.GetAddress()); err != nil {
		return fmt.Errorf(
			"removing peer, pokt address %s: %w",
			peer.GetAddress(),
			err,
		)
	}

	libp2pPeer, err := Libp2pAddrInfoFromPeer(peer)
	if err != nil {
		return fmt.Errorf(
			"converting peer info, pokt address: %s: %w",
			peer.GetAddress(),
			err,
		)
	}

	p2pNet.host.Peerstore().RemovePeer(libp2pPeer.ID)
	return nil
}

func (p2pNet *libp2pNetwork) Close() error {
	return p2pNet.host.Close()
}

// setupPeerstoreProvider attempts to retrieve the peerstore provider from the
// bus, if one is registered, otherwise returns a new `persistencePeerstoreProvider`.
func (p2pNet *libp2pNetwork) setupPeerstoreProvider() providers.PeerstoreProvider {
	pstoreProviderModule, err := p2pNet.GetBus().GetModulesRegistry().GetModule(peerstore_provider.ModuleName)
	if err != nil {
		pstoreProviderModule = persABP.NewPersistencePeerstoreProvider(p2pNet.GetBus())
	}
	return pstoreProviderModule.(providers.PeerstoreProvider)
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

// setupHost iterates through peers in given `pstore`, converting peer info for
// use with libp2p and adding it to the underlying libp2p host's peerstore.
// (see: https://pkg.go.dev/github.com/libp2p/go-libp2p@v0.26.2/core/host#Host)
// (see: https://pkg.go.dev/github.com/libp2p/go-libp2p@v0.26.2/core/peerstore#Peerstore)
func (p2pNet *libp2pNetwork) setupHost() error {
	for _, peer := range p2pNet.pstore.GetAllPeers() {
		pubKey, err := Libp2pPublicKeyFromPeer(peer)
		if err != nil {
			return fmt.Errorf(
				"converting peer public key, pokt address: %s: %w",
				peer.GetAddress(),
				err,
			)
		}
		libp2pPeer, err := Libp2pAddrInfoFromPeer(peer)
		if err != nil {
			return fmt.Errorf(
				"converting peer info, pokt address: %s: %w",
				peer.GetAddress(),
				err,
			)
		}

		p2pNet.host.Peerstore().AddAddrs(libp2pPeer.ID, libp2pPeer.Addrs, defaultPeerTTL)
		if err := p2pNet.host.Peerstore().AddPubKey(libp2pPeer.ID, pubKey); err != nil {
			return fmt.Errorf(
				"adding peer public key, pokt address: %s: %w",
				peer.GetAddress(),
				err,
			)
		}
	}
	return nil
}

// setup initializes p2pNet.pstore using the PeerstoreProvider
// and CurrentHeightProvider registered on the bus, if preseent.
func (p2pNet *libp2pNetwork) setup() (err error) {
	peerstoreProvider := p2pNet.setupPeerstoreProvider()
	currentHeightProvider := p2pNet.setupCurrentHeightProvider()

	p2pNet.pstore, err = peerstoreProvider.GetStakedPeerstoreAtHeight(currentHeightProvider.CurrentHeight())
	if err != nil {
		return fmt.Errorf("getting staked peerstore: %w", err)
	}

	if err := p2pNet.setupHost(); err != nil {
		return err
	}
	return nil
}
