package test

import (
	"bytes"
	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/shared/crypto"
	typesGenesis "github.com/pokt-network/pocket/shared/types/genesis"
	"testing"
)

func TestInsertValidatorAndExists(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	validator := NewTestValidator()
	validator2 := NewTestValidator()
	if err := db.InsertValidator(validator.Address, validator.PublicKey, validator.Output, false, DefaultStakeStatus, DefaultStake, DefaultStake, DefaultPauseHeight, DefaultUnstakingHeight); err != nil {
		t.Fatal(err)
	}
	exists, err := db.GetValidatorExists(validator.Address)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("actor that should exist does not")
	}
	exists, err = db.GetValidatorExists(validator2.Address)
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal("actor that should not exist, appears to")
	}
}

func TestUpdateValidator(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	validator := NewTestValidator()
	if err := db.InsertValidator(validator.Address, validator.PublicKey, validator.Output, false, DefaultStakeStatus, DefaultStake, DefaultStake, DefaultPauseHeight, DefaultUnstakingHeight); err != nil {
		t.Fatal(err)
	}
	_, _, stakedTokens, _, _, _, _, _, err := db.GetValidator(validator.Address)
	if err != nil {
		t.Fatal(err)
	}
	if err = db.UpdateValidator(validator.Address, validator.ServiceUrl, StakeToUpdate); err != nil {
		t.Fatal(err)
	}
	_, _, stakedTokens, _, _, _, _, _, err = db.GetValidator(validator.Address)
	if err != nil {
		t.Fatal(err)
	}
	if stakedTokens != StakeToUpdate {
		t.Fatal("stake not updated")
	}
}

func TestDeleteValidator(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	validator := NewTestValidator()
	if err := db.InsertValidator(validator.Address, validator.PublicKey, validator.Output, false, DefaultStakeStatus, DefaultStake, DefaultStake, DefaultPauseHeight, DefaultUnstakingHeight); err != nil {
		t.Fatal(err)
	}
	_, _, _, _, _, _, _, _, err := db.GetValidator(validator.Address)
	if err != nil {
		t.Fatal(err)
	}
	if err = db.DeleteValidator(validator.Address); err != nil {
		t.Fatal(err)
	}
	addr, _, _, _, _, _, _, _, err := db.GetValidator(validator.Address)
	if err != nil {
		t.Fatal(err)
	}
	if len(addr) != 0 {
		t.Fatal("validator not nullified")
	}
}

func TestGetValidatorsReadyToUnstake(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	validator := NewTestValidator()
	if err := db.InsertValidator(validator.Address, validator.PublicKey, validator.Output, false, 1, DefaultStake, DefaultStake, DefaultPauseHeight, DefaultUnstakingHeight); err != nil {
		t.Fatal(err)
	}
	// test SetValidatorUnstakingHeightAndStatus
	if err := db.SetValidatorUnstakingHeightAndStatus(validator.Address, 0, 1); err != nil {
		t.Fatal(err)
	}
	// test GetValidatorsReadyToUnstake
	validators, err := db.GetValidatorsReadyToUnstake(0, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(validators) != 1 {
		t.Fatal("wrong number of actors")
	}
	if !bytes.Equal(validator.Address, validators[0].Address) {
		t.Fatal("unexpected actor returned")
	}
}

func TestGetValidatorStatus(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	validator := NewTestValidator()
	if err := db.InsertValidator(validator.Address, validator.PublicKey, validator.Output, false, 1, DefaultStake, DefaultStake, DefaultPauseHeight, DefaultUnstakingHeight); err != nil {
		t.Fatal(err)
	}
	status, err := db.GetValidatorStatus(validator.Address)
	if err != nil {
		t.Fatal(err)
	}
	if status != DefaultStakeStatus {
		t.Fatalf("unexpected status: got %d expected %d", status, DefaultStakeStatus)
	}
}

func TestGetValidatorPauseHeightIfExists(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	validator := NewTestValidator()
	if err := db.InsertValidator(validator.Address, validator.PublicKey, validator.Output, false, 1, DefaultStake, DefaultStake, DefaultPauseHeight, DefaultUnstakingHeight); err != nil {
		t.Fatal(err)
	}
	height, err := db.GetValidatorPauseHeightIfExists(validator.Address)
	if err != nil {
		t.Fatal(err)
	}
	if height != DefaultPauseHeight {
		t.Fatalf("unexpected pauseHeight: got %d expected %d", DefaultPauseHeight, DefaultStakeStatus)
	}
}

func TestSetValidatorsStatusAndUnstakingHeightPausedBefore(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	validator := NewTestValidator()
	if err := db.InsertValidator(validator.Address, validator.PublicKey, validator.Output, false, 1, DefaultStake, DefaultStake, 0, DefaultUnstakingHeight); err != nil {
		t.Fatal(err)
	}
	if err := db.SetValidatorsStatusAndUnstakingHeightPausedBefore(1, 0, 1); err != nil {
		t.Fatal(err)
	}
	_, _, _, _, _, _, unstakingHeight, _, err := db.GetValidator(validator.Address)
	if err != nil {
		t.Fatal(err)
	}
	if unstakingHeight != 0 {
		t.Fatal("unexpected unstaking height")
	}
}

func TestSetValidatorPauseHeight(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	validator := NewTestValidator()
	if err := db.InsertValidator(validator.Address, validator.PublicKey, validator.Output, false, 1, DefaultStake, DefaultStake, 0, DefaultUnstakingHeight); err != nil {
		t.Fatal(err)
	}
	if err := db.SetValidatorPauseHeight(validator.Address, int64(PauseHeightToSet)); err != nil {
		t.Fatal(err)
	}
	_, _, _, _, _, pauseHeight, _, _, err := db.GetValidator(validator.Address)
	if err != nil {
		t.Fatal(err)
	}
	if pauseHeight != int64(PauseHeightToSet) {
		t.Fatal("unexpected pause height")
	}
}

func TestGetValidatorOutputAddress(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	validator := NewTestValidator()
	if err := db.InsertValidator(validator.Address, validator.PublicKey, validator.Output, false, 1, DefaultStake, DefaultStake, 0, DefaultUnstakingHeight); err != nil {
		t.Fatal(err)
	}
	output, err := db.GetValidatorOutputAddress(validator.Address)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(output, validator.Output) {
		t.Fatal("unexpected output address")
	}
}

func NewTestValidator() typesGenesis.Validator {
	pub1, _ := crypto.GeneratePublicKey()
	addr1 := pub1.Address()
	addr2, _ := crypto.GenerateAddress()
	return typesGenesis.Validator{
		Address:         addr1,
		PublicKey:       pub1.Bytes(),
		Paused:          false,
		Status:          typesGenesis.DefaultStakeStatus,
		ServiceUrl:      DefaultServiceUrl,
		StakedTokens:    typesGenesis.DefaultStake,
		PausedHeight:    0,
		UnstakingHeight: 0,
		Output:          addr2,
	}
}
