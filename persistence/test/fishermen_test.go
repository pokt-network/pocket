package test

import (
	"bytes"
	"testing"

	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/shared/crypto"
	typesGenesis "github.com/pokt-network/pocket/shared/types/genesis"
	"github.com/stretchr/testify/require"
)

func TestInsertFishermanAndExists(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	fisherman := NewTestFisherman()
	fisherman2 := NewTestFisherman()
	err := db.InsertFisherman(fisherman.Address, fisherman.PublicKey, fisherman.Output, false, DefaultStakeStatus, DefaultStake, DefaultStake, DefaultChains, DefaultPauseHeight, DefaultUnstakingHeight)
	require.NoError(t, err)
	exists, err := db.GetFishermanExists(fisherman.Address)
	require.NoError(t, err)
	if !exists {
		t.Fatal("actor that should exist does not")
	}
	exists, err = db.GetFishermanExists(fisherman2.Address)
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
	fisherman := NewTestFisherman()
	err := db.InsertFisherman(fisherman.Address, fisherman.PublicKey, fisherman.Output, false, DefaultStakeStatus, DefaultStake, DefaultStake, DefaultChains, DefaultPauseHeight, DefaultUnstakingHeight)
	require.NoError(t, err)
	_, _, stakedTokens, _, _, _, _, _, chains, err := db.GetFisherman(fisherman.Address)
	require.NoError(t, err)
	err = db.UpdateFisherman(fisherman.Address, fisherman.ServiceUrl, StakeToUpdate, ChainsToUpdate)
	require.NoError(t, err)
	_, _, stakedTokens, _, _, _, _, _, chains, err = db.GetFisherman(fisherman.Address)
	require.NoError(t, err)
	if chains[0] != ChainsToUpdate[0] {
		t.Fatal("chains not updated")
	}
	if stakedTokens != StakeToUpdate {
		t.Fatal("stake not updated")
	}
}

func TestDeleteFisherman(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	fisherman := NewTestFisherman()
	err := db.InsertFisherman(fisherman.Address, fisherman.PublicKey, fisherman.Output, false, DefaultStakeStatus, DefaultStake, DefaultStake, DefaultChains, DefaultPauseHeight, DefaultUnstakingHeight)
	require.NoError(t, err)
	_, _, _, _, _, _, _, _, chains, err := db.GetFisherman(fisherman.Address)
	require.NoError(t, err)
	err = db.DeleteFisherman(fisherman.Address)
	require.NoError(t, err)
	_, _, _, _, _, _, _, _, chains, err = db.GetFisherman(fisherman.Address)
	require.NoError(t, err)
	if len(chains) != 0 {
		t.Fatal("chains not nullified")
	}
}

func TestGetFishermansReadyToUnstake(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	fisherman := NewTestFisherman()
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
	fisherman := NewTestFisherman()
	err := db.InsertFisherman(fisherman.Address, fisherman.PublicKey, fisherman.Output, false, 1, DefaultStake, DefaultStake, DefaultChains, DefaultPauseHeight, DefaultUnstakingHeight)
	require.NoError(t, err)
	status, err := db.GetFishermanStatus(fisherman.Address)
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
	fisherman := NewTestFisherman()
	err := db.InsertFisherman(fisherman.Address, fisherman.PublicKey, fisherman.Output, false, 1, DefaultStake, DefaultStake, DefaultChains, DefaultPauseHeight, DefaultUnstakingHeight)
	require.NoError(t, err)
	height, err := db.GetFishermanPauseHeightIfExists(fisherman.Address)
	require.NoError(t, err)
	if height != DefaultPauseHeight {
		t.Fatalf("unexpected pauseHeight: got %d expected %d", DefaultPauseHeight, DefaultStakeStatus)
	}
}

func TestSetFishermansStatusAndUnstakingHeightPausedBefore(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	fisherman := NewTestFisherman()
	err := db.InsertFisherman(fisherman.Address, fisherman.PublicKey, fisherman.Output, false, 1, DefaultStake, DefaultStake, DefaultChains, 0, DefaultUnstakingHeight)
	require.NoError(t, err)
	err = db.SetFishermansStatusAndUnstakingHeightPausedBefore(1, 0, 1)
	require.NoError(t, err)
	_, _, _, _, _, _, unstakingHeight, _, _, err := db.GetFisherman(fisherman.Address)
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
	fisherman := NewTestFisherman()
	err := db.InsertFisherman(fisherman.Address, fisherman.PublicKey, fisherman.Output, false, 1, DefaultStake, DefaultStake, DefaultChains, 0, DefaultUnstakingHeight)
	require.NoError(t, err)
	err = db.SetFishermanPauseHeight(fisherman.Address, int64(PauseHeightToSet))
	require.NoError(t, err)
	_, _, _, _, _, pauseHeight, _, _, _, err := db.GetFisherman(fisherman.Address)
	require.NoError(t, err)
	if pauseHeight != int64(PauseHeightToSet) {
		t.Fatal("unexpected pause height")
	}
}

func TestGetFishermanOutputAddress(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	fisherman := NewTestFisherman()
	err := db.InsertFisherman(fisherman.Address, fisherman.PublicKey, fisherman.Output, false, 1, DefaultStake, DefaultStake, DefaultChains, 0, DefaultUnstakingHeight)
	require.NoError(t, err)
	output, err := db.GetFishermanOutputAddress(fisherman.Address)
	require.NoError(t, err)
	if !bytes.Equal(output, fisherman.Output) {
		t.Fatal("unexpected output address")
	}
}

func NewTestFisherman() typesGenesis.Fisherman {
	pub1, _ := crypto.GeneratePublicKey()
	addr1 := pub1.Address()
	addr2, _ := crypto.GenerateAddress()
	return typesGenesis.Fisherman{
		Address:         addr1,
		PublicKey:       pub1.Bytes(),
		Paused:          false,
		Status:          typesGenesis.DefaultStakeStatus,
		Chains:          typesGenesis.DefaultChains,
		ServiceUrl:      DefaultServiceUrl,
		StakedTokens:    typesGenesis.DefaultStake,
		PausedHeight:    0,
		UnstakingHeight: 0,
		Output:          addr2,
	}
}
