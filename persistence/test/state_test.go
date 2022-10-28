package test

// func TestStateHash_DeterministicStateWhenUpdatingAppStake(t *testing.T) {
// 	// These hashes were determined manually by running the test, but hardcoded to guarantee
// 	// that the business logic doesn't change and that they remain deterministic.
// 	encodedAppHash := []string{
// 		"62adad6925267abe075dc62ffb9b8d960709409b097b75dd6b3ea4cce31d1482",
// 		"c1af3fda156bce4162df755f0095ae4f909477fc385f761c6e8d2ef6eb2d9fa6",
// 		"e65d0c2cd78f180d774bfe43e52a49fad490bf208fcc6d167f2b6543ab280cb9",
// 	}

// 	for i := 0; i < 3; i++ {
// 		// Get the context at the new height and retrieve one of the apps
// 		height := int64(i + 1)
// 		heightBz := heightToBytes(height)
// 		expectedAppHash := encodedAppHash[i]

// 		db := NewTestPostgresContext(t, height)

// 		apps, err := db.GetAllApps(height)
// 		require.NoError(t, err)
// 		app := apps[0]

// 		addrBz, err := hex.DecodeString(app.GetAddress())
// 		require.NoError(t, err)

// 		// Update the app's stake
// 		newStakeAmount := types.BigIntToString(big.NewInt(height + int64(420000000000)))
// 		err = db.SetAppStakeAmount(addrBz, newStakeAmount)
// 		require.NoError(t, err)

// 		// NOTE: The tx does not currently affect the state hash
// 		txBz := []byte("a tx, i am, which set the app stake amount to " + newStakeAmount)
// 		// txResult := types.DefaultTx
// 		// err = db.StoreTransaction(txBz)
// 		// require.NoError(t, err)

// 		// Update & commit the state hash
// 		appHash, err := db.UpdateAppHash()
// 		require.NoError(t, err)
// 		require.Equal(t, expectedAppHash, hex.EncodeToString(appHash))

// 		err = db.Commit([]byte("proposer"), []byte("quorumCert"))
// 		require.NoError(t, err)

// 		// Verify the block contents
// 		blockBz, err := testPersistenceMod.GetBlockStore().Get(heightBz)
// 		require.NoError(t, err)

// 		var block types.Block
// 		err = codec.GetCodec().Unmarshal(blockBz, &block)
// 		require.NoError(t, err)
// 		require.Len(t, block.Transactions, 1)
// 		require.Equal(t, txBz, block.Transactions[0])
// 		require.Equal(t, expectedAppHash, block.Hash) // block
// 		if i > 0 {
// 			require.Equal(t, encodedAppHash[i-1], block.PrevHash) // chain
// 		}

// 	}
// }

// // Tests/debug to implement:
// // - Visibility into what's in the tree
// // - Benchmarking many inserts
// // - Release / revert mid block and making sure everything is reverted
// // - Thinking about how it can be synched
// // - Playing back several blocks

// func heightToBytes(height int64) []byte {
// 	heightBytes := make([]byte, 8)
// 	binary.LittleEndian.PutUint64(heightBytes, uint64(height))
// 	return heightBytes
// }
