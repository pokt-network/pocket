package test

import (
	"encoding/hex"
	"log"
	"testing"

	"github.com/pokt-network/pocket/persistence/types"

	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/stretchr/testify/require"
)

func FuzzApplication(f *testing.F) {
	fuzzSingleProtocolActor(f,
		NewTestGenericActor(types.ApplicationActor, newTestApp),
		GetGenericActor(types.ApplicationActor, getTestApp),
		types.ApplicationActor)
}

func TestInsertAppAndExists(t *testing.T) {
	db := NewTestPostgresContext(t, 0)

	app, err := createAndInsertDefaultTestApp(db)
	require.NoError(t, err)

	db.Height = 1

	app2, err := createAndInsertDefaultTestApp(db)
	require.NoError(t, err)

	addrBz, err := hex.DecodeString(app.Address)
	require.NoError(t, err)
	addrBz2, err := hex.DecodeString(app2.Address)
	require.NoError(t, err)

	exists, err := db.GetAppExists(addrBz, 0)
	require.NoError(t, err)
	require.True(t, exists, "actor that should exist at previous height does not")
	exists, err = db.GetAppExists(addrBz, 1)
	require.NoError(t, err)
	require.True(t, exists, "actor that should exist at current height does not")

	exists, err = db.GetAppExists(addrBz2, 0)
	require.NoError(t, err)
	require.False(t, exists, "actor that should not exist at previous height appears to")
	exists, err = db.GetAppExists(addrBz2, 1)
	require.NoError(t, err)
	require.True(t, exists, "actor that should exist at current height does not")
}

func TestUpdateApp(t *testing.T) {
	db := NewTestPostgresContext(t, 0)

	app, err := createAndInsertDefaultTestApp(db)
	require.NoError(t, err)

	addrBz, err := hex.DecodeString(app.Address)
	require.NoError(t, err)

	_, _, stakedTokens, _, _, _, _, chains, err := db.GetApp(addrBz, 0)
	require.NoError(t, err)
	require.Equal(t, DefaultChains, chains, "default chains incorrect for current height")
	require.Equal(t, DefaultStake, stakedTokens, "default stake incorrect for current height")

	db.Height = 1

	require.NotEqual(t, DefaultStake, StakeToUpdate)   // sanity check to make sure the tests are correct
	require.NotEqual(t, DefaultChains, ChainsToUpdate) // sanity check to make sure the tests are correct
	err = db.UpdateApp(addrBz, app.GenericParam, StakeToUpdate, ChainsToUpdate)
	require.NoError(t, err)

	_, _, stakedTokens, _, _, _, _, chains, err = db.GetApp(addrBz, 0)
	require.NoError(t, err)
	require.Equal(t, DefaultChains, chains, "default chains incorrect for previous height")
	require.Equal(t, DefaultStake, stakedTokens, "default stake incorrect for previous height")

	_, _, stakedTokens, _, _, _, _, chains, err = db.GetApp(addrBz, 1)
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

	addrBz, err := hex.DecodeString(app.Address)
	require.NoError(t, err)
	addrBz2, err := hex.DecodeString(app2.Address)
	require.NoError(t, err)
	addrBz3, err := hex.DecodeString(app3.Address)
	require.NoError(t, err)

	// Unstake app at height 0
	err = db.SetAppUnstakingHeightAndStatus(addrBz, 0, persistence.UnstakingStatus)
	require.NoError(t, err)

	// Unstake app2 and app3 at height 1
	err = db.SetAppUnstakingHeightAndStatus(addrBz2, 1, persistence.UnstakingStatus)
	require.NoError(t, err)
	err = db.SetAppUnstakingHeightAndStatus(addrBz3, 1, persistence.UnstakingStatus)
	require.NoError(t, err)

	// Check unstaking apps at height 0
	unstakingApps, err := db.GetAppsReadyToUnstake(0, persistence.UnstakingStatus)
	require.NoError(t, err)
	require.Equal(t, 1, len(unstakingApps), "wrong number of actors ready to unstake at height 0")
	require.Equal(t, app.Address, hex.EncodeToString(unstakingApps[0].GetAddress()), "unexpected application actor returned")

	// Check unstaking apps at height 1
	unstakingApps, err = db.GetAppsReadyToUnstake(1, persistence.UnstakingStatus)
	require.NoError(t, err)
	require.Equal(t, 2, len(unstakingApps), "wrong number of actors ready to unstake at height 1")
	require.ElementsMatch(t, [][]byte{addrBz2, addrBz3}, [][]byte{unstakingApps[0].GetAddress(), unstakingApps[1].GetAddress()})
}

func TestGetAppStatus(t *testing.T) {
	db := NewTestPostgresContext(t, 1)

	app, err := createAndInsertDefaultTestApp(db)
	require.NoError(t, err)
	addrBz, err := hex.DecodeString(app.Address)
	require.NoError(t, err)

	// Check status before the app exists
	status, err := db.GetAppStatus(addrBz, 0)
	require.Error(t, err)
	require.Equal(t, status, persistence.UndefinedStakingStatus, "unexpected status")

	// Check status after the app exists
	status, err = db.GetAppStatus(addrBz, 1)
	require.NoError(t, err)
	require.Equal(t, status, DefaultStakeStatus, "unexpected status")
}

