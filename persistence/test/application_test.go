package test

import (
	"testing"

	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/shared/crypto"
	typesGenesis "github.com/pokt-network/pocket/shared/types/genesis"
	"github.com/stretchr/testify/require"
)

func TestInsertAppAndExists(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	app := NewTestApp(t)
	app2 := NewTestApp(t)

	err := db.InsertApp(app.Address, app.PublicKey, app.Output, false, DefaultStakeStatus, DefaultMaxRelays, DefaultStake, DefaultChains, DefaultPauseHeight, DefaultUnstakingHeight)
	require.NoError(t, err)

	exists, err := db.GetAppExists(app.Address)
	require.NoError(t, err)
	require.True(t, exists, "actor that should exist does not")

	exists, err = db.GetAppExists(app2.Address)
	require.NoError(t, err)
	require.False(t, exists, "actor that should not exist, appears to")
}

func TestUpdateApp(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	app := NewTestApp(t)

	err := db.InsertApp(app.Address, app.PublicKey, app.Output, false, DefaultStakeStatus, DefaultMaxRelays, DefaultStake, DefaultChains, DefaultPauseHeight, DefaultUnstakingHeight)
	require.NoError(t, err)

	_, _, stakedTokens, _, _, _, _, _, chains, err := db.GetApp(app.Address)
	require.NoError(t, err)

	require.Equal(t, chains, DefaultChains, "default chains incorrect")
	require.Equal(t, stakedTokens, DefaultStake, "default stake incorrect")

	err = db.UpdateApp(app.Address, app.MaxRelays, StakeToUpdate, ChainsToUpdate)
	require.NoError(t, err)

	_, _, stakedTokens, _, _, _, _, _, chains, err = db.GetApp(app.Address)
	require.NoError(t, err)

	require.Equal(t, chains, ChainsToUpdate, "chains not updated")
	require.Equal(t, stakedTokens, StakeToUpdate, "stake not updated")
}

func TestDeleteApp(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	app := NewTestApp(t)

	err := db.InsertApp(app.Address, app.PublicKey, app.Output, false, DefaultStakeStatus, DefaultMaxRelays, DefaultStake, DefaultChains, DefaultPauseHeight, DefaultUnstakingHeight)
	require.NoError(t, err)

	_, _, _, _, _, _, _, _, chains, err := db.GetApp(app.Address)
	require.NoError(t, err)
	require.NotEmpty(t, chains, "chains should not be empty but are")

	err = db.DeleteApp(app.Address)
	require.NoError(t, err)

	_, _, _, _, _, _, _, _, chains, err = db.GetApp(app.Address)
	require.Empty(t, chains, "chains should be nullified but are not")
	require.Error(t, err) // DISCUSS(drewsky): A change I made to `GetApp` made this return an error after `DeleteApp`. What's the expected behavior?
}

func TestGetAppsReadyToUnstake(t *testing.T) {
	height := int64(0)

	db := persistence.PostgresContext{
		Height: height,
		DB:     *PostgresDB,
	}
	app := NewTestApp(t)

	err := db.InsertApp(app.Address, app.PublicKey, app.Output, false, DefaultStakeStatus, DefaultStake, DefaultStake, DefaultChains, DefaultPauseHeight, DefaultUnstakingHeight)
	require.NoError(t, err)

	err = db.SetAppUnstakingHeightAndStatus(app.Address, height, persistence.UnstakingStatus)
	require.NoError(t, err)

	apps, err := db.GetAppsReadyToUnstake(height, persistence.UnstakingStatus)
	require.NoError(t, err)

	require.Equal(t, 1, len(apps), "wrong number of actors: apps should be 1 but are not")
	require.Equal(t, app.Address, apps[0].Address, "unexpected application actor returned")
}

func TestGetAppStatus(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	app := NewTestApp(t)

	err := db.InsertApp(app.Address, app.PublicKey, app.Output, false, DefaultStakeStatus, DefaultStake, DefaultStake, DefaultChains, DefaultPauseHeight, DefaultUnstakingHeight)
	require.NoError(t, err)

	status, err := db.GetAppStatus(app.Address)
	require.NoError(t, err)
	require.Equal(t, status, DefaultStakeStatus, "unexpected status: got %d expected %d", status, DefaultStakeStatus)
}

func TestGetPauseHeightIfExists(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	app := NewTestApp(t)
	err := db.InsertApp(app.Address, app.PublicKey, app.Output, false, DefaultStakeStatus, DefaultStake, DefaultStake, DefaultChains, DefaultPauseHeight, DefaultUnstakingHeight)
	require.NoError(t, err)

	height, err := db.GetAppPauseHeightIfExists(app.Address)
	require.NoError(t, err)
	require.Equal(t, height, DefaultPauseHeight, "unexpected pause height")
}

func TestSetAppsStatusAndUnstakingHeightPausedBefore(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	app := NewTestApp(t)

	// DISCUS(drewsky): Why are we not using `DefaultPauseHeight` here?
	err := db.InsertApp(app.Address, app.PublicKey, app.Output, false, DefaultStakeStatus, DefaultStake, DefaultStake, DefaultChains, 0, DefaultUnstakingHeight)
	require.NoError(t, err)

	unstakingHeightSet := int64(0)
	err = db.SetAppsStatusAndUnstakingHeightPausedBefore(1, unstakingHeightSet, 1)
	require.NoError(t, err)

	_, _, _, _, _, _, unstakingHeight, _, _, err := db.GetApp(app.Address)
	require.NoError(t, err)
	require.Equal(t, unstakingHeight, unstakingHeightSet, "unstaking height was not set correctly")
}

func TestSetAppPauseHeight(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	app := NewTestApp(t)

	err := db.InsertApp(app.Address, app.PublicKey, app.Output, false, DefaultStakeStatus, DefaultStake, DefaultStake, DefaultChains, DefaultPauseHeight, DefaultUnstakingHeight)
	require.NoError(t, err)

	err = db.SetAppPauseHeight(app.Address, int64(PauseHeightToSet))
	require.NoError(t, err)

	_, _, _, _, _, pausedHeight, _, _, _, err := db.GetApp(app.Address)
	require.NoError(t, err)
	require.Equal(t, pausedHeight, int64(PauseHeightToSet), "pause height not updated")
}

func TestGetAppOutputAddress(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	app := NewTestApp(t)

	// DISCUS(drewsky): Why are we not using `DefaultPauseHeight` here?
	err := db.InsertApp(app.Address, app.PublicKey, app.Output, false, DefaultStakeStatus, DefaultStake, DefaultStake, DefaultChains, 0, DefaultUnstakingHeight)
	require.NoError(t, err)

	output, err := db.GetAppOutputAddress(app.Address)
	require.NoError(t, err)
	require.Equal(t, output, app.Output, "unexpected output address")
}

func NewTestApp(t *testing.T) typesGenesis.App {
	pub1, err := crypto.GeneratePublicKey()
	require.NoError(t, err)
	addr1 := pub1.Address()

	addr2, err := crypto.GenerateAddress()
	require.NoError(t, err)

	return typesGenesis.App{
		Address:         addr1,
		PublicKey:       pub1.Bytes(),
		Paused:          false,
		Status:          typesGenesis.DefaultStakeStatus,
		Chains:          typesGenesis.DefaultChains,
		MaxRelays:       DefaultMaxRelays,
		StakedTokens:    typesGenesis.DefaultStake,
		PausedHeight:    uint64(DefaultPauseHeight),
		UnstakingHeight: DefaultUnstakingHeight,
		Output:          addr2,
	}
}
