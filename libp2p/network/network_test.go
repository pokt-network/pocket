package network

import (
	"context"
	"testing"

	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/shared/crypto"
)

func TestLibp2pNetwork_AddPeerToAddrBook(t *testing.T) {
	p2pNet := newTestLibp2pNetwork(t)
	peerstore := p2pNet.host.Peerstore()

	// NB: assert initial state
	require.Len(t, p2pNet.addrBookMap, 1)

	var (
		existingAddrStr string
		existingPeer    *types.NetworkPeer
	)
	for addr, peer := range p2pNet.addrBookMap {
		existingAddrStr = addr
		existingPeer = peer
	}
	require.NotEmpty(t, existingAddrStr)
	require.NotNil(t, existingPeer)

	existingPeerInfo, err := Libp2pAddrInfoFromPeer(existingPeer)
	require.NoError(t, err)

	existingPeerstoreAddrs := peerstore.Addrs(existingPeerInfo.ID)
	require.Len(t, existingPeerstoreAddrs, 1)

	existingPeerMultiaddr, err := Libp2pMultiaddrFromServiceUrl(existingPeer.ServiceUrl)
	require.NoError(t, err)
	require.Equal(t, existingPeerstoreAddrs[0].String(), existingPeerMultiaddr.String())

	newPublicKey, err := crypto.GeneratePublicKey()
	newPoktAddr := newPublicKey.Address()
	require.NoError(t, err)

	newPeer := &types.NetworkPeer{
		PublicKey: newPublicKey,
		Address:   newPoktAddr,
		// Exercises DNS resolution.
		// IMPROVE: this test will be flakey in the presence of DNS interruptions.
		ServiceUrl: "www.google.com:8080",
	}
	newPeerInfo, err := Libp2pAddrInfoFromPeer(newPeer)
	require.NoError(t, err)
	newPeerMultiaddr := newPeerInfo.Addrs[0]

	// NB: add to address book
	err = p2pNet.AddPeerToAddrBook(newPeer)
	require.NoError(t, err)

	require.Len(t, p2pNet.addrBookMap, 2)
	require.Equal(t, p2pNet.addrBookMap[existingAddrStr], existingPeer)
	require.Equal(t, p2pNet.addrBookMap[newPeer.Address.String()], newPeer)

	existingPeerstoreAddrs = peerstore.Addrs(existingPeerInfo.ID)
	newPeerstoreAddrs := peerstore.Addrs(newPeerInfo.ID)

	require.Len(t, existingPeerstoreAddrs, 1)
	require.Len(t, newPeerstoreAddrs, 1)
	require.Equal(t, newPeerstoreAddrs[0].String(), newPeerMultiaddr.String())
}

func TestLibp2pNetwork_RemovePeerToAddrBook(t *testing.T) {
	p2pNet := newTestLibp2pNetwork(t)
	peerstore := p2pNet.host.Peerstore()

	// NB: assert initial state
	require.Len(t, p2pNet.addrBookMap, 1)

	var (
		existingAddrStr string
		existingPeer    *types.NetworkPeer
	)
	for addr, peer := range p2pNet.addrBookMap {
		existingAddrStr = addr
		existingPeer = peer
	}
	require.NotEmpty(t, existingAddrStr)
	require.NotNil(t, existingPeer)

	existingPeerInfo, err := Libp2pAddrInfoFromPeer(existingPeer)
	require.NoError(t, err)

	existingPeerstoreAddrs := peerstore.Addrs(existingPeerInfo.ID)
	require.Len(t, existingPeerstoreAddrs, 1)

	// TECHDEBT: currently, the internal addrbook is storing multiaddrs
	// in the serviceUrl field as opposed to transforming them (back and forth).
	existingPeerMultiaddr, err := Libp2pMultiaddrFromServiceUrl(existingPeer.ServiceUrl)
	require.NoError(t, err)
	require.Equal(t, existingPeerstoreAddrs[0].String(), existingPeerMultiaddr.String())

	err = p2pNet.RemovePeerFromAddrBook(existingPeer)
	require.NoError(t, err)

	require.Len(t, p2pNet.addrBookMap, 0)

	// NB: peerstore implementations seem to only remove peer keys and
	// metadata but not the embedded AddrBook entry.
	// (see: https://github.com/libp2p/go-libp2p/blob/v0.25.1/p2p/host/peerstore/pstoremem/peerstore.go#L108)
	// (see: https://github.com/libp2p/go-libp2p/blob/v0.25.1/p2p/host/peerstore/pstoreds/peerstore.go#L187)

	existingPeerstoreAddrs = peerstore.Addrs(existingPeerInfo.ID)
	require.Len(t, existingPeerstoreAddrs, 1)
}

func newTestLibp2pNetwork(t *testing.T) *libp2pNetwork {
	ctx := context.Background()

	// INCOMPLETE (SOON OBSOLETE): Only testing pocket address book <-> libp2p
	// peerstore integration. No need to mock an entire network, just a
	// starting pocket address book.
	runtimeConfigs := createMockRuntimeMgrs(t, 1)
	busMock := createMockBus(t, runtimeConfigs[0], 1)
	consensusMock := prepareConsensusMock(t, busMock)

	prepareBusMock(busMock, consensusMock)

	logger_ := logger.Global.CreateLoggerForModule("test_module")

	// NB: will bind to a random, available port for the duration of this test.
	host, err := libp2p.New(libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/0"))
	require.NoError(t, err)
	defer host.Close()

	pubsub_, err := pubsub.NewFloodSub(ctx, host)
	require.NoError(t, err)

	topic, err := pubsub_.Join("test_protocol")
	require.NoError(t, err)

	p2pNetwork, err := NewLibp2pNetwork(
		busMock,
		logger_,
		host,
		topic,
	)
	require.NoError(t, err)

	libp2pNet, ok := p2pNetwork.(*libp2pNetwork)
	if !assert.True(t, ok) {
		t.Fatalf("p2pNetwork type: %T", p2pNetwork)
	}
	return libp2pNet
}
