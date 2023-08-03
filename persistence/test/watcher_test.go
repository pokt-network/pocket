package test

import (
	"encoding/hex"
	"log"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pokt-network/pocket/persistence"
	ptypes "github.com/pokt-network/pocket/persistence/types"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/crypto"
)

func FuzzWatcher(f *testing.F) {
	fuzzSingleProtocolActor(f,
		newTestGenericActor(ptypes.WatcherActor, newTestWatcher),
		getGenericActor(ptypes.WatcherActor, getTestWatcher),
		ptypes.WatcherActor)
}

func TestGetSetWatcherStakeAmount(t *testing.T) {
	db := NewTestPostgresContext(t, 1)
	getTestGetSetStakeAmountTest(t, db, createAndInsertDefaultTestWatcher, db.GetWatcherStakeAmount, db.SetWatcherStakeAmount, 1)
}

func TestGetWatcherUpdatedAtHeight(t *testing.T) {
	getWatcherUpdatedFunc := func(db *persistence.PostgresContext, height int64) ([]*coreTypes.Actor, error) {
		return db.GetActorsUpdated(ptypes.WatcherActor, height)
	}
	getAllActorsUpdatedAtHeightTest(t, createAndInsertDefaultTestWatcher, getWatcherUpdatedFunc, 1)
}

func TestInsertWatcherAndExists(t *testing.T) {
	db := NewTestPostgresContext(t, 0)

	watcher, err := createAndInsertDefaultTestWatcher(db)
	require.NoError(t, err)

	db.Height = 1

	watcher2, err := createAndInsertDefaultTestWatcher(db)
	require.NoError(t, err)

	addrBz, err := hex.DecodeString(watcher.Address)
	require.NoError(t, err)
	addrBz2, err := hex.DecodeString(watcher2.Address)
	require.NoError(t, err)

	exists, err := db.GetWatcherExists(addrBz, 0)
	require.NoError(t, err)
	require.True(t, exists, "actor that should exist at previous height does not")
	exists, err = db.GetWatcherExists(addrBz, 1)
	require.NoError(t, err)
	require.True(t, exists, "actor that should exist at current height does not")

	exists, err = db.GetWatcherExists(addrBz2, 0)
	require.NoError(t, err)
	require.False(t, exists, "actor that should not exist at previous height watcherears to")
	exists, err = db.GetWatcherExists(addrBz2, 1)
	require.NoError(t, err)
	require.True(t, exists, "actor that should exist at current height does not")
}

func TestUpdateWatcher(t *testing.T) {
	db := NewTestPostgresContext(t, 0)

	watcher, err := createAndInsertDefaultTestWatcher(db)
	require.NoError(t, err)

	addrBz, err := hex.DecodeString(watcher.Address)
	require.NoError(t, err)

	watch, err := db.GetWatcher(addrBz, 0)
	require.NoError(t, err)
	require.NotNil(t, watch)
	require.Equal(t, DefaultChains, watch.Chains, "default chains incorrect for current height")
	require.Equal(t, DefaultStake, watch.StakedAmount, "default stake incorrect for current height")

	db.Height = 1

	require.NotEqual(t, DefaultStake, StakeToUpdate)   // sanity check to make sure the tests are correct
	require.NotEqual(t, DefaultChains, ChainsToUpdate) // sanity check to make sure the tests are correct
	err = db.UpdateWatcher(addrBz, watcher.ServiceUrl, StakeToUpdate, ChainsToUpdate)
	require.NoError(t, err)

	watch, err = db.GetWatcher(addrBz, 0)
	require.NoError(t, err)
	require.NotNil(t, watch)
	require.Equal(t, DefaultChains, watch.Chains, "default chains incorrect for current height")
	require.Equal(t, DefaultStake, watch.StakedAmount, "default stake incorrect for current height")

	watch, err = db.GetWatcher(addrBz, 1)
	require.NoError(t, err)
	require.NotNil(t, watch)
	require.Equal(t, ChainsToUpdate, watch.Chains, "chains not updated for current height")
	require.Equal(t, StakeToUpdate, watch.StakedAmount, "stake not updated for current height")
}

func TestGetWatchersReadyToUnstake(t *testing.T) {
	db := NewTestPostgresContext(t, 0)

	watcher, err := createAndInsertDefaultTestWatcher(db)
	require.NoError(t, err)

	watcher2, err := createAndInsertDefaultTestWatcher(db)
	require.NoError(t, err)

	watcher3, err := createAndInsertDefaultTestWatcher(db)
	require.NoError(t, err)

	addrBz, err := hex.DecodeString(watcher.Address)
	require.NoError(t, err)
	addrBz2, err := hex.DecodeString(watcher2.Address)
	require.NoError(t, err)
	addrBz3, err := hex.DecodeString(watcher3.Address)
	require.NoError(t, err)

	// Unstake watcher at height 0
	err = db.SetWatcherUnstakingHeightAndStatus(addrBz, 0, int32(coreTypes.StakeStatus_Unstaking))
	require.NoError(t, err)

	// Unstake watcher2 and watcher3 at height 1
	err = db.SetWatcherUnstakingHeightAndStatus(addrBz2, 1, int32(coreTypes.StakeStatus_Unstaking))
	require.NoError(t, err)
	err = db.SetWatcherUnstakingHeightAndStatus(addrBz3, 1, int32(coreTypes.StakeStatus_Unstaking))
	require.NoError(t, err)

	// Check unstaking watchers at height 0
	unstakingWatchers, err := db.GetWatchersReadyToUnstake(0, int32(coreTypes.StakeStatus_Unstaking))
	require.NoError(t, err)
	require.Equal(t, 1, len(unstakingWatchers), "wrong number of actors ready to unstake at height 0")
	require.Equal(t, watcher.Address, unstakingWatchers[0].GetAddress(), "unexpected watcherlication actor returned")

	// Check unstaking watchers at height 1
	unstakingWatchers, err = db.GetWatchersReadyToUnstake(1, int32(coreTypes.StakeStatus_Unstaking))
	require.NoError(t, err)
	require.Equal(t, 2, len(unstakingWatchers), "wrong number of actors ready to unstake at height 1")
	require.ElementsMatch(t, []string{watcher2.Address, watcher3.Address}, []string{unstakingWatchers[0].Address, unstakingWatchers[1].Address})
}

