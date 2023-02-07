package test

import (
	"encoding/hex"
	"log"
	"testing"

	"github.com/pokt-network/pocket/persistence/types"

	"github.com/pokt-network/pocket/persistence"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/stretchr/testify/require"
)

func FuzzValidator(f *testing.F) {
	fuzzSingleProtocolActor(f,
		newTestGenericActor(types.ValidatorActor, newTestValidator),
		getGenericActor(types.ValidatorActor, getTestValidator),
		types.ValidatorActor)
}

func TestGetSetValidatorStakeAmount(t *testing.T) {
	db := NewTestPostgresContext(t, 1)
	getTestGetSetStakeAmountTest(t, db, createAndInsertDefaultTestValidator, db.GetValidatorStakeAmount, db.SetValidatorStakeAmount, 1)
}

func TestGetValidatorUpdatedAtHeight(t *testing.T) {
	getValidatorsUpdatedFunc := func(db *persistence.PostgresContext, height int64) ([]*coreTypes.Actor, error) {
		return db.GetActorsUpdated(types.ValidatorActor, height)
	}
	getAllActorsUpdatedAtHeightTest(t, createAndInsertDefaultTestValidator, getValidatorsUpdatedFunc, 5)
}

func TestInsertValidatorAndExists(t *testing.T) {
	db := NewTestPostgresContext(t, 0)

	validator, err := createAndInsertDefaultTestValidator(db)
	require.NoError(t, err)

	db.Height = 1

	validator2, err := createAndInsertDefaultTestValidator(db)
	require.NoError(t, err)

	addrBz, err := hex.DecodeString(validator.Address)
	require.NoError(t, err)

	addrBz2, err := hex.DecodeString(validator2.Address)
	require.NoError(t, err)

	exists, err := db.GetValidatorExists(addrBz, 0)
	require.NoError(t, err)
	require.True(t, exists, "actor that should exist at previous height does not")
	exists, err = db.GetValidatorExists(addrBz, 1)
	require.NoError(t, err)
	require.True(t, exists, "actor that should exist at current height does not")

	exists, err = db.GetValidatorExists(addrBz2, 0)
	require.NoError(t, err)
	require.False(t, exists, "actor that should not exist at previous height does")
	exists, err = db.GetValidatorExists(addrBz2, 1)
	require.NoError(t, err)
	require.True(t, exists, "actor that should exist at current height does not")
}

func TestUpdateValidator(t *testing.T) {
	db := NewTestPostgresContext(t, 0)

	validator, err := createAndInsertDefaultTestValidator(db)
	require.NoError(t, err)

	addrBz, err := hex.DecodeString(validator.Address)
	require.NoError(t, err)

	_, _, stakedTokens, _, _, _, _, err := db.GetValidator(addrBz, 0)
	require.NoError(t, err)
	require.Equal(t, DefaultStake, stakedTokens, "default stake incorrect for current height")

	db.Height = 1

	require.NotEqual(t, DefaultStake, StakeToUpdate) // sanity check to make sure the tests are correct
	err = db.UpdateValidator(addrBz, validator.GenericParam, StakeToUpdate)
	require.NoError(t, err)

	_, _, stakedTokens, _, _, _, _, err = db.GetValidator(addrBz, 0)
	require.NoError(t, err)
	require.Equal(t, DefaultStake, stakedTokens, "default stake incorrect for previous height")

	_, _, stakedTokens, _, _, _, _, err = db.GetValidator(addrBz, 1)
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

	addrBz, err := hex.DecodeString(validator.Address)
	require.NoError(t, err)

	addrBz2, err := hex.DecodeString(validator2.Address)
	require.NoError(t, err)

	addrBz3, err := hex.DecodeString(validator3.Address)
	require.NoError(t, err)

	// Unstake validator at height 0
	err = db.SetValidatorUnstakingHeightAndStatus(addrBz, 0, persistence.UnstakingStatus)
	require.NoError(t, err)

	// Unstake validator2 and validator3 at height 1
	err = db.SetValidatorUnstakingHeightAndStatus(addrBz2, 1, persistence.UnstakingStatus)
	require.NoError(t, err)
	err = db.SetValidatorUnstakingHeightAndStatus(addrBz3, 1, persistence.UnstakingStatus)
	require.NoError(t, err)

	// Check unstaking validators at height 0
	unstakingValidators, err := db.GetValidatorsReadyToUnstake(0, persistence.UnstakingStatus)
	require.NoError(t, err)
	require.Equal(t, 1, len(unstakingValidators), "wrong number of actors ready to unstake at height 0")
	// require.Equal(t, validator.Address, hex.EncodeToString(unstakingValidators[0].GetAddress()), "unexpected validatorlication actor returned")
	// addr, _ := hex.DecodeString(validator.Address)
	require.Equal(t, addrBz, unstakingValidators[0].GetAddress(), "unexpected unstaking validator returned")

	// Check unstaking validators at height 1
	unstakingValidators, err = db.GetValidatorsReadyToUnstake(1, persistence.UnstakingStatus)
	require.NoError(t, err)
	require.Equal(t, 2, len(unstakingValidators), "wrong number of actors ready to unstake at height 1")
	require.ElementsMatch(t, []string{validator2.Address, validator3.Address}, []string{unstakingValidators[0].Address, unstakingValidators[1].Address})
}

