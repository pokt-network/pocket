package raintree

import (
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/foxcpp/go-mockdns"
	"github.com/golang/mock/gomock"
	"github.com/libp2p/go-libp2p/p2p/host/peerstore/pstoremem"
	mocknet "github.com/libp2p/go-libp2p/p2p/net/mock"
	"github.com/stretchr/testify/require"

	"github.com/pokt-network/pocket/internal/testutil"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	mocksP2P "github.com/pokt-network/pocket/p2p/types/mocks"
	"github.com/pokt-network/pocket/runtime/configs"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	mockModules "github.com/pokt-network/pocket/shared/modules/mocks"
)

const (
	serviceURLFormat = "val_%d:42069"
	addrAlphabet     = "ABCDEFGHIJKLMNOPQRSTUVWXYZ["
)

type ExpectedRainTreeNetworkConfig struct {
	numNodes          int
	numExpectedLevels int
}

type ExpectedRainTreeMessageTarget struct {
	level int
	left  string
	right string
}
type ExpectedRainTreeMessageProp struct {
	orig     byte
	numNodes int
	addrList string
	targets  []ExpectedRainTreeMessageTarget
}

func TestRainTree_Peerstore_HandleUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)
	pubKey, err := cryptoPocket.GeneratePublicKey()
	require.NoError(t, err)

	testCases := []ExpectedRainTreeNetworkConfig{
		// 0 levels
		{1, 0}, // Just self
		// 1 level
		{2, 1},
		{3, 1},
		// 2 levels
		{4, 2},
		{9, 2},
		// 3 levels
		{10, 3},
		{27, 3},
		// 4 levels
		{28, 4},
		{81, 4},
		// 5 levels
		{82, 5},
		// 10 levels
		{59049, 10},
		// 11 levels
		{59050, 11},
		// 19 levels
		// NOTE: This does not scale to 1,000,000,000 (1B) nodes because it's too slow.
		//       However, optimizing the code to handle 1B nodes would be a very premature optimization
		//       at this stage in the project's lifecycle, so the comment is simply left to inform
		//       future readers.
		// {1000000000, 19},
	}

	for _, testCase := range testCases {
		n := testCase.numNodes
		t.Run(fmt.Sprintf("n=%d", n), func(t *testing.T) {
			pstore := getPeerstore(t, n-1)

			err = pstore.AddPeer(&typesP2P.NetworkPeer{
				PublicKey:  pubKey,
				Address:    pubKey.Address(),
				ServiceURL: "10.0.0.1:42069",
			})
			require.NoError(t, err)

			mockBus := mockBus(ctrl)
			pstoreProviderMock := mockPeerstoreProvider(ctrl, pstore)
			currentHeightProviderMock := mockCurrentHeightProvider(ctrl, 0)

			libp2pMockNet, err := mocknet.WithNPeers(1)
			require.NoError(t, err)

			netCfg := RainTreeConfig{
				Host:                  libp2pMockNet.Hosts()[0],
				Addr:                  pubKey.Address(),
				PeerstoreProvider:     pstoreProviderMock,
				CurrentHeightProvider: currentHeightProviderMock,
			}

			network, err := NewRainTreeNetwork(mockBus, netCfg)
			require.NoError(t, err)

			rainTree := network.(*rainTreeNetwork)

			peersManagerStateView, actualMaxNumLevels := rainTree.peersManager.getPeersViewWithLevels()

			require.Equal(t, n, network.GetPeerstore().Size())
			require.Len(t, peersManagerStateView.GetAddrs(), n)
			if n < 100 {
				// This is can be slow when `n` is very large.
				require.ElementsMatchf(t, pstore.GetPeerList(), peersManagerStateView.GetPeers(), "peers don't match")
			} else {
				require.Len(t, peersManagerStateView.GetPeers(), n)
			}
			require.Equal(t, testCase.numExpectedLevels, int(actualMaxNumLevels))
		})
	}
}

