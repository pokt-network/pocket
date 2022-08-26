package test

import (
	"encoding/hex"
	"fmt"
	"github.com/pokt-network/pocket/persistence/types"
	"log"
	"testing"

	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/stretchr/testify/require"
)

// TODO(andrew): Rename `addrBz` to `fishAddrBz` so tests are easier to read and understand. Ditto in all other locations.

func FuzzFisherman(f *testing.F) {
	fuzzSingleProtocolActor(f,
		NewTestGenericActor(types.FishermanActor, newTestFisherman),
		GetGenericActor(types.FishermanActor, getTestFisherman),
		types.FishermanActor)
}

func TestGetSetFishermanStakeAmount(t *testing.T) {
	db := NewTestPostgresContext(t, 1)
	getTestGetSetStakeAmountTest(t, db, createAndInsertDefaultTestFisherman, db.GetFishermanStakeAmount, db.SetFishermanStakeAmount, 1)
}

func TestInsertFishermanAndExists(t *testing.T) {
	db := NewTestPostgresContext(t, 0)

	fisherman, err := createAndInsertDefaultTestFisherman(db)
	require.NoError(t, err)

	db.Height = 1

	fisherman2, err := createAndInsertDefaultTestFisherman(db)
	require.NoError(t, err)

	addrBz, err := hex.DecodeString(fisherman.Address)
	require.NoError(t, err)
	addrBz2, err := hex.DecodeString(fisherman2.Address)
	require.NoError(t, err)

	exists, err := db.GetFishermanExists(addrBz, 0)
	require.NoError(t, err)
	require.True(t, exists, "actor that should exist at previous height does not")
	exists, err = db.GetFishermanExists(addrBz, 1)
	require.NoError(t, err)
	require.True(t, exists, "actor that should exist at current height does not")

	exists, err = db.GetFishermanExists(addrBz2, 0)
	require.NoError(t, err)
	require.False(t, exists, "actor that should not exist at previous height fishermanears to")
	exists, err = db.GetFishermanExists(addrBz2, 1)
	require.NoError(t, err)
	require.True(t, exists, "actor that should exist at current height does not")
}

func TestUpdateFisherman(t *testing.T) {
	db := NewTestPostgresContext(t, 0)

	fisherman, err := createAndInsertDefaultTestFisherman(db)
	require.NoError(t, err)

	addrBz, err := hex.DecodeString(fisherman.Address)
	require.NoError(t, err)

	_, _, stakedTokens, _, _, _, _, chains, err := db.GetFisherman(addrBz, 0)
	require.NoError(t, err)
	require.Equal(t, DefaultChains, chains, "default chains incorrect for current height")
	require.Equal(t, DefaultStake, stakedTokens, "default stake incorrect for current height")

	db.Height = 1

	require.NotEqual(t, DefaultStake, StakeToUpdate)   // sanity check to make sure the tests are correct
	require.NotEqual(t, DefaultChains, ChainsToUpdate) // sanity check to make sure the tests are correct
	err = db.UpdateFisherman(addrBz, fisherman.GenericParam, StakeToUpdate, ChainsToUpdate)
	require.NoError(t, err)

	_, _, stakedTokens, _, _, _, _, chains, err = db.GetFisherman(addrBz, 0)
	require.NoError(t, err)
	require.Equal(t, DefaultChains, chains, "default chains incorrect for previous height")
	require.Equal(t, DefaultStake, stakedTokens, "default stake incorrect for previous height")

	_, _, stakedTokens, _, _, _, _, chains, err = db.GetFisherman(addrBz, 1)
	require.NoError(t, err)
	require.Equal(t, ChainsToUpdate, chains, "chains not updated for current height")
	require.Equal(t, StakeToUpdate, stakedTokens, "stake not updated for current height")
}

func TestGetFishermenReadyToUnstake(t *testing.T) {
	db := NewTestPostgresContext(t, 0)

	fisherman, err := createAndInsertDefaultTestFisherman(db)
	require.NoError(t, err)

	fisherman2, err := createAndInsertDefaultTestFisherman(db)
	require.NoError(t, err)

	fisherman3, err := createAndInsertDefaultTestFisherman(db)
	require.NoError(t, err)

	addrBz, err := hex.DecodeString(fisherman.Address)
	require.NoError(t, err)
	addrBz2, err := hex.DecodeString(fisherman2.Address)
	require.NoError(t, err)
	addrBz3, err := hex.DecodeString(fisherman3.Address)
	require.NoError(t, err)

	// Unstake fisherman at height 0
	err = db.SetFishermanUnstakingHeightAndStatus(addrBz, 0, persistence.UnstakingStatus)
	require.NoError(t, err)

	// Unstake fisherman2 and fisherman3 at height 1
	err = db.SetFishermanUnstakingHeightAndStatus(addrBz2, 1, persistence.UnstakingStatus)
	require.NoError(t, err)
	err = db.SetFishermanUnstakingHeightAndStatus(addrBz3, 1, persistence.UnstakingStatus)
	require.NoError(t, err)

	// Check unstaking fishermans at height 0
	unstakingFishermen, err := db.GetFishermenReadyToUnstake(0, persistence.UnstakingStatus)
	require.NoError(t, err)
	require.Equal(t, 1, len(unstakingFishermen), "wrong number of actors ready to unstake at height 0")
	require.Equal(t, fisherman.Address, hex.EncodeToString(unstakingFishermen[0].GetAddress()), "unexpected fishermanlication actor returned")

	// Check unstaking fishermans at height 1
	unstakingFishermen, err = db.GetFishermenReadyToUnstake(1, persistence.UnstakingStatus)
	require.NoError(t, err)
	require.Equal(t, 2, len(unstakingFishermen), "wrong number of actors ready to unstake at height 1")
	require.ElementsMatch(t, [][]byte{addrBz2, addrBz3}, [][]byte{unstakingFishermen[0].GetAddress(), unstakingFishermen[1].GetAddress()})
}