func TestGetAppPauseHeightIfExists(t *testing.T) {
	db := NewTestPostgresContext(t, 1)

	app, err := createAndInsertDefaultTestApp(db)
	require.NoError(t, err)
	addrBz, err := hex.DecodeString(app.Address)
	require.NoError(t, err)

	// Check pause height when app does not exist
	pauseHeight, err := db.GetAppPauseHeightIfExists(addrBz, 0)
	require.Error(t, err)
	require.Equal(t, pauseHeight, DefaultPauseHeight, "unexpected pause height")

	// Check pause height when app does not exist
	pauseHeight, err = db.GetAppPauseHeightIfExists(addrBz, 1)
	require.NoError(t, err)
	require.Equal(t, pauseHeight, DefaultPauseHeight, "unexpected pause height")
}

func TestSetAppPauseHeightAndUnstakeLater(t *testing.T) {
	db := NewTestPostgresContext(t, 0)

	app, err := createAndInsertDefaultTestApp(db)
	require.NoError(t, err)

	pauseHeight := int64(1)
	unstakingHeight := pauseHeight + 10
	addrBz, err := hex.DecodeString(app.Address)
	require.NoError(t, err)

	err = db.SetAppPauseHeight(addrBz, pauseHeight)
	require.NoError(t, err)

	_, _, _, _, _, appPausedHeight, _, _, err := db.GetApp(addrBz, db.Height)
	require.NoError(t, err)
	require.Equal(t, pauseHeight, appPausedHeight, "pause height not updated")

	err = db.SetAppStatusAndUnstakingHeightIfPausedBefore(pauseHeight+1, unstakingHeight, -1 /*unused*/)
	require.NoError(t, err)

	_, _, _, _, _, _, appUnstakingHeight, _, err := db.GetApp(addrBz, db.Height)
	require.NoError(t, err)
	require.Equal(t, unstakingHeight, appUnstakingHeight, "unstaking height was not set correctly")
}

func TestGetAppOutputAddress(t *testing.T) {
	db := NewTestPostgresContext(t, 0)

	app, err := createAndInsertDefaultTestApp(db)
	require.NoError(t, err)
	addrBz, err := hex.DecodeString(app.Address)
	require.NoError(t, err)
	output, err := db.GetAppOutputAddress(addrBz, 0)
	require.NoError(t, err)
	require.Equal(t, hex.EncodeToString(output), app.Output, "unexpected output address")
}

func newTestApp() (*types.Actor, error) {
	operatorKey, err := crypto.GeneratePublicKey()
	if err != nil {
		return nil, err
	}

	outputAddr, err := crypto.GenerateAddress()
	if err != nil {
		return nil, err
	}

	return &types.Actor{
		Address:         hex.EncodeToString(operatorKey.Address()),
		PublicKey:       hex.EncodeToString(operatorKey.Bytes()),
		Chains:          DefaultChains,
		GenericParam:    DefaultMaxRelays,
		StakedAmount:    DefaultStake,
		PausedHeight:    DefaultPauseHeight,
		UnstakingHeight: DefaultUnstakingHeight,
		Output:          hex.EncodeToString(outputAddr),
	}, nil
}

func TestGetSetStakeAmount(t *testing.T) {
	var newStakeAmount = "new_stake_amount"
	db := NewTestPostgresContext(t, 1)

	app, err := createAndInsertDefaultTestApp(db)
	require.NoError(t, err)

	addrBz, err := hex.DecodeString(app.Address)
	require.NoError(t, err)

	// Check stake amount before
	stakeAmount, err := db.GetAppStakeAmount(1, addrBz)
	require.NoError(t, err)
	require.Equal(t, DefaultStake, stakeAmount, "unexpected beginning stakeAmount")

	// Check stake amount after
	err = db.SetAppStakeAmount(addrBz, newStakeAmount)
	require.NoError(t, err)
	stakeAmountAfter, err := db.GetAppStakeAmount(1, addrBz)
	require.NoError(t, err)
	require.Equal(t, newStakeAmount, stakeAmountAfter, "unexpected status")
}

func createAndInsertDefaultTestApp(db *persistence.PostgresContext) (*types.Actor, error) {
	app, err := newTestApp()
	if err != nil {
		return nil, err
	}
	// TODO(andrew): Avoid the use of `log.Fatal(fmt.Sprintf`
	// TODO(andrew): Use `require.NoError` instead of `log.Fatal` in tests`
	addrBz, err := hex.DecodeString(app.Address)
	if err != nil {
		log.Fatalf("an error occurred converting address to bytes %s", app.Address)
	}
	pubKeyBz, err := hex.DecodeString(app.PublicKey)
	if err != nil {
		log.Fatalf("an error occurred converting pubKey to bytes %s", app.PublicKey)
	}
	outputBz, err := hex.DecodeString(app.Output)
	if err != nil {
		log.Fatalf("an error occurred converting output to bytes %s", app.Output)
	}
	return app, db.InsertApp(
		addrBz,
		pubKeyBz,
		outputBz,
		false,
		DefaultStakeStatus,
		DefaultMaxRelays,
		DefaultStake,
		DefaultChains,
		DefaultPauseHeight,
		DefaultUnstakingHeight)
}

func getTestApp(db *persistence.PostgresContext, address []byte) (*types.Actor, error) {
	operator, publicKey, stakedTokens, maxRelays, outputAddress, pauseHeight, unstakingHeight, chains, err := db.GetApp(address, db.Height)
	if err != nil {
		return nil, err
	}

	return &types.Actor{
		Address:         operator,
		PublicKey:       publicKey,
		Chains:          chains,
		GenericParam:    maxRelays,
		StakedAmount:    stakedTokens,
		PausedHeight:    pauseHeight,
		UnstakingHeight: unstakingHeight,
		Output:          outputAddress,
	}, nil
}
