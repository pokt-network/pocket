package test

import (
	"encoding/hex"
	"testing"

	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/persistence/schema"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types/genesis"
	"github.com/stretchr/testify/require"
)

func FuzzApplication(f *testing.F) {
	fuzzSingleProtocolActor(f,
		NewTestGenericActor(schema.ApplicationActor, newTestApp),
		GetGenericActor(schema.ApplicationActor, getTestApp),
		schema.ApplicationActor)
}

func TestInsertAppAndExists(t *testing.T) {
	db := NewTestPostgresContext(t, 0)

	app, err := createAndInsertDefaultTestApp(db)
	require.NoError(t, err)

	db.Height = 1

	app2, err := createAndInsertDefaultTestApp(db)
	require.NoError(t, err)

	exists, err := db.GetAppExists(app.Address, 0)
	require.NoError(t, err)
	require.True(t, exists, "actor that should exist at previous height does not")

	exists, err = db.GetAppExists(app.Address, 1)
	require.NoError(t, err)
	require.True(t, exists, "actor that should exist at current height does not")

	exists, err = db.GetAppExists(app2.Address, 0)
	require.NoError(t, err)
	require.False(t, exists, "actor that should not exist at previous height appears to")

	exists, err = db.GetAppExists(app2.Address, 1)
	require.NoError(t, err)
	require.True(t, exists, "actor that should exist at current height does not")
}

func TestUpdateApp(t *testing.T) {
	db := NewTestPostgresContext(t, 0)

	app, err := createAndInsertDefaultTestApp(db)
	require.NoError(t, err)

	_, _, stakedTokens, _, _, _, _, chains, err := db.GetApp(app.Address, 0)
	require.NoError(t, err)
	require.Equal(t, DefaultChains, chains, "default chains incorrect for current height")
	require.Equal(t, DefaultStake, stakedTokens, "default stake incorrect for current height")

	db.Height = 1

	require.NotEqual(t, DefaultStake, StakeToUpdate)   // sanity check to make sure the tests are correct
	require.NotEqual(t, DefaultChains, ChainsToUpdate) // sanity check to make sure the tests are correct
	err = db.UpdateApp(app.Address, app.MaxRelays, StakeToUpdate, ChainsToUpdate)
	require.NoError(t, err)

	_, _, stakedTokens, _, _, _, _, chains, err = db.GetApp(app.Address, 0)
	require.NoError(t, err)
	require.Equal(t, DefaultChains, chains, "default chains incorrect for previous height")
	require.Equal(t, DefaultStake, stakedTokens, "default stake incorrect for previous height")

	_, _, stakedTokens, _, _, _, _, chains, err = db.GetApp(app.Address, 1)
	require.NoError(t, err)
	require.Equal(t, ChainsToUpdate, chains, "chains not updated for current height")
	require.Equal(t, StakeToUpdate, stakedTokens, "stake not updated for current height")
}

func TestGetAppsReadyToUnstake(t *testing.T) {
	db := NewTestPostgresContext(t, 0)

	app, err := createAndInsertDefaultTestApp(db)
	require.NoError(t, err)

	app2, err := createAndInsertDefaultTestApp(db)
	require.NoError(t, err)

	app3, err := createAndInsertDefaultTestApp(db)
	require.NoError(t, err)

	// Unstake app at height 0
	err = db.SetAppUnstakingHeightAndStatus(app.Address, 0, persistence.UnstakingStatus)
	require.NoError(t, err)

	// Unstake app2 and app3 at height 1
	err = db.SetAppUnstakingHeightAndStatus(app2.Address, 1, persistence.UnstakingStatus)
	require.NoError(t, err)
	err = db.SetAppUnstakingHeightAndStatus(app3.Address, 1, persistence.UnstakingStatus)
	require.NoError(t, err)

	// Check unstaking apps at height 0
	unstakingApps, err := db.GetAppsReadyToUnstake(0, persistence.UnstakingStatus)
	require.NoError(t, err)
	require.Equal(t, 1, len(unstakingApps), "wrong number of actors ready to unstake at height 0")
	require.Equal(t, app.Address, unstakingApps[0].Address, "unexpected application actor returned")

	// Check unstaking apps at height 1
	unstakingApps, err = db.GetAppsReadyToUnstake(1, persistence.UnstakingStatus)
	require.NoError(t, err)
	require.Equal(t, 2, len(unstakingApps), "wrong number of actors ready to unstake at height 1")
	require.ElementsMatch(t, [][]byte{app2.Address, app3.Address}, [][]byte{unstakingApps[0].Address, unstakingApps[1].Address})
}

func TestGetAppStatus(t *testing.T) {
	db := NewTestPostgresContext(t, 1)

	app, err := createAndInsertDefaultTestApp(db)
	require.NoError(t, err)

	// Check status before the app exists
	status, err := db.GetAppStatus(app.Address, 0)
	require.Error(t, err)
	require.Equal(t, status, persistence.UndefinedStakingStatus, "unexpected status")

	// Check status after the app exists
	status, err = db.GetAppStatus(app.Address, 1)
	require.NoError(t, err)
	require.Equal(t, status, DefaultStakeStatus, "unexpected status")
}