func TestGetFishermanStatus(t *testing.T) {
	db := NewTestPostgresContext(t, 1)

	fisherman, err := createAndInsertDefaultTestFisherman(db)
	require.NoError(t, err)

	addrBz, err := hex.DecodeString(fisherman.Address)
	require.NoError(t, err)

	// Check status before the fisherman exists
	status, err := db.GetFishermanStatus(addrBz, 0)
	require.Error(t, err)
	require.Equal(t, status, persistence.UndefinedStakingStatus, "unexpected status")

	// Check status after the fisherman exists
	status, err = db.GetFishermanStatus(addrBz, 1)
	require.NoError(t, err)
	require.Equal(t, status, DefaultStakeStatus, "unexpected status")
}

func TestGetFishermanPauseHeightIfExists(t *testing.T) {
	db := NewTestPostgresContext(t, 1)

	fisherman, err := createAndInsertDefaultTestFisherman(db)
	require.NoError(t, err)

	addrBz, err := hex.DecodeString(fisherman.Address)
	require.NoError(t, err)

	// Check pause height when fisherman does not exist
	pauseHeight, err := db.GetFishermanPauseHeightIfExists(addrBz, 0)
	require.Error(t, err)
	require.Equal(t, pauseHeight, DefaultPauseHeight, "unexpected pause height")

	// Check pause height when fisherman does not exist
	pauseHeight, err = db.GetFishermanPauseHeightIfExists(addrBz, 1)
	require.NoError(t, err)
	require.Equal(t, pauseHeight, DefaultPauseHeight, "unexpected pause height")
}

func TestSetFishermanPauseHeightAndUnstakeLater(t *testing.T) {
	db := NewTestPostgresContext(t, 0)

	fisherman, err := createAndInsertDefaultTestFisherman(db)
	require.NoError(t, err)

	pauseHeight := int64(1)
	unstakingHeight := pauseHeight + 10

	addrBz, err := hex.DecodeString(fisherman.Address)
	require.NoError(t, err)

	err = db.SetFishermanPauseHeight(addrBz, pauseHeight)
	require.NoError(t, err)

	_, _, _, _, _, fishermanPausedHeight, _, _, err := db.GetFisherman(addrBz, db.Height)
	require.NoError(t, err)
	require.Equal(t, pauseHeight, fishermanPausedHeight, "pause height not updated")

	err = db.SetFishermanStatusAndUnstakingHeightIfPausedBefore(pauseHeight+1, unstakingHeight, -1 /*unused*/)
	require.NoError(t, err)

	_, _, _, _, _, _, fishermanUnstakingHeight, _, err := db.GetFisherman(addrBz, db.Height)
	require.NoError(t, err)
	require.Equal(t, unstakingHeight, fishermanUnstakingHeight, "unstaking height was not set correctly")
}

func TestGetFishermanOutputAddress(t *testing.T) {
	db := NewTestPostgresContext(t, 0)

	fisherman, err := createAndInsertDefaultTestFisherman(db)
	require.NoError(t, err)

	addrBz, err := hex.DecodeString(fisherman.Address)
	require.NoError(t, err)

	output, err := db.GetFishermanOutputAddress(addrBz, 0)
	require.NoError(t, err)
	require.Equal(t, hex.EncodeToString(output), fisherman.Output, "unexpected output address")
}

func newTestFisherman() (*types.Actor, error) {
	operatorKey, err := crypto.GeneratePublicKey()
	if err != nil {
		return nil, err
	}

	outputAddr, err := crypto.GenerateAddress()
	if err != nil {
		return nil, err
	}

	return &types.Actor{
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

func createAndInsertDefaultTestFisherman(db *persistence.PostgresContext) (*types.Actor, error) {
	fisherman, err := newTestFisherman()
	if err != nil {
		return nil, err
	}
	addrBz, err := hex.DecodeString(fisherman.Address)
	if err != nil {
		log.Fatal(fmt.Sprintf("an error occurred converting address to bytes %s", fisherman.Address))
	}
	pubKeyBz, err := hex.DecodeString(fisherman.PublicKey)
	if err != nil {
		log.Fatal(fmt.Sprintf("an error occurred converting pubKey to bytes %s", fisherman.PublicKey))
	}
	outputBz, err := hex.DecodeString(fisherman.Output)
	if err != nil {
		log.Fatal(fmt.Sprintf("an error occurred converting output to bytes %s", fisherman.Output))
	}
	return fisherman, db.InsertFisherman(
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

func getTestFisherman(db *persistence.PostgresContext, address []byte) (*types.Actor, error) {
	operator, publicKey, stakedTokens, serviceURL, outputAddress, pauseHeight, unstakingHeight, chains, err := db.GetFisherman(address, db.Height)
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

	return &types.Actor{
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
