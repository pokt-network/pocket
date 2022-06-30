package test

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/persistence/schema"
	"github.com/pokt-network/pocket/shared/crypto"
	typesGenesis "github.com/pokt-network/pocket/shared/types/genesis"
	"github.com/stretchr/testify/require"
)

func FuzzValidator(f *testing.F) {
	fuzzProtocolActor(f,
		NewTestGenericActor(schema.ValidatorActor, newTestValidator),
		GetGenericActor(schema.ValidatorActor, GetTestValidator),
		schema.ValidatorActor)
}

func TestInsertValidatorAndExists(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	validator := NewTestValidator(t)
	validator2 := NewTestValidator(t)
	err := db.InsertValidator(validator.Address, validator.PublicKey, validator.Output, false, DefaultStakeStatus, DefaultStake, DefaultStake, DefaultPauseHeight, DefaultUnstakingHeight)
	require.NoError(t, err)
	exists, err := db.GetValidatorExists(validator.Address, db.Height)
	require.NoError(t, err)
	if !exists {
		t.Fatal("actor that should exist does not")
	}
	exists, err = db.GetValidatorExists(validator2.Address, db.Height)
	require.NoError(t, err)
	if exists {
		t.Fatal("actor that should not exist, appears to")
	}
}

func TestUpdateValidator(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	validator := NewTestValidator(t)
	err := db.InsertValidator(validator.Address, validator.PublicKey, validator.Output, false, DefaultStakeStatus, DefaultStake, DefaultStake, DefaultPauseHeight, DefaultUnstakingHeight)
	require.NoError(t, err)
	_, _, stakedTokens, _, _, _, _, err := db.GetValidator(validator.Address, db.Height)
	require.NoError(t, err)
	err = db.UpdateValidator(validator.Address, validator.ServiceUrl, StakeToUpdate)
	require.NoError(t, err)
	_, _, stakedTokens, _, _, _, _, err = db.GetValidator(validator.Address, db.Height)
	require.NoError(t, err)
	if stakedTokens != StakeToUpdate {
		t.Fatal("stake not updated")
	}
}

func TestGetValidatorsReadyToUnstake(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	db.ClearAllDebug()
	validator := NewTestValidator(t)
	err := db.InsertValidator(validator.Address, validator.PublicKey, validator.Output, false, 1, DefaultStake, DefaultStake, DefaultPauseHeight, DefaultUnstakingHeight)
	require.NoError(t, err)
	// test SetValidatorUnstakingHeightAndStatus
	err = db.SetValidatorUnstakingHeightAndStatus(validator.Address, 0, 1)
	require.NoError(t, err)
	// test GetValidatorsReadyToUnstake
	validators, err := db.GetValidatorsReadyToUnstake(0, 1)
	require.NoError(t, err)
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
	validator := NewTestValidator(t)
	err := db.InsertValidator(validator.Address, validator.PublicKey, validator.Output, false, 1, DefaultStake, DefaultStake, DefaultPauseHeight, DefaultUnstakingHeight)
	require.NoError(t, err)
	status, err := db.GetValidatorStatus(validator.Address, db.Height)
	require.NoError(t, err)
	if status != DefaultStakeStatus {
		t.Fatalf("unexpected status: got %d expected %d", status, DefaultStakeStatus)
	}
}

func TestGetValidatorPauseHeightIfExists(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	validator := NewTestValidator(t)
	err := db.InsertValidator(validator.Address, validator.PublicKey, validator.Output, false, 1, DefaultStake, DefaultStake, DefaultPauseHeight, DefaultUnstakingHeight)
	require.NoError(t, err)
	height, err := db.GetValidatorPauseHeightIfExists(validator.Address, db.Height)
	require.NoError(t, err)
	if height != DefaultPauseHeight {
		t.Fatalf("unexpected pausedHeight: got %d expected %d", DefaultPauseHeight, DefaultStakeStatus)
	}
}

func TestSetValidatorsStatusAndUnstakingHeightPausedBefore(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	validator := NewTestValidator(t)
	err := db.InsertValidator(validator.Address, validator.PublicKey, validator.Output, false, 1, DefaultStake, DefaultStake, 0, DefaultUnstakingHeight)
	require.NoError(t, err)
	err = db.SetValidatorsStatusAndUnstakingHeightPausedBefore(1, 0, 1)
	require.NoError(t, err)
	_, _, _, _, _, unstakingHeight, _, err := db.GetValidator(validator.Address, db.Height)
	require.NoError(t, err)
	if unstakingHeight != 0 {
		t.Fatal("unexpected unstaking height")
	}
}

func TestSetValidatorPauseHeight(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	validator := NewTestValidator(t)
	err := db.InsertValidator(validator.Address, validator.PublicKey, validator.Output, false, 1, DefaultStake, DefaultStake, 0, DefaultUnstakingHeight)
	require.NoError(t, err)
	err = db.SetValidatorPauseHeight(validator.Address, int64(PauseHeightToSet))
	require.NoError(t, err)
	_, _, _, _, _, pausedHeight, _, err := db.GetValidator(validator.Address, db.Height)
	require.NoError(t, err)
	if pausedHeight != int64(PauseHeightToSet) {
		t.Fatal("unexpected pause height")
	}
}

func TestGetValidatorOutputAddress(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	validator := NewTestValidator(t)
	err := db.InsertValidator(validator.Address, validator.PublicKey, validator.Output, false, 1, DefaultStake, DefaultStake, 0, DefaultUnstakingHeight)
	require.NoError(t, err)
	output, err := db.GetValidatorOutputAddress(validator.Address, db.Height)
	require.NoError(t, err)
	if !bytes.Equal(output, validator.Output) {
		t.Fatal("unexpected output address")
	}
}

func NewTestValidator(t *testing.T) typesGenesis.Validator {
	v, err := newTestValidator()
	require.NoError(t, err)
	return v
}

func newTestValidator() (typesGenesis.Validator, error) {
	pub1, err := crypto.GeneratePublicKey()
	if err != nil {
		return typesGenesis.Validator{}, nil
	}
	addr1 := pub1.Address()
	addr2, err := crypto.GenerateAddress()
	if err != nil {
		return typesGenesis.Validator{}, nil
	}
	return typesGenesis.Validator{
		Address:         addr1,
		PublicKey:       pub1.Bytes(),
		Paused:          false,
		Status:          typesGenesis.DefaultStakeStatus,
		ServiceUrl:      DefaultServiceUrl,
		StakedTokens:    typesGenesis.DefaultStake,
		PausedHeight:    uint64(DefaultPauseHeight),
		UnstakingHeight: DefaultUnstakingHeight,
		Output:          addr2,
	}, nil
}

func GetTestValidator(db persistence.PostgresContext, address []byte) (*typesGenesis.Validator, error) {
	operator, publicKey, stakedTokens, serviceURL, outputAddress, pauseHeight, unstakingHeight, err := db.GetValidator(address, db.Height)
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
	return &typesGenesis.Validator{
		Address:         addr,
		PublicKey:       pubKey,
		Paused:          false,
		Status:          int32(status),
		ServiceUrl:      serviceURL,
		StakedTokens:    stakedTokens,
		PausedHeight:    uint64(pauseHeight),
		UnstakingHeight: unstakingHeight,
		Output:          outputAddr,
	}, nil
}
