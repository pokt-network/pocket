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

func FuzzServicer(f *testing.F) {
	fuzzSingleProtocolActor(f,
		newTestGenericActor(types.ServicerActor, newTestServicer),
		getGenericActor(types.ServicerActor, getTestServicer),
		types.ServicerActor)
}

func TestGetSetServicerStakeAmount(t *testing.T) {
	db := NewTestPostgresContext(t, 1)
	getTestGetSetStakeAmountTest(t, db, createAndInsertDefaultTestServicer, db.GetServicerStakeAmount, db.SetServicerStakeAmount, 1)
}

func TestGetServicerUpdatedAtHeight(t *testing.T) {
	getServicerUpdatedFunc := func(db *persistence.PostgresContext, height int64) ([]*coreTypes.Actor, error) {
		return db.GetActorsUpdated(types.ServicerActor, height)
	}
	getAllActorsUpdatedAtHeightTest(t, createAndInsertDefaultTestServicer, getServicerUpdatedFunc, 1)
}

func TestInsertServicerAndExists(t *testing.T) {
	db := NewTestPostgresContext(t, 0)

	servicer, err := createAndInsertDefaultTestServicer(db)
	require.NoError(t, err)

	db.Height = 1

	servicer2, err := createAndInsertDefaultTestServicer(db)
	require.NoError(t, err)

	addrBz, err := hex.DecodeString(servicer.Address)
	require.NoError(t, err)
	addrBz2, err := hex.DecodeString(servicer2.Address)
	require.NoError(t, err)

	exists, err := db.GetServicerExists(addrBz, 0)
	require.NoError(t, err)
	require.True(t, exists, "actor that should exist at previous height does not")
	exists, err = db.GetServicerExists(addrBz, 1)
	require.NoError(t, err)
	require.True(t, exists, "actor that should exist at current height does not")

	exists, err = db.GetServicerExists(addrBz2, 0)
	require.NoError(t, err)
	require.False(t, exists, "actor that should not exist at previous height servicerears to")
	exists, err = db.GetServicerExists(addrBz2, 1)
	require.NoError(t, err)
	require.True(t, exists, "actor that should exist at current height does not")
}

func TestUpdateServicer(t *testing.T) {
	db := NewTestPostgresContext(t, 0)

	servicer, err := createAndInsertDefaultTestServicer(db)
	require.NoError(t, err)

	addrBz, err := hex.DecodeString(servicer.Address)
	require.NoError(t, err)

	_, _, stakedTokens, _, _, _, _, chains, err := db.GetServicer(addrBz, 0)
	require.NoError(t, err)
	require.Equal(t, DefaultChains, chains, "default chains incorrect for current height")
	require.Equal(t, DefaultStake, stakedTokens, "default stake incorrect for current height")

	db.Height = 1

	require.NotEqual(t, DefaultStake, StakeToUpdate)   // sanity check to make sure the tests are correct
	require.NotEqual(t, DefaultChains, ChainsToUpdate) // sanity check to make sure the tests are correct
	err = db.UpdateServicer(addrBz, servicer.GenericParam, StakeToUpdate, ChainsToUpdate)
	require.NoError(t, err)

	_, _, stakedTokens, _, _, _, _, chains, err = db.GetServicer(addrBz, 0)
	require.NoError(t, err)
	require.Equal(t, DefaultChains, chains, "default chains incorrect for previous height")
	require.Equal(t, DefaultStake, stakedTokens, "default stake incorrect for previous height")

	_, _, stakedTokens, _, _, _, _, chains, err = db.GetServicer(addrBz, 1)
	require.NoError(t, err)
	require.Equal(t, ChainsToUpdate, chains, "chains not updated for current height")
	require.Equal(t, StakeToUpdate, stakedTokens, "stake not updated for current height")
}

func TestGetServicersReadyToUnstake(t *testing.T) {
	db := NewTestPostgresContext(t, 0)

	servicer, err := createAndInsertDefaultTestServicer(db)
	require.NoError(t, err)

	servicer2, err := createAndInsertDefaultTestServicer(db)
	require.NoError(t, err)

	servicer3, err := createAndInsertDefaultTestServicer(db)
	require.NoError(t, err)

	addrBz, err := hex.DecodeString(servicer.Address)
	require.NoError(t, err)

	addrBz2, err := hex.DecodeString(servicer2.Address)
	require.NoError(t, err)

	addrBz3, err := hex.DecodeString(servicer3.Address)
	require.NoError(t, err)

	// Unstake servicer at height 0
	err = db.SetServicerUnstakingHeightAndStatus(addrBz, 0, persistence.UnstakingStatus)
	require.NoError(t, err)

	// Unstake servicer2 and servicer3 at height 1
	err = db.SetServicerUnstakingHeightAndStatus(addrBz2, 1, persistence.UnstakingStatus)
	require.NoError(t, err)
	err = db.SetServicerUnstakingHeightAndStatus(addrBz3, 1, persistence.UnstakingStatus)
	require.NoError(t, err)

	// Check unstaking servicers at height 0
	unstakingServicers, err := db.GetServicersReadyToUnstake(0, persistence.UnstakingStatus)
	require.NoError(t, err)
	require.Equal(t, 1, len(unstakingServicers), "wrong number of actors ready to unstake at height 0")
	require.Equal(t, servicer.Address, unstakingServicers[0].Address, "unexpected servicerlication actor returned")

	// Check unstaking servicers at height 1
	unstakingServicers, err = db.GetServicersReadyToUnstake(1, persistence.UnstakingStatus)
	require.NoError(t, err)
	require.Equal(t, 2, len(unstakingServicers), "wrong number of actors ready to unstake at height 1")
	require.ElementsMatch(t, []string{servicer2.Address, servicer3.Address}, []string{unstakingServicers[0].Address, unstakingServicers[1].Address})
}