func BenchmarkPeerstoreUpdates(b *testing.B) {
	ctrl := gomock.NewController(gomock.TestReporter(b))
	addr, err := cryptoPocket.GenerateAddress()
	require.NoError(b, err)

	testCases := []ExpectedRainTreeNetworkConfig{
		// Small
		{9, 2},
		// Large
		{59050, 11},
		// INVESTIGATE(olshansky/team): Does not scale to 1,000,000,000 nodes
		// {1000000000, 19},
	}

	// the test will add this arbitrary number of addresses after the initial initialization (done via NewRainTreeNetwork)
	// this is to add extra subsequent work that -should- grow linearly and it's actually going to test AddressBook updates
	// not simply initializations.
	numAddressesToBeAdded := 1000

	for _, testCase := range testCases {
		n := testCase.numNodes
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			pstore := getPeerstore(nil, n-1)
			err := pstore.AddPeer(&typesP2P.NetworkPeer{Address: addr})
			require.NoError(b, err)

			mockBus := mockBus(ctrl)
			pstoreProviderMock := mockPeerstoreProvider(ctrl, pstore)
			currentHeightProviderMock := mockCurrentHeightProvider(ctrl, 0)
			hostMock := mocksP2P.NewMockHost(ctrl)
			netCfg := RainTreeConfig{
				Host:                  hostMock,
				Addr:                  addr,
				PeerstoreProvider:     pstoreProviderMock,
				CurrentHeightProvider: currentHeightProviderMock,
			}

			network, err := NewRainTreeNetwork(mockBus, netCfg)
			require.NoError(b, err)

			rainTree := network.(*rainTreeNetwork)

			peersManagerStateView, actualMaxNumLevels := rainTree.peersManager.getPeersViewWithLevels()

			require.Equal(b, n, network.GetPeerstore().Size())
			require.Equal(b, n, len(peersManagerStateView.GetAddrs()))
			require.ElementsMatchf(b, pstore.GetPeerList(), peersManagerStateView.GetAddrs(), "peers don't match")
			require.Equal(b, testCase.numExpectedLevels, int(actualMaxNumLevels))

			for i := 0; i < numAddressesToBeAdded; i++ {
				newAddr, err := cryptoPocket.GenerateAddress()
				require.NoError(b, err)
				err = rainTree.AddPeer(&typesP2P.NetworkPeer{Address: newAddr})
				require.NoError(b, err)
			}

			peersManagerStateView = rainTree.peersManager.GetPeersView()

			require.Equal(b, n+numAddressesToBeAdded, network.GetPeerstore().Size())
			require.Equal(b, n+numAddressesToBeAdded, len(peersManagerStateView.GetAddrs()))
		})
	}
}

// Generates an address book with a random set of `n` addresses
func getPeerstore(t *testing.T, n int) typesP2P.Peerstore {
	pstore := make(typesP2P.PeerAddrMap)
	for i := 0; i < n; i++ {
		privKey, err := cryptoPocket.GeneratePrivateKey()
		if t != nil {
			require.NoError(t, err)
		}

		err = pstore.AddPeer(&typesP2P.NetworkPeer{
			PublicKey:  privKey.PublicKey(),
			Address:    privKey.Address(),
			ServiceURL: "10.0.0.1:42069",
		})
		if t != nil {
			require.NoError(t, err)
		}
	}
	return pstore
}

func TestRainTree_MessageTargets_SixNodes(t *testing.T) {
	// 		                  A
	// 		   ┌──────────────┴────────┬────────────────────┐
	// 		   C                       A                    E
	//   ┌─────┴──┬─────┐        ┌─────┴──┬─────┐     ┌─────┴──┬─────┐
	//   D        C     E        B        A     C     F        E     A
	prop := &ExpectedRainTreeMessageProp{'A', 6, "ABCDEF", []ExpectedRainTreeMessageTarget{
		{2, "C", "E"},
		{1, "B", "C"},
	}}
	testRainTreeMessageTargets(t, prop)
}

func TestRainTree_MessageTargets_NineNodes(t *testing.T) {
	//                      A
	//       ┌──────────────┴────────┬────────────────────┐
	//       D                       A                    G
	// ┌─────┴──┬─────┐        ┌─────┴──┬─────┐     ┌─────┴──┬─────┐
	// F        D     H        C        A     E     I        G     B
	prop := &ExpectedRainTreeMessageProp{'A', 9, "ABCDEFGHI", []ExpectedRainTreeMessageTarget{
		{2, "D", "G"},
		{1, "C", "E"},
	}}
	testRainTreeMessageTargets(t, prop)
}
func TestRainTree_MessageTargets_TwentySevenNodes(t *testing.T) {

	// 		                                                             O
	// 		                ┌────────────────────────────────────────────┴───────────────────────┬─────────────────────────────────────────────────────────────────┐
	// 		                X                                                                    O                                                                 F
	//       ┌──────────────┴────────┬────────────────────┐                       ┌──────────────┴────────┬────────────────────┐                    ┌──────────────┴────────┬────────────────────┐
	//       C                       X                    I                       U                       O                    [                    L                       F                    R
	// ┌─────┴──┬─────┐        ┌─────┴──┬─────┐     ┌─────┴──┬─────┐        ┌─────┴──┬─────┐        ┌─────┴──┬─────┐     ┌─────┴──┬─────┐     ┌─────┴──┬─────┐        ┌─────┴──┬─────┐     ┌─────┴──┬─────┐
	// G        C     K        A        X     E     M        I     Z        Y        U     B        S        O     W     D        [     Q     P        L     T        J        F     N     V        R     H
	prop := &ExpectedRainTreeMessageProp{'O', 27, "OPQRSTUVWXYZ[ABCDEFGHIJKLMN", []ExpectedRainTreeMessageTarget{
		{3, "X", "F"},
		{2, "U", "["},
		{1, "S", "W"},
	}}
	testRainTreeMessageTargets(t, prop)
}

