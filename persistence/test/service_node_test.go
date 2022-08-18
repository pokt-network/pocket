package test

import (
	"encoding/hex"
	"fmt"
	"log"
	"testing"

	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/persistence/schema"
	"github.com/pokt-network/pocket/shared/crypto"
	typesGenesis "github.com/pokt-network/pocket/shared/types/genesis"
	"github.com/stretchr/testify/require"
)

func FuzzServiceNode(f *testing.F) {
	fuzzSingleProtocolActor(f,
		NewTestGenericActor(schema.ServiceNodeActor, newTestServiceNode),
		GetGenericActor(schema.ServiceNodeActor, getTestServiceNode),
		schema.ServiceNodeActor)
}

func TestGetSetServiceNodeStakeAmount(t *testing.T) {
	db := NewTestPostgresContext(t, 1)
	getTestGetSetStakeAmountTest(t, db, createAndInsertDefaultTestServiceNode, db.GetServiceNodeStakeAmount, db.SetServiceNodeStakeAmount, 1)
}

func TestInsertServiceNodeAndExists(t *testing.T) {
	db := NewTestPostgresContext(t, 0)

	serviceNode, err := createAndInsertDefaultTestServiceNode(db)
	require.NoError(t, err)

	db.Height = 1

	serviceNode2, err := createAndInsertDefaultTestServiceNode(db)
	require.NoError(t, err)

	addrBz, err := hex.DecodeString(serviceNode.Address)
	require.NoError(t, err)
	addrBz2, err := hex.DecodeString(serviceNode2.Address)
	require.NoError(t, err)

	exists, err := db.GetServiceNodeExists(addrBz, 0)
	require.NoError(t, err)
	require.True(t, exists, "actor that should exist at previous height does not")
	exists, err = db.GetServiceNodeExists(addrBz, 1)
	require.NoError(t, err)
	require.True(t, exists, "actor that should exist at current height does not")

	exists, err = db.GetServiceNodeExists(addrBz2, 0)
	require.NoError(t, err)
	require.False(t, exists, "actor that should not exist at previous height serviceNodeears to")
	exists, err = db.GetServiceNodeExists(addrBz2, 1)
	require.NoError(t, err)
	require.True(t, exists, "actor that should exist at current height does not")
}

func TestUpdateServiceNode(t *testing.T) {
	db := NewTestPostgresContext(t, 0)

	serviceNode, err := createAndInsertDefaultTestServiceNode(db)
	require.NoError(t, err)

	addrBz, err := hex.DecodeString(serviceNode.Address)
	require.NoError(t, err)

	_, _, stakedTokens, _, _, _, _, chains, err := db.GetServiceNode(addrBz, 0)
	require.NoError(t, err)
	require.Equal(t, DefaultChains, chains, "default chains incorrect for current height")
	require.Equal(t, DefaultStake, stakedTokens, "default stake incorrect for current height")

	db.Height = 1

	require.NotEqual(t, DefaultStake, StakeToUpdate)   // sanity check to make sure the tests are correct
	require.NotEqual(t, DefaultChains, ChainsToUpdate) // sanity check to make sure the tests are correct
	err = db.UpdateServiceNode(addrBz, serviceNode.GenericParam, StakeToUpdate, ChainsToUpdate)
	require.NoError(t, err)

	_, _, stakedTokens, _, _, _, _, chains, err = db.GetServiceNode(addrBz, 0)
	require.NoError(t, err)
	require.Equal(t, DefaultChains, chains, "default chains incorrect for previous height")
	require.Equal(t, DefaultStake, stakedTokens, "default stake incorrect for previous height")

	_, _, stakedTokens, _, _, _, _, chains, err = db.GetServiceNode(addrBz, 1)
	require.NoError(t, err)
	require.Equal(t, ChainsToUpdate, chains, "chains not updated for current height")
	require.Equal(t, StakeToUpdate, stakedTokens, "stake not updated for current height")
}

func TestGetServiceNodesReadyToUnstake(t *testing.T) {
	db := NewTestPostgresContext(t, 0)

	serviceNode, err := createAndInsertDefaultTestServiceNode(db)
	require.NoError(t, err)

	serviceNode2, err := createAndInsertDefaultTestServiceNode(db)
	require.NoError(t, err)

	serviceNode3, err := createAndInsertDefaultTestServiceNode(db)
	require.NoError(t, err)

	addrBz, err := hex.DecodeString(serviceNode.Address)
	require.NoError(t, err)

	addrBz2, err := hex.DecodeString(serviceNode2.Address)
	require.NoError(t, err)

	addrBz3, err := hex.DecodeString(serviceNode3.Address)
	require.NoError(t, err)

	// Unstake serviceNode at height 0
	err = db.SetServiceNodeUnstakingHeightAndStatus(addrBz, 0, persistence.UnstakingStatus)
	require.NoError(t, err)

	// Unstake serviceNode2 and serviceNode3 at height 1
	err = db.SetServiceNodeUnstakingHeightAndStatus(addrBz2, 1, persistence.UnstakingStatus)
	require.NoError(t, err)
	err = db.SetServiceNodeUnstakingHeightAndStatus(addrBz3, 1, persistence.UnstakingStatus)
	require.NoError(t, err)

	// Check unstaking serviceNodes at height 0
	unstakingServiceNodes, err := db.GetServiceNodesReadyToUnstake(0, persistence.UnstakingStatus)
	require.NoError(t, err)
	require.Equal(t, 1, len(unstakingServiceNodes), "wrong number of actors ready to unstake at height 0")
	require.Equal(t, serviceNode.Address, hex.EncodeToString(unstakingServiceNodes[0].Address), "unexpected serviceNodelication actor returned")

	// Check unstaking serviceNodes at height 1
	unstakingServiceNodes, err = db.GetServiceNodesReadyToUnstake(1, persistence.UnstakingStatus)
	require.NoError(t, err)
	require.Equal(t, 2, len(unstakingServiceNodes), "wrong number of actors ready to unstake at height 1")
	require.ElementsMatch(t, [][]byte{addrBz2, addrBz3}, [][]byte{unstakingServiceNodes[0].Address, unstakingServiceNodes[1].Address})
}

