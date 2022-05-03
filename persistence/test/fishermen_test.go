package test

import (
	"bytes"
	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/shared/crypto"
	typesGenesis "github.com/pokt-network/pocket/shared/types/genesis"
	"testing"
)

func TestInsertFishermanAndExists(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	fisherman := NewTestFisherman()
	fisherman2 := NewTestFisherman()
	if err := db.InsertFisherman(fisherman.Address, fisherman.PublicKey, fisherman.Output, false, DefaultStakeStatus, DefaultStake, DefaultStake, DefaultChains, DefaultPauseHeight, DefaultUnstakingHeight); err != nil {
		t.Fatal(err)
	}
	exists, err := db.GetFishermanExists(fisherman.Address)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("actor that should exist does not")
	}
	exists, err = db.GetFishermanExists(fisherman2.Address)
	if err != nil {
		t.Fatal(err)
	}
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
	if err := db.InsertFisherman(fisherman.Address, fisherman.PublicKey, fisherman.Output, false, DefaultStakeStatus, DefaultStake, DefaultStake, DefaultChains, DefaultPauseHeight, DefaultUnstakingHeight); err != nil {
		t.Fatal(err)
	}
	_, _, stakedTokens, _, _, _, _, _, chains, err := db.GetFisherman(fisherman.Address)
	if err != nil {
		t.Fatal(err)
	}
	if err = db.UpdateFisherman(fisherman.Address, fisherman.ServiceUrl, StakeToUpdate, ChainsToUpdate); err != nil {
		t.Fatal(err)
	}
	_, _, stakedTokens, _, _, _, _, _, chains, err = db.GetFisherman(fisherman.Address)
	if err != nil {
		t.Fatal(err)
	}
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
	if err := db.InsertFisherman(fisherman.Address, fisherman.PublicKey, fisherman.Output, false, DefaultStakeStatus, DefaultStake, DefaultStake, DefaultChains, DefaultPauseHeight, DefaultUnstakingHeight); err != nil {
		t.Fatal(err)
	}
	_, _, _, _, _, _, _, _, chains, err := db.GetFisherman(fisherman.Address)
	if err != nil {
		t.Fatal(err)
	}
	if err = db.DeleteFisherman(fisherman.Address); err != nil {
		t.Fatal(err)
	}
	_, _, _, _, _, _, _, _, chains, err = db.GetFisherman(fisherman.Address)
	if err != nil {
		t.Fatal(err)
	}
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
	if err := db.InsertFisherman(fisherman.Address, fisherman.PublicKey, fisherman.Output, false, 1, DefaultStake, DefaultStake, DefaultChains, DefaultPauseHeight, DefaultUnstakingHeight); err != nil {
		t.Fatal(err)
	}
	// test SetFishermanUnstakingHeightAndStatus
	if err := db.SetFishermanUnstakingHeightAndStatus(fisherman.Address, 0, 1); err != nil {
		t.Fatal(err)
	}
	// test GetFishermansReadyToUnstake
	fishermans, err := db.GetFishermanReadyToUnstake(0, 1)
	if err != nil {
		t.Fatal(err)
	}
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
	if err := db.InsertFisherman(fisherman.Address, fisherman.PublicKey, fisherman.Output, false, 1, DefaultStake, DefaultStake, DefaultChains, DefaultPauseHeight, DefaultUnstakingHeight); err != nil {
		t.Fatal(err)
	}
	status, err := db.GetFishermanStatus(fisherman.Address)
	if err != nil {
		t.Fatal(err)
	}
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
	if err := db.InsertFisherman(fisherman.Address, fisherman.PublicKey, fisherman.Output, false, 1, DefaultStake, DefaultStake, DefaultChains, DefaultPauseHeight, DefaultUnstakingHeight); err != nil {
		t.Fatal(err)
	}
	height, err := db.GetFishermanPauseHeightIfExists(fisherman.Address)
	if err != nil {
		t.Fatal(err)
	}
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
	if err := db.InsertFisherman(fisherman.Address, fisherman.PublicKey, fisherman.Output, false, 1, DefaultStake, DefaultStake, DefaultChains, 0, DefaultUnstakingHeight); err != nil {
		t.Fatal(err)
	}
	if err := db.SetFishermansStatusAndUnstakingHeightPausedBefore(1, 0, 1); err != nil {
		t.Fatal(err)
	}
	_, _, _, _, _, _, unstakingHeight, _, _, err := db.GetFisherman(fisherman.Address)
	if err != nil {
		t.Fatal(err)
	}
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
	if err := db.InsertFisherman(fisherman.Address, fisherman.PublicKey, fisherman.Output, false, 1, DefaultStake, DefaultStake, DefaultChains, 0, DefaultUnstakingHeight); err != nil {
		t.Fatal(err)
	}
	if err := db.SetFishermanPauseHeight(fisherman.Address, int64(PauseHeightToSet)); err != nil {
		t.Fatal(err)
	}
	_, _, _, _, _, pauseHeight, _, _, _, err := db.GetFisherman(fisherman.Address)
	if err != nil {
		t.Fatal(err)
	}
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
	if err := db.InsertFisherman(fisherman.Address, fisherman.PublicKey, fisherman.Output, false, 1, DefaultStake, DefaultStake, DefaultChains, 0, DefaultUnstakingHeight); err != nil {
		t.Fatal(err)
	}
	output, err := db.GetFishermanOutputAddress(fisherman.Address)
	if err != nil {
		t.Fatal(err)
	}
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
