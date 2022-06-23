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
	fuzzActor(f, newTestGenericApp, query.InsertAppQuery, GetGenericApp, false, query.UpdateAppQuery,
		query.UpdateAppChainsQuery, query.AppChainsTableName, query.AppsReadyToUnstakeQuery,
		query.AppUnstakingHeightQuery, query.AppPauseHeightQuery, query.AppQuery, query.AppChainsQuery,
		query.UpdateAppUnstakingHeightQuery, query.UpdateAppPausedHeightQuery, query.UpdateAppsPausedBefore,
		query.AppOutputAddressQuery)
}

func TestInsertAppAndExists(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	app := NewTestApp(t)
	app2 := NewTestApp(t)

	err := db.InsertApp(app.Address, app.PublicKey, app.Output, false, DefaultStakeStatus, DefaultMaxRelays, DefaultStake, DefaultChains, DefaultPauseHeight, DefaultUnstakingHeight)
	require.NoError(t, err)

	exists, err := db.GetAppExists(app.Address, db.Height)
	require.NoError(t, err)
	require.True(t, exists, "actor that should exist does not")

	exists, err = db.GetAppExists(app2.Address, db.Height)
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

	_, _, stakedTokens, _, _, _, _, chains, err := db.GetApp(app.Address, db.Height)
	require.NoError(t, err)

	require.Equal(t, chains, DefaultChains, "default chains incorrect")
	require.Equal(t, stakedTokens, DefaultStake, "default stake incorrect")

	err = db.UpdateApp(app.Address, app.MaxRelays, StakeToUpdate, ChainsToUpdate)
	require.NoError(t, err)

	_, _, stakedTokens, _, _, _, _, chains, err = db.GetApp(app.Address, db.Height)
	require.NoError(t, err)

	require.Equal(t, chains, ChainsToUpdate, "chains not updated")
	require.Equal(t, stakedTokens, StakeToUpdate, "stake not updated")
}

func TestGetAppsReadyToUnstake(t *testing.T) {
	height := int64(0)

	db := persistence.PostgresContext{
		Height: height,
		DB:     *PostgresDB,
	}
	db.ClearAllDebug()
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

	status, err := db.GetAppStatus(app.Address, 0)
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

	height, err := db.GetAppPauseHeightIfExists(app.Address, 0)
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

	_, _, _, _, _, unstakingHeight, _, _, err := db.GetApp(app.Address, db.Height)
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

	_, _, _, _, _, pausedHeight, _, _, err := db.GetApp(app.Address, db.Height)
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

	output, err := db.GetAppOutputAddress(app.Address, 0)
	require.NoError(t, err)
	require.Equal(t, output, app.Output, "unexpected output address")
}

func NewTestApp(t *testing.T) typesGenesis.App {
	app, err := newTestApp()
	require.NoError(t, err)
	return app
}

func newTestApp() (typesGenesis.App, error) {
	pub1, err := crypto.GeneratePublicKey()
	if err != nil {
		return typesGenesis.App{}, err
	}
	addr1 := pub1.Address()

	addr2, err := crypto.GenerateAddress()
	if err != nil {
		return typesGenesis.App{}, err
	}
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
	}, nil
}

func newTestGenericApp() (query.GenericActor, error) {
	app, err := newTestApp()
	if err != nil {
		return query.GenericActor{}, err
	}
	return query.GenericActor{
		Address:         hex.EncodeToString(app.Address),
		PublicKey:       hex.EncodeToString(app.PublicKey),
		StakedTokens:    app.StakedTokens,
		GenericParam:    app.MaxRelays,
		OutputAddress:   hex.EncodeToString(app.Output),
		PausedHeight:    int64(app.PausedHeight),
		UnstakingHeight: app.UnstakingHeight,
		Chains:          app.Chains,
	}, nil
}

func GetGenericApp(db persistence.PostgresContext, address string) (*query.GenericActor, error) {
	addr, err := hex.DecodeString(address)
	if err != nil {
		return nil, err
	}
	app, err := GetTestApp(db, addr)
	if err != nil {
		return nil, err
	}
	return &query.GenericActor{
		Address:         hex.EncodeToString(app.Address),
		PublicKey:       hex.EncodeToString(app.PublicKey),
		StakedTokens:    app.StakedTokens,
		GenericParam:    app.MaxRelays,
		OutputAddress:   hex.EncodeToString(app.Output),
		PausedHeight:    int64(app.PausedHeight),
		UnstakingHeight: app.UnstakingHeight,
		Chains:          app.Chains,
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
	status := -1
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
