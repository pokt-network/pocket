package test

import (
	"bytes"
	"encoding/hex"
	"testing"

	query "github.com/pokt-network/pocket/persistence/schema"

	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/shared/crypto"
	typesGenesis "github.com/pokt-network/pocket/shared/types/genesis"
	"github.com/stretchr/testify/require"
)

func FuzzFishermen(f *testing.F) {
	fuzzProtocolActor(f,
		NewTestGenericActor(newTestFisherman),
		GetGenericActor(GetTestFisherman),
		query.FishermanActor)

}

func TestInsertFishermanAndExists(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	fisherman := NewTestFisherman(t)
	fisherman2 := NewTestFisherman(t)
	err := db.InsertFisherman(fisherman.Address, fisherman.PublicKey, fisherman.Output, false, DefaultStakeStatus, DefaultStake, DefaultStake, DefaultChains, DefaultPauseHeight, DefaultUnstakingHeight)
	require.NoError(t, err)
	exists, err := db.GetFishermanExists(fisherman.Address, db.Height)
	require.NoError(t, err)
	if !exists {
		t.Fatal("actor that should exist does not")
	}
	exists, err = db.GetFishermanExists(fisherman2.Address, db.Height)
	require.NoError(t, err)
	if exists {
		t.Fatal("actor that should not exist, appears to")
	}
}

func TestUpdateFisherman(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	fisherman := NewTestFisherman(t)
	err := db.InsertFisherman(fisherman.Address, fisherman.PublicKey, fisherman.Output, false, DefaultStakeStatus, DefaultStake, DefaultStake, DefaultChains, DefaultPauseHeight, DefaultUnstakingHeight)
	require.NoError(t, err)
	_, _, stakedTokens, _, _, _, _, chains, err := db.GetFisherman(fisherman.Address, db.Height)
	require.NoError(t, err)
	err = db.UpdateFisherman(fisherman.Address, fisherman.ServiceUrl, StakeToUpdate, ChainsToUpdate)
	require.NoError(t, err)
	_, _, stakedTokens, _, _, _, _, chains, err = db.GetFisherman(fisherman.Address, db.Height)
	require.NoError(t, err)
	if chains[0] != ChainsToUpdate[0] {
		t.Fatal("chains not updated")
	}
	if stakedTokens != StakeToUpdate {
		t.Fatal("stake not updated")
	}
}

func TestDeleteFisherman(t *testing.T) {
	//db := persistence.PostgresContext{ DEPRECATED NO OP
	//	Height: 0,
	//	DB:     *PostgresDB,
	//}
	//fisherman := NewTestFisherman(t)
	//err := db.InsertFisherman(fisherman.Address, fisherman.PublicKey, fisherman.Output, false, DefaultStakeStatus, DefaultStake, DefaultStake, DefaultChains, DefaultPauseHeight, DefaultUnstakingHeight)
	//require.NoError(t, err)
	//_, _, _, _, _, _, _, chains, err := db.GetFisherman(fisherman.Address, db.Height)
	//require.NoError(t, err)
	//err = db.DeleteFisherman(fisherman.Address)
	//require.NoError(t, err)
	//_, _, _, _, _, _, _, chains, err = db.GetFisherman(fisherman.Address, db.Height)
	//require.NoError(t, err)
	//if len(chains) != 0 {
	//	t.Fatal("chains not nullified")
	//}
}

func TestGetFishermansReadyToUnstake(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	fisherman := NewTestFisherman(t)
	db.ClearAllDebug()
	err := db.InsertFisherman(fisherman.Address, fisherman.PublicKey, fisherman.Output, false, 1, DefaultStake, DefaultStake, DefaultChains, DefaultPauseHeight, DefaultUnstakingHeight)
	require.NoError(t, err)
	// test SetFishermanUnstakingHeightAndStatus
	err = db.SetFishermanUnstakingHeightAndStatus(fisherman.Address, 0, 1)
	require.NoError(t, err)
	// test GetFishermansReadyToUnstake
	fishermans, err := db.GetFishermanReadyToUnstake(0, 1)
	require.NoError(t, err)
	if len(fishermans) != 1 {
		t.Fatal("wrong number of actors")
	}
	if !bytes.Equal(fisherman.Address, fishermans[0].Address) {
		t.Fatal("unexpected actor returned")
	}
}

func TestGetFishermanStatus(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	fisherman := NewTestFisherman(t)
	err := db.InsertFisherman(fisherman.Address, fisherman.PublicKey, fisherman.Output, false, 1, DefaultStake, DefaultStake, DefaultChains, DefaultPauseHeight, DefaultUnstakingHeight)
	require.NoError(t, err)
	status, err := db.GetFishermanStatus(fisherman.Address, db.Height)
	require.NoError(t, err)
	if status != DefaultStakeStatus {
		t.Fatalf("unexpected status: got %d expected %d", status, DefaultStakeStatus)
	}
}