func testRainTreeMessageTargets(t *testing.T, expectedMsgProp *ExpectedRainTreeMessageProp) {
	ctrl := gomock.NewController(t)
	busMock := mockModules.NewMockBus(ctrl)
	consensusMock := mockModules.NewMockConsensusModule(ctrl)
	consensusMock.EXPECT().CurrentHeight().Return(uint64(1)).AnyTimes()
	busMock.EXPECT().GetConsensusModule().Return(consensusMock).AnyTimes()
	persistenceMock := mockModules.NewMockPersistenceModule(ctrl)
	busMock.EXPECT().GetPersistenceModule().Return(persistenceMock).AnyTimes()
	runtimeMgrMock := mockModules.NewMockRuntimeMgr(ctrl)
	busMock.EXPECT().GetRuntimeMgr().Return(runtimeMgrMock).AnyTimes()
	runtimeMgrMock.EXPECT().GetConfig().Return(configs.NewDefaultConfig()).AnyTimes()

	mockAlphabetValidatorServiceURLsDNS(t)
	pstore := getAlphabetPeerstore(t, expectedMsgProp.numNodes)
	pstoreProviderMock := mockPeerstoreProvider(ctrl, pstore)
	currentHeightProviderMock := mockCurrentHeightProvider(ctrl, 1)

	libp2pPStore, err := pstoremem.NewPeerstore()
	require.NoError(t, err)

	hostMock := mocksP2P.NewMockHost(ctrl)
	hostMock.EXPECT().Peerstore().Return(libp2pPStore).AnyTimes()

	netCfg := RainTreeConfig{
		Host:                  hostMock,
		Addr:                  []byte{expectedMsgProp.orig},
		PeerstoreProvider:     pstoreProviderMock,
		CurrentHeightProvider: currentHeightProviderMock,
	}

	network, err := NewRainTreeNetwork(busMock, netCfg)
	require.NoError(t, err)
	rainTree := network.(*rainTreeNetwork)

	rainTree.SetBus(busMock)

	peersManagerStateView := rainTree.peersManager.GetPeersView()

	require.Equal(t, strings.Join(peersManagerStateView.GetAddrs(), ""), strToAddrList(expectedMsgProp.addrList))

	i, found := rainTree.peersManager.getSelfIndexInPeersView()
	require.True(t, found)
	require.Equal(t, i, 0)

	for _, target := range expectedMsgProp.targets {
		actualTargets := rainTree.getTargetsAtLevel(uint32(target.level))

		require.True(t, shouldSendToTarget(actualTargets[0]))
		require.Equal(t, cryptoPocket.Address(target.left), actualTargets[0].address)

		require.True(t, shouldSendToTarget(actualTargets[1]))
		require.Equal(t, cryptoPocket.Address(target.right), actualTargets[1].address)
	}
}

// Generates an address book with a constant set 27 addresses; ['A', ..., 'Z']
func getAlphabetPeerstore(t *testing.T, n int) typesP2P.Peerstore {
	pstore := make(typesP2P.PeerAddrMap)
	for i, ch := range addrAlphabet {
		if i >= n {
			return pstore
		}
		pubKey, err := cryptoPocket.GeneratePublicKey()
		require.NoError(t, err)

		err = pstore.AddPeer(&typesP2P.NetworkPeer{
			PublicKey:  pubKey,
			ServiceURL: fmt.Sprintf(serviceURLFormat, i),
			Address:    []byte{byte(ch)},
		})
		require.NoError(t, err)
	}
	return pstore
}

func strToAddrList(s string) string {
	return hex.EncodeToString([]byte(s))
}

func mockAlphabetValidatorServiceURLsDNS(t *testing.T) (done func()) {
	zones := make(map[string]mockdns.Zone)
	for i := range addrAlphabet {
		serviceURL, err := url.Parse(fmt.Sprintf("scheme://"+serviceURLFormat, i))
		require.NoError(t, err)

		fqdn := fmt.Sprintf("%s.", serviceURL.Hostname())
		zones[fqdn] = mockdns.Zone{
			A: []string{fmt.Sprintf("10.0.0.%d", i+1)},
		}
	}

	return testutil.PrepareDNSMock(zones)
}