func TestGetAppPauseHeightIfExists(t *testing.T) {
	db := NewTestPostgresContext(t, 1)

	app, err := createAndInsertDefaultTestApp(db)
	require.NoError(t, err)

	// Check pause height when app does not exist
	pauseHeight, err := db.GetAppPauseHeightIfExists(app.Address, 0)
	require.Error(t, err)
	require.Equal(t, pauseHeight, DefaultPauseHeight, "unexpected pause height")

	// Check pause height when app does not exist
	pauseHeight, err = db.GetAppPauseHeightIfExists(app.Address, 1)
	require.NoError(t, err)
	require.Equal(t, pauseHeight, DefaultPauseHeight, "unexpected pause height")
}

func TestSetAppPauseHeightAndUnstakeLater(t *testing.T) {
	db := NewTestPostgresContext(t, 0)

	app, err := createAndInsertDefaultTestApp(db)
	require.NoError(t, err)

	pauseHeight := int64(1)
	unstakingHeight := pauseHeight + 10

	err = db.SetAppPauseHeight(app.Address, pauseHeight)
	require.NoError(t, err)

	_, _, _, _, _, appPausedHeight, _, _, err := db.GetApp(app.Address, db.Height)
	require.NoError(t, err)
	require.Equal(t, pauseHeight, appPausedHeight, "pause height not updated")

	err = db.SetAppStatusAndUnstakingHeightIfPausedBefore(pauseHeight+1, unstakingHeight, -1 /*unused*/)
	require.NoError(t, err)

	_, _, _, _, _, _, appUnstakingHeight, _, err := db.GetApp(app.Address, db.Height)
	require.NoError(t, err)
	require.Equal(t, unstakingHeight, appUnstakingHeight, "unstaking height was not set correctly")
}

func TestGetAppOutputAddress(t *testing.T) {
	db := NewTestPostgresContext(t, 0)

	app, err := createAndInsertDefaultTestApp(db)
	require.NoError(t, err)

	output, err := db.GetAppOutputAddress(app.Address, 0)
	require.NoError(t, err)
	require.Equal(t, output, app.Output, "unexpected output address")
}

func newTestApp() (*genesis.App, error) {
	operatorKey, err := crypto.GeneratePublicKey()
	if err != nil {
		return nil, err
	}

	outputAddr, err := crypto.GenerateAddress()
	if err != nil {
		return nil, err
	}

	return &genesis.App{
		Address:         operatorKey.Address(),
		PublicKey:       operatorKey.Bytes(),
		Paused:          false,
		Status:          genesis.DefaultStakeStatus,
		Chains:          genesis.DefaultChains,
		MaxRelays:       DefaultMaxRelays,
		StakedTokens:    genesis.DefaultStake,
		PausedHeight:    DefaultPauseHeight,
		UnstakingHeight: DefaultUnstakingHeight,
		Output:          outputAddr,
	}, nil
}

// TODO_IN_THIS_COMMIT: We are only calling these functions and tests for apps, but need to
// generalize to other actors.
func TestGetSetStakeAmount(t *testing.T) {
	var newStakeAmount = "new_stake_amount"
	db := NewTestPostgresContext(t, 1)

	app, err := createAndInsertDefaultTestApp(db)
	require.NoError(t, err)

	// Check stake amount before
	stakeAmount, err := db.GetAppStakeAmount(1, app.Address)
	require.NoError(t, err)
	require.Equal(t, DefaultStake, stakeAmount, "unexpected beginning stakeAmount")

	// Check stake amount after
	err = db.SetAppStakeAmount(app.Address, newStakeAmount)
	require.NoError(t, err)
	stakeAmountAfter, err := db.GetAppStakeAmount(1, app.Address)
	require.NoError(t, err)
	require.Equal(t, newStakeAmount, stakeAmountAfter, "unexpected status")
}

func TestGetAllApps(t *testing.T) {
	db := NewTestPostgresContext(t, 0)

	updateApp := func(db *persistence.PostgresContext, app *genesis.App) error {
		return db.UpdateApp(app.Address, OlshanskyURL, app.MaxRelays, OlshanskyChains)
	}

	getAllActorsTest(t, db, db.GetAllApps, createAndInsertDefaultTestApp, updateApp, 1)
}

func createAndInsertDefaultTestApp(db *persistence.PostgresContext) (*genesis.App, error) {
	app, err := newTestApp()
	if err != nil {
		return nil, err
	}

	return app, db.InsertApp(
		app.Address,
		app.PublicKey,
		app.Output,
		false,
		DefaultStakeStatus,
		DefaultMaxRelays,
		DefaultStake,
		DefaultChains,
		DefaultPauseHeight,
		DefaultUnstakingHeight)
}

func getTestApp(db *persistence.PostgresContext, address []byte) (*genesis.App, error) {
	operator, publicKey, stakedTokens, maxRelays, outputAddress, pauseHeight, unstakingHeight, chains, err := db.GetApp(address, db.Height)
	if err != nil {
		return nil, err
	}

	operatorAddr, err := hex.DecodeString(operator)
	if err != nil {
		return nil, err
	}

	operatorPubKey, err := hex.DecodeString(publicKey)
	if err != nil {
		return nil, err
	}

	outputAddr, err := hex.DecodeString(outputAddress)
	if err != nil {
		return nil, err
	}

	return &genesis.App{
		Address:         operatorAddr,
		PublicKey:       operatorPubKey,
		Paused:          false,
		Status:          persistence.UnstakingHeightToStatus(unstakingHeight),
		Chains:          chains,
		MaxRelays:       maxRelays,
		StakedTokens:    stakedTokens,
		PausedHeight:    pauseHeight,
		UnstakingHeight: unstakingHeight,
		Output:          outputAddr,
	}, nil
}