func TestGetServiceNodeStatus(t *testing.T) {
	db := &persistence.PostgresContext{
		Height: 1, // intentionally set to a non-zero height
		DB:     *testPostgresDB,
	}

	serviceNode, err := createAndInsertDefaultTestServiceNode(db)
	require.NoError(t, err)

	addrBz, err := hex.DecodeString(serviceNode.Address)
	require.NoError(t, err)

	// Check status before the serviceNode exists
	status, err := db.GetServiceNodeStatus(addrBz, 0)
	require.Error(t, err)
	require.Equal(t, status, persistence.UndefinedStakingStatus, "unexpected status")

	// Check status after the serviceNode exists
	status, err = db.GetServiceNodeStatus(addrBz, 1)
	require.NoError(t, err)
	require.Equal(t, status, DefaultStakeStatus, "unexpected status")
}

func TestGetServiceNodePauseHeightIfExists(t *testing.T) {
	db := &persistence.PostgresContext{
		Height: 1, // intentionally set to a non-zero height
		DB:     *testPostgresDB,
	}

	serviceNode, err := createAndInsertDefaultTestServiceNode(db)
	require.NoError(t, err)

	addrBz, err := hex.DecodeString(serviceNode.Address)
	require.NoError(t, err)

	// Check pause height when serviceNode does not exist
	pauseHeight, err := db.GetServiceNodePauseHeightIfExists(addrBz, 0)
	require.Error(t, err)
	require.Equal(t, pauseHeight, DefaultPauseHeight, "unexpected pause height")

	// Check pause height when serviceNode does not exist
	pauseHeight, err = db.GetServiceNodePauseHeightIfExists(addrBz, 1)
	require.NoError(t, err)
	require.Equal(t, pauseHeight, DefaultPauseHeight, "unexpected pause height")
}

func TestSetServiceNodePauseHeightAndUnstakeLater(t *testing.T) {
	db := NewTestPostgresContext(t, 0)

	serviceNode, err := createAndInsertDefaultTestServiceNode(db)
	require.NoError(t, err)

	pauseHeight := int64(1)
	unstakingHeight := pauseHeight + 10

	addrBz, err := hex.DecodeString(serviceNode.Address)
	require.NoError(t, err)

	err = db.SetServiceNodePauseHeight(addrBz, pauseHeight)
	require.NoError(t, err)

	_, _, _, _, _, serviceNodePausedHeight, _, _, err := db.GetServiceNode(addrBz, db.Height)
	require.NoError(t, err)
	require.Equal(t, pauseHeight, serviceNodePausedHeight, "pause height not updated")

	err = db.SetServiceNodeStatusAndUnstakingHeightIfPausedBefore(pauseHeight+1, unstakingHeight, -1 /*unused*/)
	require.NoError(t, err)

	_, _, _, _, _, _, serviceNodeUnstakingHeight, _, err := db.GetServiceNode(addrBz, db.Height)
	require.NoError(t, err)
	require.Equal(t, unstakingHeight, serviceNodeUnstakingHeight, "unstaking height was not set correctly")
}

func TestGetServiceNodeOutputAddress(t *testing.T) {
	db := NewTestPostgresContext(t, 0)

	serviceNode, err := createAndInsertDefaultTestServiceNode(db)
	require.NoError(t, err)

	addrBz, err := hex.DecodeString(serviceNode.Address)
	require.NoError(t, err)

	output, err := db.GetServiceNodeOutputAddress(addrBz, 0)
	require.NoError(t, err)
	require.Equal(t, hex.EncodeToString(output), serviceNode.Output, "unexpected output address")
}

func newTestServiceNode() (*typesGenesis.Actor, error) {
	operatorKey, err := crypto.GeneratePublicKey()
	if err != nil {
		return nil, err
	}

	outputAddr, err := crypto.GenerateAddress()
	if err != nil {
		return nil, err
	}

	return &typesGenesis.Actor{
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

func createAndInsertDefaultTestServiceNode(db *persistence.PostgresContext) (*typesGenesis.Actor, error) {
	serviceNode, err := newTestServiceNode()
	if err != nil {
		return nil, err
	}
	addrBz, err := hex.DecodeString(serviceNode.Address)
	if err != nil {
		log.Fatal(fmt.Sprintf("an error occurred converting address to bytes %s", serviceNode.Address))
	}
	pubKeyBz, err := hex.DecodeString(serviceNode.PublicKey)
	if err != nil {
		log.Fatal(fmt.Sprintf("an error occurred converting pubKey to bytes %s", serviceNode.PublicKey))
	}
	outputBz, err := hex.DecodeString(serviceNode.Output)
	if err != nil {
		log.Fatal(fmt.Sprintf("an error occurred converting output to bytes %s", serviceNode.Output))
	}
	return serviceNode, db.InsertServiceNode(
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

func getTestServiceNode(db persistence.PostgresContext, address []byte) (*typesGenesis.Actor, error) {
	operator, publicKey, stakedTokens, serviceURL, outputAddress, pauseHeight, unstakingHeight, chains, err := db.GetServiceNode(address, db.Height)
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

	return &typesGenesis.Actor{
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