func TestGetFishermanPauseHeightIfExists(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	fisherman := NewTestFisherman(t)
	err := db.InsertFisherman(fisherman.Address, fisherman.PublicKey, fisherman.Output, false, 1, DefaultStake, DefaultStake, DefaultChains, DefaultPauseHeight, DefaultUnstakingHeight)
	require.NoError(t, err)
	height, err := db.GetFishermanPauseHeightIfExists(fisherman.Address, db.Height)
	require.NoError(t, err)
	if height != DefaultPauseHeight {
		t.Fatalf("unexpected pausedHeight: got %d expected %d", DefaultPauseHeight, DefaultStakeStatus)
	}
}

func TestSetFishermansStatusAndUnstakingHeightPausedBefore(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	fisherman := NewTestFisherman(t)
	err := db.InsertFisherman(fisherman.Address, fisherman.PublicKey, fisherman.Output, false, 1, DefaultStake, DefaultStake, DefaultChains, 0, DefaultUnstakingHeight)
	require.NoError(t, err)
	err = db.SetFishermansStatusAndUnstakingHeightPausedBefore(1, 0, 1)
	require.NoError(t, err)
	_, _, _, _, _, _, unstakingHeight, _, err := db.GetFisherman(fisherman.Address, db.Height)
	require.NoError(t, err)
	if unstakingHeight != 0 {
		t.Fatal("unexpected unstaking height")
	}
}

func TestSetFishermanPauseHeight(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	fisherman := NewTestFisherman(t)
	err := db.InsertFisherman(fisherman.Address, fisherman.PublicKey, fisherman.Output, false, 1, DefaultStake, DefaultStake, DefaultChains, 0, DefaultUnstakingHeight)
	require.NoError(t, err)
	err = db.SetFishermanPauseHeight(fisherman.Address, int64(PauseHeightToSet))
	require.NoError(t, err)
	_, _, _, _, _, pausedHeight, _, _, err := db.GetFisherman(fisherman.Address, db.Height)
	require.NoError(t, err)
	if pausedHeight != int64(PauseHeightToSet) {
		t.Fatal("unexpected pause height")
	}
}

func TestGetFishermanOutputAddress(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	fisherman := NewTestFisherman(t)
	err := db.InsertFisherman(fisherman.Address, fisherman.PublicKey, fisherman.Output, false, 1, DefaultStake, DefaultStake, DefaultChains, 0, DefaultUnstakingHeight)
	require.NoError(t, err)
	output, err := db.GetFishermanOutputAddress(fisherman.Address, db.Height)
	require.NoError(t, err)
	if !bytes.Equal(output, fisherman.Output) {
		t.Fatal("unexpected output address")
	}
}

func NewTestFisherman(t *testing.T) typesGenesis.Fisherman {
	fish, err := newTestFisherman()
	require.NoError(t, err)
	return fish
}

func newTestFisherman() (typesGenesis.Fisherman, error) {
	pubKey, err := crypto.GeneratePublicKey()
	if err != nil {
		return typesGenesis.Fisherman{}, err
	}
	outputAddr, err := crypto.GenerateAddress()
	if err != nil {
		return typesGenesis.Fisherman{}, err
	}
	return typesGenesis.Fisherman{
		Address:         pubKey.Address(),
		PublicKey:       pubKey.Bytes(),
		Paused:          false,
		Status:          typesGenesis.DefaultStakeStatus,
		Chains:          typesGenesis.DefaultChains,
		ServiceUrl:      DefaultServiceUrl,
		StakedTokens:    typesGenesis.DefaultStake,
		PausedHeight:    uint64(DefaultPauseHeight),
		UnstakingHeight: DefaultUnstakingHeight,
		Output:          outputAddr,
	}, nil
}

func GetTestFisherman(db persistence.PostgresContext, address []byte) (*typesGenesis.Fisherman, error) {
	operator, publicKey, stakedTokens, serviceURL, outputAddress, pauseHeight, unstakingHeight, chains, err := db.GetFisherman(address, db.Height)
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
	return &typesGenesis.Fisherman{
		Address:         addr,
		PublicKey:       pubKey,
		Paused:          false,
		Status:          int32(status),
		Chains:          chains,
		ServiceUrl:      serviceURL,
		StakedTokens:    stakedTokens,
		PausedHeight:    uint64(pauseHeight),
		UnstakingHeight: unstakingHeight,
		Output:          outputAddr,
	}, nil
}
