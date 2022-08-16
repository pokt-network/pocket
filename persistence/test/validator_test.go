package test

import (
	"encoding/hex"
	"testing"

	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/persistence/schema"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types/genesis"
	typesGenesis "github.com/pokt-network/pocket/shared/types/genesis"
	"github.com/stretchr/testify/require"
)

func FuzzValidator(f *testing.F) {
	fuzzSingleProtocolActor(f,
		NewTestGenericActor(schema.ValidatorActor, newTestValidator),
		GetGenericActor(schema.ValidatorActor, getTestValidator),
		schema.ValidatorActor)
}

func TestInsertValidatorAndExists(t *testing.T) {
	db := NewTestPostgresContext(t, 0)

	validator, err := createAndInsertDefaultTestValidator(db)
	require.NoError(t, err)

	db.Height = 1

	validator2, err := createAndInsertDefaultTestValidator(db)
	require.NoError(t, err)

	exists, err := db.GetValidatorExists(validator.Address, 0)
	require.NoError(t, err)
	require.True(t, exists, "actor that should exist at previous height does not")

	exists, err = db.GetValidatorExists(validator.Address, 1)
	require.NoError(t, err)
	require.True(t, exists, "actor that should exist at current height does not")

	exists, err = db.GetValidatorExists(validator2.Address, 0)
	require.NoError(t, err)
	require.False(t, exists, "actor that should not exist at previous height appears to")

	exists, err = db.GetValidatorExists(validator2.Address, 1)
	require.NoError(t, err)
	require.True(t, exists, "actor that should exist at current height does not")
}

func TestUpdateValidator(t *testing.T) {
	db := NewTestPostgresContext(t, 0)

	validator, err := createAndInsertDefaultTestValidator(db)
	require.NoError(t, err)

	_, _, stakedTokens, _, _, _, _, err := db.GetValidator(validator.Address, 0)
	require.NoError(t, err)
	require.Equal(t, DefaultStake, stakedTokens, "default stake incorrect for current height")

	db.Height = 1

	require.NotEqual(t, DefaultStake, StakeToUpdate) // sanity check to make sure the tests are correct
	err = db.UpdateValidator(validator.Address, validator.ServiceUrl, StakeToUpdate)
	require.NoError(t, err)

	_, _, stakedTokens, _, _, _, _, err = db.GetValidator(validator.Address, 0)
	require.NoError(t, err)
	require.Equal(t, DefaultStake, stakedTokens, "default stake incorrect for previous height")

	_, _, stakedTokens, _, _, _, _, err = db.GetValidator(validator.Address, 1)
	require.NoError(t, err)
	require.Equal(t, StakeToUpdate, stakedTokens, "stake not updated for current height")
}

func TestGetValidatorsReadyToUnstake(t *testing.T) {
	db := NewTestPostgresContext(t, 0)

	validator, err := createAndInsertDefaultTestValidator(db)
	require.NoError(t, err)

	validator2, err := createAndInsertDefaultTestValidator(db)
	require.NoError(t, err)

	validator3, err := createAndInsertDefaultTestValidator(db)
	require.NoError(t, err)

	// Unstake validator at height 0
	err = db.SetValidatorUnstakingHeightAndStatus(validator.Address, 0, persistence.UnstakingStatus)
	require.NoError(t, err)

	// Unstake validator2 and validator3 at height 1
	err = db.SetValidatorUnstakingHeightAndStatus(validator2.Address, 1, persistence.UnstakingStatus)
	require.NoError(t, err)
	err = db.SetValidatorUnstakingHeightAndStatus(validator3.Address, 1, persistence.UnstakingStatus)
	require.NoError(t, err)

	// Check unstaking validators at height 0
	unstakingValidators, err := db.GetValidatorsReadyToUnstake(0, persistence.UnstakingStatus)
	require.NoError(t, err)
	require.Equal(t, 1, len(unstakingValidators), "wrong number of actors ready to unstake at height 0")
	require.Equal(t, validator.Address, unstakingValidators[0].Address, "unexpected validatorlication actor returned")

	// Check unstaking validators at height 1
	unstakingValidators, err = db.GetValidatorsReadyToUnstake(1, persistence.UnstakingStatus)
	require.NoError(t, err)
	require.Equal(t, 2, len(unstakingValidators), "wrong number of actors ready to unstake at height 1")
	require.ElementsMatch(t, [][]byte{validator2.Address, validator3.Address}, [][]byte{unstakingValidators[0].Address, unstakingValidators[1].Address})
}

func TestGetValidatorStatus(t *testing.T) {
	db := NewTestPostgresContext(t, 1)

	validator, err := createAndInsertDefaultTestValidator(db)
	require.NoError(t, err)

	// Check status before the validator exists
	status, err := db.GetValidatorStatus(validator.Address, 0)
	require.Error(t, err)
	require.Equal(t, status, persistence.UndefinedStakingStatus, "unexpected status")

	// Check status after the validator exists
	status, err = db.GetValidatorStatus(validator.Address, 1)
	require.NoError(t, err)
	require.Equal(t, status, DefaultStakeStatus, "unexpected status")
}

