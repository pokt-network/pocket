package test

import (
	"encoding/hex"
	"testing"

	"github.com/pokt-network/pocket/persistence"
	query "github.com/pokt-network/pocket/persistence/schema"
	"github.com/pokt-network/pocket/shared/crypto"
	typesGenesis "github.com/pokt-network/pocket/shared/types/genesis"
	"github.com/stretchr/testify/require"
)

func FuzzApplication(f *testing.F) {
	fuzzProtocolActor(f,
		NewTestGenericActor(query.ApplicationActor, newTestApp),
		GetGenericActor(query.ApplicationActor, GetTestApp),
		query.ApplicationActor)
}

func TestInsertAppAndExists(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}

	app, err := newTestApp()
	require.NoError(t, err)

	err = db.InsertApp(
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
	require.NoError(t, err)

	db.Height = 1

	app2, err := newTestApp()
	require.NoError(t, err)

	err = db.InsertApp(
		app2.Address,
		app2.PublicKey,
		app2.Output,
		false,
		DefaultStakeStatus,
		DefaultMaxRelays,
		DefaultStake,
		DefaultChains,
		DefaultPauseHeight,
		DefaultUnstakingHeight)
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
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	app, err := newTestApp()
	require.NoError(t, err)

	err = db.InsertApp(
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
	require.NoError(t, err)

	_, _, stakedTokens, _, _, _, _, chains, err := db.GetApp(app.Address, 0)
	require.NoError(t, err)
	require.Equal(t, chains, DefaultChains, "default chains incorrect for current height")
	require.Equal(t, stakedTokens, DefaultStake, "default stake incorrect for current height")

	db.Height = 1

	require.NotEqual(t, DefaultStake, StakeToUpdate)   // sanity check to make sure the tests are correct
	require.NotEqual(t, DefaultChains, ChainsToUpdate) // sanity check to make sure the tests are correct
	err = db.UpdateApp(app.Address, app.MaxRelays, StakeToUpdate, ChainsToUpdate)
	require.NoError(t, err)

	_, _, stakedTokens, _, _, _, _, chains, err = db.GetApp(app.Address, 0)
	require.NoError(t, err)
	require.Equal(t, chains, DefaultChains, "default chains incorrect for previous height")
	require.Equal(t, stakedTokens, DefaultStake, "default stake incorrect for previous height")

	_, _, stakedTokens, _, _, _, _, chains, err = db.GetApp(app.Address, 1)
	require.NoError(t, err)
	require.Equal(t, chains, ChainsToUpdate, "chains not updated for current height")
	require.Equal(t, stakedTokens, StakeToUpdate, "stake not updated for current height")
}

func TestGetAppsReadyToUnstake(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}

	app, err := newTestApp()
	require.NoError(t, err)

	app2, err := newTestApp()
	require.NoError(t, err)

	err = db.InsertApp(
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
	require.NoError(t, err)

	err = db.InsertApp(
		app2.Address,
		app2.PublicKey,
		app2.Output,
		false,
		DefaultStakeStatus,
		DefaultMaxRelays,
		DefaultStake,
		DefaultChains,
		DefaultPauseHeight,
		DefaultUnstakingHeight)
	require.NoError(t, err)

	err = db.SetAppUnstakingHeightAndStatus(app.Address, 0, persistence.UnstakingStatus)
	require.NoError(t, err)

	err = db.SetAppUnstakingHeightAndStatus(app2.Address, 1, persistence.UnstakingStatus)
	require.NoError(t, err)

	apps, err := db.GetAppsReadyToUnstake(0, persistence.UnstakingStatus)
	require.NoError(t, err)
	require.Equal(t, 1, len(apps), "wrong number of actors ready to unstake")
	require.Equal(t, app.Address, apps[0].Address, "unexpected application actor returned")

	apps, err = db.GetAppsReadyToUnstake(1, persistence.UnstakingStatus)
	require.NoError(t, err)
	require.Equal(t, 1, len(apps), "wrong number of actors ready to unstake")

	require.Equal(t, app2.Address, apps[0].Address, "unexpected application actor returned")
}

func TestGetAppStatus(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	app, err := newTestApp()
	require.NoError(t, err)

	err = db.InsertApp(
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
	require.NoError(t, err)

	status, err := db.GetAppStatus(app.Address, 0)
	require.NoError(t, err)
	require.Equal(t, status, DefaultStakeStatus, "unexpected status")

	// WTF
	status, err = db.GetAppStatus(app.Address, 1)
	require.NoError(t, err)
	require.Equal(t, status, DefaultStakeStatus, "unexpected status")
}

func TestGetPauseHeightIfExists(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	app, err := newTestApp()
	require.NoError(t, err)

	err = db.InsertApp(
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
	require.NoError(t, err)

	pauseHeight, err := db.GetAppPauseHeightIfExists(app.Address, 0)
	require.NoError(t, err)
	require.Equal(t, pauseHeight, DefaultPauseHeight, "unexpected pause height")

	// WTF
	pauseHeight, err = db.GetAppPauseHeightIfExists(app.Address, 1)
	require.NoError(t, err)
	require.Equal(t, pauseHeight, DefaultPauseHeight, "unexpected pause height")
}

func TestSetAppStatusAndUnstakingHeightPausedBefore(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	app, err := newTestApp()
	require.NoError(t, err)

	err = db.InsertApp(
		app.Address,
		app.PublicKey,
		app.Output,
		false,
		DefaultStakeStatus,
		DefaultMaxRelays,
		DefaultStake,
		DefaultChains,
		0,
		DefaultUnstakingHeight)
	require.NoError(t, err)

	unstakingHeightSet := int64(0)
	err = db.SetAppStatusAndUnstakingHeightPausedBefore(1, unstakingHeightSet, 1)
	require.NoError(t, err)

	_, _, _, _, _, unstakingHeight, _, _, err := db.GetApp(app.Address, db.Height)
	require.NoError(t, err)
	require.Equal(t, unstakingHeight, unstakingHeightSet, "unstaking height was not set correctly")
}

func TestSetAppPauseHeight(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	app, err := newTestApp()
	require.NoError(t, err)

	err = db.InsertApp(
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
	require.NoError(t, err)

	err = db.SetAppPauseHeight(app.Address, int64(PauseHeightToSet))
	require.NoError(t, err)

	_, _, _, _, _, pausedHeight, _, _, err := db.GetApp(app.Address, db.Height)
	require.NoError(t, err)
	require.Equal(t, pausedHeight, int64(PauseHeightToSet), "pause height not updated")
}

func TestGetAppOutputAddress(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	app, err := newTestApp()
	require.NoError(t, err)

	err = db.InsertApp(
		app.Address,
		app.PublicKey,
		app.Output,
		false,
		DefaultStakeStatus,
		DefaultMaxRelays,
		DefaultStake,
		DefaultChains,
		0, //DefaultPauseHeight, DISCUS(drewsky): Why are we not using `DefaultPauseHeight` here?
		DefaultUnstakingHeight)
	require.NoError(t, err)

	output, err := db.GetAppOutputAddress(app.Address, 0)
	require.NoError(t, err)
	require.Equal(t, output, app.Output, "unexpected output address")
}

func newTestApp() (*typesGenesis.App, error) {
	operatorKey, err := crypto.GeneratePublicKey()
	if err != nil {
		return nil, err
	}

	outputAddr, err := crypto.GenerateAddress()
	if err != nil {
		return nil, err
	}

	return &typesGenesis.App{
		Address:         operatorKey.Address(),
		PublicKey:       operatorKey.Bytes(),
		Paused:          false,
		Status:          typesGenesis.DefaultStakeStatus,
		Chains:          typesGenesis.DefaultChains,
		MaxRelays:       DefaultMaxRelays,
		StakedTokens:    typesGenesis.DefaultStake,
		PausedHeight:    uint64(DefaultPauseHeight),
		UnstakingHeight: DefaultUnstakingHeight,
		Output:          outputAddr,
	}, nil
}

func GetTestApp(db persistence.PostgresContext, address []byte) (*typesGenesis.App, error) {
	operator, publicKey, stakedTokens, maxRelays, outputAddress, pauseHeight, unstakingHeight, chains, err := db.GetApp(address, db.Height)
	if err != nil {
		return nil, err
	}

	addr, err := hex.DecodeString(operator)
	if err != nil {
		return nil, err
	}

	pubKey, err := hex.DecodeString(publicKey)
	if err != nil {
		return nil, err
	}

	outputAddr, err := hex.DecodeString(outputAddress)
	if err != nil {
		return nil, err
	}

	status := persistence.UndefinedStakingStatus
	switch unstakingHeight {
	case -1:
		status = persistence.StakedStatus
	case unstakingHeight:
		status = persistence.UnstakingStatus
	default:
		status = persistence.UnstakedStatus
	}

	return &typesGenesis.App{
		Address:         addr,
		PublicKey:       pubKey,
		Paused:          false,
		Status:          int32(status),
		Chains:          chains,
		MaxRelays:       maxRelays,
		StakedTokens:    stakedTokens,
		PausedHeight:    uint64(pauseHeight),
		UnstakingHeight: unstakingHeight,
		Output:          outputAddr,
	}, nil
}
