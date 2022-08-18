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

func FuzzFisherman(f *testing.F) {
	fuzzSingleProtocolActor(f,
		NewTestGenericActor(schema.FishermanActor, newTestFisherman),
		GetGenericActor(schema.FishermanActor, getTestFisherman),
		schema.FishermanActor)
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

	exists, err := db.GetFishermanExists(fisherman.Address, 0)
	require.NoError(t, err)
	require.True(t, exists, "actor that should exist at previous height does not")

	exists, err = db.GetFishermanExists(fisherman.Address, 1)
	require.NoError(t, err)
	require.True(t, exists, "actor that should exist at current height does not")

	exists, err = db.GetFishermanExists(fisherman2.Address, 0)
	require.NoError(t, err)
	require.False(t, exists, "actor that should not exist at previous height appears to")

	exists, err = db.GetFishermanExists(fisherman2.Address, 1)
	require.NoError(t, err)
	require.True(t, exists, "actor that should exist at current height does not")
}

func TestUpdateFisherman(t *testing.T) {
	db := NewTestPostgresContext(t, 0)

	fisherman, err := createAndInsertDefaultTestFisherman(db)
	require.NoError(t, err)

	_, _, stakedTokens, _, _, _, _, chains, err := db.GetFisherman(fisherman.Address, 0)
	require.NoError(t, err)
	require.Equal(t, DefaultChains, chains, "default chains incorrect for current height")
	require.Equal(t, DefaultStake, stakedTokens, "default stake incorrect for current height")

	db.Height = 1

	require.NotEqual(t, DefaultStake, StakeToUpdate)   // sanity check to make sure the tests are correct
	require.NotEqual(t, DefaultChains, ChainsToUpdate) // sanity check to make sure the tests are correct
	err = db.UpdateFisherman(fisherman.Address, fisherman.ServiceUrl, StakeToUpdate, ChainsToUpdate)
	require.NoError(t, err)

	_, _, stakedTokens, _, _, _, _, chains, err = db.GetFisherman(fisherman.Address, 0)
	require.NoError(t, err)
	require.Equal(t, DefaultChains, chains, "default chains incorrect for previous height")
	require.Equal(t, DefaultStake, stakedTokens, "default stake incorrect for previous height")

	_, _, stakedTokens, _, _, _, _, chains, err = db.GetFisherman(fisherman.Address, 1)
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

	// Unstake fisherman at height 0
	err = db.SetFishermanUnstakingHeightAndStatus(fisherman.Address, 0, persistence.UnstakingStatus)
	require.NoError(t, err)

	// Unstake fisherman2 and fisherman3 at height 1
	err = db.SetFishermanUnstakingHeightAndStatus(fisherman2.Address, 1, persistence.UnstakingStatus)
	require.NoError(t, err)
	err = db.SetFishermanUnstakingHeightAndStatus(fisherman3.Address, 1, persistence.UnstakingStatus)
	require.NoError(t, err)

	// Check unstaking fishermans at height 0
	unstakingFishermen, err := db.GetFishermenReadyToUnstake(0, persistence.UnstakingStatus)
	require.NoError(t, err)
	require.Equal(t, 1, len(unstakingFishermen), "wrong number of actors ready to unstake at height 0")
	require.Equal(t, fisherman.Address, unstakingFishermen[0].Address, "unexpected fishermanlication actor returned")

	// Check unstaking fishermans at height 1
	unstakingFishermen, err = db.GetFishermenReadyToUnstake(1, persistence.UnstakingStatus)
	require.NoError(t, err)
	require.Equal(t, 2, len(unstakingFishermen), "wrong number of actors ready to unstake at height 1")
	require.ElementsMatch(t, [][]byte{fisherman2.Address, fisherman3.Address}, [][]byte{unstakingFishermen[0].Address, unstakingFishermen[1].Address})
}

func TestGetFishermanStatus(t *testing.T) {
	db := NewTestPostgresContext(t, 1)

	fisherman, err := createAndInsertDefaultTestFisherman(db)
	require.NoError(t, err)

	// Check status before the fisherman exists
	status, err := db.GetFishermanStatus(fisherman.Address, 0)
	require.Error(t, err)
	require.Equal(t, status, persistence.UndefinedStakingStatus, "unexpected status")

	// Check status after the fisherman exists
	status, err = db.GetFishermanStatus(fisherman.Address, 1)
	require.NoError(t, err)
	require.Equal(t, status, DefaultStakeStatus, "unexpected status")
}