func TestGetValidatorPauseHeightIfExists(t *testing.T) {
	db := NewTestPostgresContext(t, 1)

	validator, err := createAndInsertDefaultTestValidator(db)
	require.NoError(t, err)

	// Check pause height when validator does not exist
	pauseHeight, err := db.GetValidatorPauseHeightIfExists(validator.Address, 0)
	require.Error(t, err)
	require.Equal(t, pauseHeight, DefaultPauseHeight, "unexpected pause height")

	// Check pause height when validator does not exist
	pauseHeight, err = db.GetValidatorPauseHeightIfExists(validator.Address, 1)
	require.NoError(t, err)
	require.Equal(t, pauseHeight, DefaultPauseHeight, "unexpected pause height")
}

func TestSetValidatorPauseHeightAndUnstakeLater(t *testing.T) {
	db := NewTestPostgresContext(t, 0)

	validator, err := createAndInsertDefaultTestValidator(db)
	require.NoError(t, err)

	pauseHeight := int64(1)
	unstakingHeight := pauseHeight + 10

	err = db.SetValidatorPauseHeight(validator.Address, pauseHeight)
	require.NoError(t, err)

	_, _, _, _, _, validatorPausedHeight, _, err := db.GetValidator(validator.Address, db.Height)
	require.NoError(t, err)
	require.Equal(t, pauseHeight, validatorPausedHeight, "pause height not updated")

	err = db.SetValidatorsStatusAndUnstakingHeightIfPausedBefore(pauseHeight+1, unstakingHeight, -1 /*unused*/)
	require.NoError(t, err)

	_, _, _, _, _, _, validatorUnstakingHeight, err := db.GetValidator(validator.Address, db.Height)
	require.NoError(t, err)
	require.Equal(t, unstakingHeight, validatorUnstakingHeight, "unstaking height was not set correctly")
}

func TestGetValidatorOutputAddress(t *testing.T) {
	db := NewTestPostgresContext(t, 0)

	validator, err := createAndInsertDefaultTestValidator(db)
	require.NoError(t, err)

	output, err := db.GetValidatorOutputAddress(validator.Address, 0)
	require.NoError(t, err)
	require.Equal(t, output, validator.Output, "unexpected output address")
}

func TestGetAllValidators(t *testing.T) {
	db := NewTestPostgresContext(t, 0)

	updateValidator := func(db *persistence.PostgresContext, val *genesis.Validator) error {
		return db.UpdateValidator(val.Address, OlshanskyURL, val.StakedTokens)
	}

	getAllActorsTest(t, db, db.GetAllValidators, createAndInsertDefaultTestValidator, updateValidator, 5)
}

func newTestValidator() (*typesGenesis.Validator, error) {
	operatorKey, err := crypto.GeneratePublicKey()
	if err != nil {
		return nil, err
	}

	outputAddr, err := crypto.GenerateAddress()
	if err != nil {
		return nil, err
	}

	return &typesGenesis.Validator{
		Address:         operatorKey.Address(),
		PublicKey:       operatorKey.Bytes(),
		Paused:          false,
		Status:          typesGenesis.DefaultStakeStatus,
		ServiceUrl:      DefaultServiceUrl,
		StakedTokens:    typesGenesis.DefaultStake,
		PausedHeight:    DefaultPauseHeight,
		UnstakingHeight: DefaultUnstakingHeight,
		Output:          outputAddr,
	}, nil
}

func createAndInsertDefaultTestValidator(db *persistence.PostgresContext) (*typesGenesis.Validator, error) {
	validator, err := newTestValidator()
	if err != nil {
		return nil, err
	}

	return validator, db.InsertValidator(
		validator.Address,
		validator.PublicKey,
		validator.Output,
		false,
		DefaultStakeStatus,
		DefaultServiceUrl,
		DefaultStake,
		DefaultPauseHeight,
		DefaultUnstakingHeight)
}

func getTestValidator(db *persistence.PostgresContext, address []byte) (*typesGenesis.Validator, error) {
	operator, publicKey, stakedTokens, serviceURL, outputAddress, pauseHeight, unstakingHeight, err := db.GetValidator(address, db.Height)
	if err != nil {
		return nil, err
	}

	operatorAddr, err := hex.DecodeString(operator)
	if err != nil {
		return nil, err
	}

	operatorPubKey, err := hex.DecodeString(publicKey)
	if err != nil {
		return nil, err
	}

	outputAddr, err := hex.DecodeString(outputAddress)
	if err != nil {
		return nil, err
	}

	return &typesGenesis.Validator{
		Address:         operatorAddr,
		PublicKey:       operatorPubKey,
		Paused:          false,
		Status:          persistence.UnstakingHeightToStatus(unstakingHeight),
		ServiceUrl:      serviceURL,
		StakedTokens:    stakedTokens,
		PausedHeight:    pauseHeight,
		UnstakingHeight: unstakingHeight,
		Output:          outputAddr,
	}, nil
}