func TestGetValidatorStatus(t *testing.T) {
	db := NewTestPostgresContext(t, 1)

	validator, err := createAndInsertDefaultTestValidator(db)
	require.NoError(t, err)

	addrBz, err := hex.DecodeString(validator.Address)
	require.NoError(t, err)

	// Check status before the validator exists
	status, err := db.GetValidatorStatus(addrBz, 0)
	require.Error(t, err)
	require.Equal(t, persistence.UndefinedStakingStatus, status, "unexpected status")

	// Check status after the validator exists
	status, err = db.GetValidatorStatus(addrBz, 1)
	require.NoError(t, err)
	require.Equal(t, DefaultStakeStatus, status, "unexpected status")
}

func TestGetValidatorPauseHeightIfExists(t *testing.T) {
	db := NewTestPostgresContext(t, 1)

	validator, err := createAndInsertDefaultTestValidator(db)
	require.NoError(t, err)

	// TODO(andrew): In order to make the tests clearer (here and elsewhere), use `validatorAddrBz` instead of `addrBz`
	addrBz, err := hex.DecodeString(validator.Address)
	require.NoError(t, err)

	// Check pause height when validator does not exist
	pauseHeight, err := db.GetValidatorPauseHeightIfExists(addrBz, 0)
	require.Error(t, err)
	require.Equal(t, DefaultPauseHeight, pauseHeight, "unexpected pause height")

	// Check pause height when validator does not exist
	pauseHeight, err = db.GetValidatorPauseHeightIfExists(addrBz, 1)
	require.NoError(t, err)
	require.Equal(t, DefaultPauseHeight, pauseHeight, "unexpected pause height")
}

func TestSetValidatorPauseHeightAndUnstakeLater(t *testing.T) {
	db := NewTestPostgresContext(t, 0)

	validator, err := createAndInsertDefaultTestValidator(db)
	require.NoError(t, err)

	pauseHeight := int64(1)
	unstakingHeight := pauseHeight + 10

	addrBz, err := hex.DecodeString(validator.Address)
	require.NoError(t, err)

	err = db.SetValidatorPauseHeight(addrBz, pauseHeight)
	require.NoError(t, err)

	_, _, _, _, _, validatorPausedHeight, _, err := db.GetValidator(addrBz, db.Height)
	require.NoError(t, err)
	require.Equal(t, pauseHeight, validatorPausedHeight, "pause height not updated")

	err = db.SetValidatorsStatusAndUnstakingHeightIfPausedBefore(pauseHeight+1, unstakingHeight, -1 /*unused*/)
	require.NoError(t, err)

	_, _, _, _, _, _, validatorUnstakingHeight, err := db.GetValidator(addrBz, db.Height)
	require.NoError(t, err)
	require.Equal(t, unstakingHeight, validatorUnstakingHeight, "unstaking height was not set correctly")
}

func TestGetValidatorOutputAddress(t *testing.T) {
	db := NewTestPostgresContext(t, 0)

	validator, err := createAndInsertDefaultTestValidator(db)
	require.NoError(t, err)

	addrBz, err := hex.DecodeString(validator.Address)
	require.NoError(t, err)

	output, err := db.GetValidatorOutputAddress(addrBz, 0)
	require.NoError(t, err)
	require.Equal(t, validator.Output, hex.EncodeToString(output), "unexpected output address")
}

func newTestValidator() (*coreTypes.Actor, error) {
	operatorKey, err := crypto.GeneratePublicKey()
	if err != nil {
		return nil, err
	}

	outputAddr, err := crypto.GenerateAddress()
	if err != nil {
		return nil, err
	}

	return &coreTypes.Actor{
		Address:         hex.EncodeToString(operatorKey.Address()),
		PublicKey:       hex.EncodeToString(operatorKey.Bytes()),
		GenericParam:    DefaultServiceUrl,
		StakedAmount:    DefaultStake,
		PausedHeight:    DefaultPauseHeight,
		UnstakingHeight: DefaultUnstakingHeight,
		Output:          hex.EncodeToString(outputAddr),
	}, nil
}

func createAndInsertDefaultTestValidator(db *persistence.PostgresContext) (*coreTypes.Actor, error) {
	validator, err := newTestValidator()
	if err != nil {
		return nil, err
	}
	addrBz, err := hex.DecodeString(validator.Address)
	if err != nil {
		log.Fatalf("an error occurred converting address to bytes %s", validator.Address)
	}
	pubKeyBz, err := hex.DecodeString(validator.PublicKey)
	if err != nil {
		log.Fatalf("an error occurred converting pubKey to bytes %s", validator.PublicKey)
	}
	outputBz, err := hex.DecodeString(validator.Output)
	if err != nil {
		log.Fatalf("an error occurred converting output to bytes %s", validator.Output)
	}
	return validator, db.InsertValidator(
		addrBz,
		pubKeyBz,
		outputBz,
		false,
		DefaultStakeStatus,
		DefaultServiceUrl,
		DefaultStake,
		DefaultPauseHeight,
		DefaultUnstakingHeight)
}

func getTestValidator(db *persistence.PostgresContext, address []byte) (*coreTypes.Actor, error) {
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

	return &coreTypes.Actor{
		Address:         hex.EncodeToString(operatorAddr),
		PublicKey:       hex.EncodeToString(operatorPubKey),
		GenericParam:    serviceURL,
		StakedAmount:    stakedTokens,
		PausedHeight:    pauseHeight,
		UnstakingHeight: unstakingHeight,
		Output:          hex.EncodeToString(outputAddr),
	}, nil
}
