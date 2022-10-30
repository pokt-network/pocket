package test

import (
	"fmt"
	"log"
	"math/rand"
	"testing"

	"github.com/pokt-network/pocket/persistence/indexer"
	"github.com/pokt-network/pocket/shared/debug"
)

func BenchmarkStateHash(b *testing.B) {
	b.StopTimer()

	b.Cleanup(func() {
		if err := testPersistenceMod.ReleaseWriteContext(); err != nil {
			log.Fatalf("Error releasing write context: %v\n", err)
		}
		if err := testPersistenceMod.HandleDebugMessage(&debug.DebugMessage{
			Action:  debug.DebugMessageAction_DEBUG_PERSISTENCE_CLEAR_STATE,
			Message: nil,
		}); err != nil {
			log.Fatalf("Error clearing state: %v\n", err)
		}
	})

	// number of heights
	// number of txs per height
	// number of ops per height

	testCases := []struct {
		numHeights     int
		numTxPerHeight int
	}{
		{1, 1},
		{100, 1},
		{100, 10},
		{1000, 1},
		{1000, 10},
		{10000, 10},
		{10000, 1000},
	}

	for _, testCase := range testCases {
		numHeights := testCase.numHeights
		numTxPerHeight := testCase.numTxPerHeight
		b.Run(fmt.Sprintf("heights=%d;txPerHeight=%d", numHeights, numTxPerHeight), func(b *testing.B) {
			for h := 0; h < numHeights; h++ {
				// addrBook := getAddrBook(nil, n-1)
				// addrBook = append(addrBook, &types.NetworkPeer{Address: addr})
				// network := NewRainTreeNetwork(addr, addrBook).(*rainTreeNetwork)

				// peersManagerStateView := network.peersManager.getNetworkView()

				// require.Equal(b, n, len(peersManagerStateView.addrList))
				// require.Equal(b, n, len(peersManagerStateView.addrBookMap))
				// require.Equal(b, testCase.numExpectedLevels, int(peersManagerStateView.maxNumLevels))

				// for i := 0; i < numAddressessToBeAdded; i++ {
				// 	newAddr, err := crypto.GenerateAddress()
				// 	require.NoError(b, err)
				// 	network.AddPeerToAddrBook(&types.NetworkPeer{Address: newAddr})
				// }

				// peersManagerStateView = network.peersManager.getNetworkView()

				// require.Equal(b, n+numAddressessToBeAdded, len(peersManagerStateView.addrList))
				// require.Equal(b, n+numAddressessToBeAdded, len(peersManagerStateView.addrBookMap))

				// db := NewTestPostgresContext(b, height)

				// err = db.StoreTransaction(modules.TxResult(getRandomTxResult(height)))
				// require.NoError(t, err)

				// // db.

				// // Update the state hash
				// appHash, err := db.UpdateAppHash()
				// require.NoError(t, err)
				// require.Equal(t, expectedAppHash, hex.EncodeToString(appHash))

				// // Commit the transactions above
				// err = db.Commit([]byte("TODOproposer"), []byte("TODOquorumCert"))
				// require.NoError(t, err)
			}
		})
	}
}

func getRandomTxResult(height int64) *indexer.TxRes {
	return &indexer.TxRes{
		Tx:            getTxBytes(50),
		Height:        height,
		Index:         0,
		ResultCode:    0,
		Error:         "TODO",
		SignerAddr:    "TODO",
		RecipientAddr: "TODO",
		MessageType:   "TODO",
	}
}

func getTxBytes(numBytes int64) []byte {
	bz := make([]byte, numBytes)
	rand.Read(bz)
	return bz
}

// Random transactions
// Update state hash