func TestGetFishermanPauseHeightIfExists(t *testing.T) {
	db := NewTestPostgresContext(t, 1)

	fisherman, err := createAndInsertDefaultTestFisherman(db)
	require.NoError(t, err)

	// Check pause height when fisherman does not exist
	pauseHeight, err := db.GetFishermanPauseHeightIfExists(fisherman.Address, 0)
	require.Error(t, err)
	require.Equal(t, pauseHeight, DefaultPauseHeight, "unexpected pause height")

	// Check pause height when fisherman does not exist
	pauseHeight, err = db.GetFishermanPauseHeightIfExists(fisherman.Address, 1)
	require.NoError(t, err)
	require.Equal(t, pauseHeight, DefaultPauseHeight, "unexpected pause height")
}

func TestSetFishermanPauseHeightAndUnstakeLater(t *testing.T) {
	db := NewTestPostgresContext(t, 0)

	fisherman, err := createAndInsertDefaultTestFisherman(db)
	require.NoError(t, err)

	pauseHeight := int64(1)
	unstakingHeight := pauseHeight + 10

	err = db.SetFishermanPauseHeight(fisherman.Address, pauseHeight)
	require.NoError(t, err)

	_, _, _, _, _, fishermanPausedHeight, _, _, err := db.GetFisherman(fisherman.Address, db.Height)
	require.NoError(t, err)
	require.Equal(t, pauseHeight, fishermanPausedHeight, "pause height not updated")

	err = db.SetFishermanStatusAndUnstakingHeightIfPausedBefore(pauseHeight+1, unstakingHeight, -1 /*unused*/)
	require.NoError(t, err)

	_, _, _, _, _, _, fishermanUnstakingHeight, _, err := db.GetFisherman(fisherman.Address, db.Height)
	require.NoError(t, err)
	require.Equal(t, unstakingHeight, fishermanUnstakingHeight, "unstaking height was not set correctly")
}

func TestGetFishermanOutputAddress(t *testing.T) {
	db := NewTestPostgresContext(t, 0)

	fisherman, err := createAndInsertDefaultTestFisherman(db)
	require.NoError(t, err)

	output, err := db.GetFishermanOutputAddress(fisherman.Address, 0)
	require.NoError(t, err)
	require.Equal(t, output, fisherman.Output, "unexpected output address")
}

func newTestFisherman() (*typesGenesis.Fisherman, error) {
	operatorKey, err := crypto.GeneratePublicKey()
	if err != nil {
		return nil, err
	}

	outputAddr, err := crypto.GenerateAddress()
	if err != nil {
		return nil, err
	}

	return &typesGenesis.Fisherman{
		Address:         operatorKey.Address(),
		PublicKey:       operatorKey.Bytes(),
		Paused:          false,
		Status:          typesGenesis.DefaultStakeStatus,
		Chains:          typesGenesis.DefaultChains,
		ServiceUrl:      DefaultServiceUrl,
		StakedTokens:    typesGenesis.DefaultStake,
		PausedHeight:    DefaultPauseHeight,
		UnstakingHeight: DefaultUnstakingHeight,
		Output:          outputAddr,
	}, nil
}

func TestGetAllFish(t *testing.T) {
	db := NewTestPostgresContext(t, 0)

	updateFish := func(db *persistence.PostgresContext, fish *genesis.Fisherman) error {
		return db.UpdateFisherman(fish.Address, OlshanskyURL, fish.StakedTokens, OlshanskyChains)
	}

	getAllActorsTest(t, db, db.GetAllFishermen, createAndInsertDefaultTestFisherman, updateFish, 1)
}

func createAndInsertDefaultTestFisherman(db *persistence.PostgresContext) (*typesGenesis.Fisherman, error) {
	fisherman, err := newTestFisherman()
	if err != nil {
		return nil, err
	}

	return fisherman, db.InsertFisherman(
		fisherman.Address,
		fisherman.PublicKey,
		fisherman.Output,
		false,
		DefaultStakeStatus,
		DefaultServiceUrl,
		DefaultStake,
		DefaultChains,
		DefaultPauseHeight,
		DefaultUnstakingHeight)
}

func getTestFisherman(db *persistence.PostgresContext, address []byte) (*typesGenesis.Fisherman, error) {
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

	return &typesGenesis.Fisherman{
		Address:         operatorAddr,
		PublicKey:       operatorPubKey,
		Paused:          false,
		Status:          persistence.UnstakingHeightToStatus(unstakingHeight),
		Chains:          chains,
		ServiceUrl:      serviceURL,
		StakedTokens:    stakedTokens,
		PausedHeight:    pauseHeight,
		UnstakingHeight: unstakingHeight,
		Output:          outputAddr,
	}, nil
}
