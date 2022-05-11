package test

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types"
	typesGenesis "github.com/pokt-network/pocket/shared/types/genesis"
	"github.com/stretchr/testify/require"
)

func TestInsertAppAndExists(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	app := NewTestApp()
	app2 := NewTestApp()
	err := db.InsertApp(app.Address, app.PublicKey, app.Output, false, DefaultStakeStatus, DefaultStake, DefaultStake, DefaultChains, DefaultPauseHeight, DefaultUnstakingHeight)
	require.NoError(t, err)
	exists, err := db.GetAppExists(app.Address)
	require.NoError(t, err)
	if !exists {
		t.Fatal("actor that should exist does not")
	}
	exists, err = db.GetAppExists(app2.Address)
	require.NoError(t, err)
	if exists {
		t.Fatal("actor that should not exist, appears to")
	}
}

func TestUpdateApp(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	app := NewTestApp()
	err := db.InsertApp(app.Address, app.PublicKey, app.Output, false, DefaultStakeStatus, DefaultStake, DefaultStake, DefaultChains, DefaultPauseHeight, DefaultUnstakingHeight)
	require.NoError(t, err)
	_, _, stakedTokens, _, _, _, _, _, chains, err := db.GetApp(app.Address)
	require.NoError(t, err)
	err = db.UpdateApp(app.Address, app.MaxRelays, StakeToUpdate, ChainsToUpdate)
	require.NoError(t, err)
	_, _, stakedTokens, _, _, _, _, _, chains, err = db.GetApp(app.Address)
	require.NoError(t, err)
	if chains[0] != ChainsToUpdate[0] {
		t.Fatal("chains not updated")
	}
	if stakedTokens != StakeToUpdate {
		t.Fatal("stake not updated")
	}
}

func TestDeleteApp(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	app := NewTestApp()
	err := db.InsertApp(app.Address, app.PublicKey, app.Output, false, DefaultStakeStatus, DefaultStake, DefaultStake, DefaultChains, DefaultPauseHeight, DefaultUnstakingHeight)
	require.NoError(t, err)
	_, _, _, _, _, _, _, _, chains, err := db.GetApp(app.Address)
	require.NoError(t, err)
	err = db.DeleteApp(app.Address)
	require.NoError(t, err)
	_, _, _, _, _, _, _, _, chains, err = db.GetApp(app.Address)
	require.NoError(t, err)
	if len(chains) != 0 {
		t.Fatal("chains not nullified")
	}
}

func TestGetAppsReadyToUnstake(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	app := NewTestApp()
	err := db.InsertApp(app.Address, app.PublicKey, app.Output, false, 1, DefaultStake, DefaultStake, DefaultChains, DefaultPauseHeight, DefaultUnstakingHeight)
	require.NoError(t, err)
	// test SetAppUnstakingHeightAndStatus
	err = db.SetAppUnstakingHeightAndStatus(app.Address, 0, 1)
	require.NoError(t, err)
	// test GetAppsReadyToUnstake
	apps, err := db.GetAppsReadyToUnstake(0, 1)
	require.NoError(t, err)
	if len(apps) != 1 {
		t.Fatal("wrong number of actors")
	}
	if !bytes.Equal(app.Address, apps[0].Address) {
		t.Fatal("unexpected actor returned")
	}
}

func TestGetAppStatus(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	app := NewTestApp()
	err := db.InsertApp(app.Address, app.PublicKey, app.Output, false, 1, DefaultStake, DefaultStake, DefaultChains, DefaultPauseHeight, DefaultUnstakingHeight)
	require.NoError(t, err)
	status, err := db.GetAppStatus(app.Address)
	require.NoError(t, err)
	if status != DefaultStakeStatus {
		t.Fatalf("unexpected status: got %d expected %d", status, DefaultStakeStatus)
	}
}

func TestGetPauseHeightIfExists(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	app := NewTestApp()
	err := db.InsertApp(app.Address, app.PublicKey, app.Output, false, 1, DefaultStake, DefaultStake, DefaultChains, DefaultPauseHeight, DefaultUnstakingHeight)
	require.NoError(t, err)
	height, err := db.GetAppPauseHeightIfExists(app.Address)
	require.NoError(t, err)
	if height != DefaultPauseHeight {
		t.Fatalf("unexpected pauseHeight: got %d expected %d", DefaultPauseHeight, DefaultStakeStatus)
	}
}

func TestSetAppsStatusAndUnstakingHeightPausedBefore(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	app := NewTestApp()
	err := db.InsertApp(app.Address, app.PublicKey, app.Output, false, 1, DefaultStake, DefaultStake, DefaultChains, 0, DefaultUnstakingHeight)
	require.NoError(t, err)
	err = db.SetAppsStatusAndUnstakingHeightPausedBefore(1, 0, 1)
	require.NoError(t, err)
	_, _, _, _, _, _, unstakingHeight, _, _, err := db.GetApp(app.Address)
	require.NoError(t, err)
	if unstakingHeight != 0 {
		t.Fatal("unexpected unstaking height")
	}
}

func TestSetAppPauseHeight(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	app := NewTestApp()
	err := db.InsertApp(app.Address, app.PublicKey, app.Output, false, 1, DefaultStake, DefaultStake, DefaultChains, DefaultPauseHeight, DefaultUnstakingHeight)
	require.NoError(t, err)
	err = db.SetAppPauseHeight(app.Address, int64(PauseHeightToSet))
	require.NoError(t, err)
	_, _, _, _, _, pauseHeight, _, _, _, err := db.GetApp(app.Address)
	require.NoError(t, err)
	if pauseHeight != int64(PauseHeightToSet) {
		t.Fatal("unexpected pause height")
	}
}

func TestGetAppOutputAddress(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	app := NewTestApp()
	err := db.InsertApp(app.Address, app.PublicKey, app.Output, false, 1, DefaultStake, DefaultStake, DefaultChains, 0, DefaultUnstakingHeight)
	require.NoError(t, err)
	output, err := db.GetAppOutputAddress(app.Address)
	require.NoError(t, err)
	if !bytes.Equal(output, app.Output) {
		t.Fatal("unexpected output address")
	}
}

func NewTestApp() typesGenesis.App {
	pub1, _ := crypto.GeneratePublicKey()
	addr1 := pub1.Address()
	addr2, _ := crypto.GenerateAddress()
	defaultMaxRelays := types.BigIntToString(big.NewInt(1000000))
	return typesGenesis.App{
		Address:         addr1,
		PublicKey:       pub1.Bytes(),
		Paused:          false,
		Status:          typesGenesis.DefaultStakeStatus,
		Chains:          typesGenesis.DefaultChains,
		MaxRelays:       defaultMaxRelays,
		StakedTokens:    typesGenesis.DefaultStake,
		PausedHeight:    0,
		UnstakingHeight: 0,
		Output:          addr2,
	}
}