func TestGetServicerStatus(t *testing.T) {
	db := NewTestPostgresContext(t, 1)

	servicer, err := createAndInsertDefaultTestServicer(db)
	require.NoError(t, err)

	addrBz, err := hex.DecodeString(servicer.Address)
	require.NoError(t, err)

	// Check status before the servicer exists
	status, err := db.GetServicerStatus(addrBz, 0)
	require.Error(t, err)
	require.Equal(t, persistence.UndefinedStakingStatus, status, "unexpected status")

	// Check status after the servicer exists
	status, err = db.GetServicerStatus(addrBz, 1)
	require.NoError(t, err)
	require.Equal(t, DefaultStakeStatus, status, "unexpected status")
}

func TestGetServicerPauseHeightIfExists(t *testing.T) {
	db := NewTestPostgresContext(t, 1)

	servicer, err := createAndInsertDefaultTestServicer(db)
	require.NoError(t, err)

	addrBz, err := hex.DecodeString(servicer.Address)
	require.NoError(t, err)

	// Check pause height when servicer does not exist
	pauseHeight, err := db.GetServicerPauseHeightIfExists(addrBz, 0)
	require.Error(t, err)
	require.Equal(t, DefaultPauseHeight, pauseHeight, "unexpected pause height")

	// Check pause height when servicer does not exist
	pauseHeight, err = db.GetServicerPauseHeightIfExists(addrBz, 1)
	require.NoError(t, err)
	require.Equal(t, DefaultPauseHeight, pauseHeight, "unexpected pause height")
}

func TestSetServicerPauseHeightAndUnstakeLater(t *testing.T) {
	db := NewTestPostgresContext(t, 0)

	servicer, err := createAndInsertDefaultTestServicer(db)
	require.NoError(t, err)

	pauseHeight := int64(1)
	unstakingHeight := pauseHeight + 10

	addrBz, err := hex.DecodeString(servicer.Address)
	require.NoError(t, err)

	err = db.SetServicerPauseHeight(addrBz, pauseHeight)
	require.NoError(t, err)

	_, _, _, _, _, servicerPausedHeight, _, _, err := db.GetServicer(addrBz, db.Height)
	require.NoError(t, err)
	require.Equal(t, pauseHeight, servicerPausedHeight, "pause height not updated")

	err = db.SetServicerStatusAndUnstakingHeightIfPausedBefore(pauseHeight+1, unstakingHeight, -1 /*unused*/)
	require.NoError(t, err)

	_, _, _, _, _, _, servicerUnstakingHeight, _, err := db.GetServicer(addrBz, db.Height)
	require.NoError(t, err)
	require.Equal(t, unstakingHeight, servicerUnstakingHeight, "unstaking height was not set correctly")
}

func TestGetServicerOutputAddress(t *testing.T) {
	db := NewTestPostgresContext(t, 0)

	servicer, err := createAndInsertDefaultTestServicer(db)
	require.NoError(t, err)

	addrBz, err := hex.DecodeString(servicer.Address)
	require.NoError(t, err)

	output, err := db.GetServicerOutputAddress(addrBz, 0)
	require.NoError(t, err)
	require.Equal(t, servicer.Output, hex.EncodeToString(output), "unexpected output address")
}

func newTestServicer() (*coreTypes.Actor, error) {
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
		Chains:          DefaultChains,
		GenericParam:    DefaultServiceUrl,
		StakedAmount:    DefaultStake,
		PausedHeight:    DefaultPauseHeight,
		UnstakingHeight: DefaultUnstakingHeight,
		Output:          hex.EncodeToString(outputAddr),
	}, nil
}

func createAndInsertDefaultTestServicer(db *persistence.PostgresContext) (*coreTypes.Actor, error) {
	servicer, err := newTestServicer()
	if err != nil {
		return nil, err
	}
	addrBz, err := hex.DecodeString(servicer.Address)
	if err != nil {
		log.Fatalf("an error occurred converting address to bytes %s", servicer.Address)
	}
	pubKeyBz, err := hex.DecodeString(servicer.PublicKey)
	if err != nil {
		log.Fatalf("an error occurred converting pubKey to bytes %s", servicer.PublicKey)
	}
	outputBz, err := hex.DecodeString(servicer.Output)
	if err != nil {
		log.Fatalf("an error occurred converting output to bytes %s", servicer.Output)
	}
	return servicer, db.InsertServicer(
		addrBz,
		pubKeyBz,
		outputBz,
		false,
		DefaultStakeStatus,
		DefaultServiceUrl,
		DefaultStake,
		DefaultChains,
		DefaultPauseHeight,
		DefaultUnstakingHeight)
}

func getTestServicer(db *persistence.PostgresContext, address []byte) (*coreTypes.Actor, error) {
	operator, publicKey, stakedTokens, serviceURL, outputAddress, pauseHeight, unstakingHeight, chains, err := db.GetServicer(address, db.Height)
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
		Chains:          chains,
		GenericParam:    serviceURL,
		StakedAmount:    stakedTokens,
		PausedHeight:    pauseHeight,
		UnstakingHeight: unstakingHeight,
		Output:          hex.EncodeToString(outputAddr),
	}, nil
}