func TestGetWatcherStatus(t *testing.T) {
	db := NewTestPostgresContext(t, 1)

	watcher, err := createAndInsertDefaultTestWatcher(db)
	require.NoError(t, err)

	addrBz, err := hex.DecodeString(watcher.Address)
	require.NoError(t, err)

	// Check status before the watcher exists
	status, err := db.GetWatcherStatus(addrBz, 0)
	require.Error(t, err)
	require.Equal(t, int32(coreTypes.StakeStatus_UnknownStatus), status, "unexpected status")

	// Check status after the watcher exists
	status, err = db.GetWatcherStatus(addrBz, 1)
	require.NoError(t, err)
	require.Equal(t, DefaultStakeStatus, status, "unexpected status")
}

func TestGetWatcherPauseHeightIfExists(t *testing.T) {
	db := NewTestPostgresContext(t, 1)

	watcher, err := createAndInsertDefaultTestWatcher(db)
	require.NoError(t, err)

	addrBz, err := hex.DecodeString(watcher.Address)
	require.NoError(t, err)

	// Check pause height when watcher does not exist
	pauseHeight, err := db.GetWatcherPauseHeightIfExists(addrBz, 0)
	require.Error(t, err)
	require.Equal(t, DefaultPauseHeight, pauseHeight, "unexpected pause height")

	// Check pause height when watcher does not exist
	pauseHeight, err = db.GetWatcherPauseHeightIfExists(addrBz, 1)
	require.NoError(t, err)
	require.Equal(t, DefaultPauseHeight, pauseHeight, "unexpected pause height")
}

func TestSetWatcherPauseHeightAndUnstakeLater(t *testing.T) {
	db := NewTestPostgresContext(t, 0)

	watcher, err := createAndInsertDefaultTestWatcher(db)
	require.NoError(t, err)

	pauseHeight := int64(1)
	unstakingHeight := pauseHeight + 10

	addrBz, err := hex.DecodeString(watcher.Address)
	require.NoError(t, err)

	err = db.SetWatcherPauseHeight(addrBz, pauseHeight)
	require.NoError(t, err)

	watch, err := db.GetWatcher(addrBz, 0)
	require.NoError(t, err)
	require.NotNil(t, watch)
	require.Equal(t, pauseHeight, watch.PausedHeight, "pause height not updated")

	err = db.SetWatcherStatusAndUnstakingHeightIfPausedBefore(pauseHeight+1, unstakingHeight, -1 /*unused*/)
	require.NoError(t, err)

	watch, err = db.GetWatcher(addrBz, 0)
	require.NoError(t, err)
	require.NotNil(t, watch)
	require.Equal(t, unstakingHeight, watch.UnstakingHeight, "unstaking height was not set correctly")
}

func TestGetWatcherOutputAddress(t *testing.T) {
	db := NewTestPostgresContext(t, 0)

	watcher, err := createAndInsertDefaultTestWatcher(db)
	require.NoError(t, err)

	addrBz, err := hex.DecodeString(watcher.Address)
	require.NoError(t, err)

	output, err := db.GetWatcherOutputAddress(addrBz, 0)
	require.NoError(t, err)
	require.Equal(t, watcher.Output, hex.EncodeToString(output), "unexpected output address")
}

func newTestWatcher() (*coreTypes.Actor, error) {
	operatorKey, err := crypto.GeneratePublicKey()
	if err != nil {
		return nil, err
	}

	outputAddr, err := crypto.GenerateAddress()
	if err != nil {
		return nil, err
	}

	return &coreTypes.Actor{
		Address:         hex.EncodeToString(operatorKey.Address()),
		PublicKey:       hex.EncodeToString(operatorKey.Bytes()),
		Chains:          DefaultChains,
		ServiceUrl:      DefaultServiceURL,
		StakedAmount:    DefaultStake,
		PausedHeight:    DefaultPauseHeight,
		UnstakingHeight: DefaultUnstakingHeight,
		Output:          hex.EncodeToString(outputAddr),
	}, nil
}

func createAndInsertDefaultTestWatcher(db *persistence.PostgresContext) (*coreTypes.Actor, error) {
	watcher, err := newTestWatcher()
	if err != nil {
		return nil, err
	}
	addrBz, err := hex.DecodeString(watcher.Address)
	if err != nil {
		log.Fatalf("an error occurred converting address to bytes %s", watcher.Address)
	}
	pubKeyBz, err := hex.DecodeString(watcher.PublicKey)
	if err != nil {
		log.Fatalf("an error occurred converting pubKey to bytes %s", watcher.PublicKey)
	}
	outputBz, err := hex.DecodeString(watcher.Output)
	if err != nil {
		log.Fatalf("an error occurred converting output to bytes %s", watcher.Output)
	}
	return watcher, db.InsertWatcher(
		addrBz,
		pubKeyBz,
		outputBz,
		false,
		DefaultStakeStatus,
		DefaultServiceURL,
		DefaultStake,
		DefaultChains,
		DefaultPauseHeight,
		DefaultUnstakingHeight)
}

func getTestWatcher(db *persistence.PostgresContext, address []byte) (*coreTypes.Actor, error) {
	return db.GetWatcher(address, db.Height)
}
