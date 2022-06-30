package test

import (
	"encoding/hex"
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
		GetGenericActor(schema.ServiceNodeActor, GetTestServiceNode),
		schema.ServiceNodeActor)
}

func TestInsertServiceNodeAndExists(t *testing.T) {
	db := &persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}

	serviceNode, err := createAndInsertDefaultTestServiceNode(db)
	require.NoError(t, err)

	db.Height = 1

	serviceNode2, err := createAndInsertDefaultTestServiceNode(db)
	require.NoError(t, err)

	exists, err := db.GetServiceNodeExists(serviceNode.Address, 0)
	require.NoError(t, err)
	require.True(t, exists, "actor that should exist at previous height does not")
	exists, err = db.GetServiceNodeExists(serviceNode.Address, 1)
	require.NoError(t, err)
	require.True(t, exists, "actor that should exist at current height does not")

	exists, err = db.GetServiceNodeExists(serviceNode2.Address, 0)
	require.NoError(t, err)
	require.False(t, exists, "actor that should not exist at previous height serviceNodeears to")
	exists, err = db.GetServiceNodeExists(serviceNode2.Address, 1)
	require.NoError(t, err)
	require.True(t, exists, "actor that should exist at current height does not")
}

func TestUpdateServiceNode(t *testing.T) {
	db := &persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}

	serviceNode, err := createAndInsertDefaultTestServiceNode(db)
	require.NoError(t, err)

	_, _, stakedTokens, _, _, _, _, chains, err := db.GetServiceNode(serviceNode.Address, 0)
	require.NoError(t, err)
	require.Equal(t, DefaultChains, chains, "default chains incorrect for current height")
	require.Equal(t, DefaultStake, stakedTokens, "default stake incorrect for current height")

	db.Height = 1

	require.NotEqual(t, DefaultStake, StakeToUpdate)   // sanity check to make sure the tests are correct
	require.NotEqual(t, DefaultChains, ChainsToUpdate) // sanity check to make sure the tests are correct
	err = db.UpdateServiceNode(serviceNode.Address, serviceNode.ServiceUrl, StakeToUpdate, ChainsToUpdate)
	require.NoError(t, err)

	_, _, stakedTokens, _, _, _, _, chains, err = db.GetServiceNode(serviceNode.Address, 0)
	require.NoError(t, err)
	require.Equal(t, DefaultChains, chains, "default chains incorrect for previous height")
	require.Equal(t, DefaultStake, stakedTokens, "default stake incorrect for previous height")

	_, _, stakedTokens, _, _, _, _, chains, err = db.GetServiceNode(serviceNode.Address, 1)
	require.NoError(t, err)
	require.Equal(t, ChainsToUpdate, chains, "chains not updated for current height")
	require.Equal(t, StakeToUpdate, stakedTokens, "stake not updated for current height")
}

func TestGetServiceNodesReadyToUnstake(t *testing.T) {
	db := &persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}

	serviceNode, err := createAndInsertDefaultTestServiceNode(db)
	require.NoError(t, err)

	serviceNode2, err := createAndInsertDefaultTestServiceNode(db)
	require.NoError(t, err)

	serviceNode3, err := createAndInsertDefaultTestServiceNode(db)
	require.NoError(t, err)

	// Unstake serviceNode at height 0
	err = db.SetServiceNodeUnstakingHeightAndStatus(serviceNode.Address, 0, persistence.UnstakingStatus)
	require.NoError(t, err)

	// Unstake serviceNode2 and serviceNode3 at height 1
	err = db.SetServiceNodeUnstakingHeightAndStatus(serviceNode2.Address, 1, persistence.UnstakingStatus)
	require.NoError(t, err)
	err = db.SetServiceNodeUnstakingHeightAndStatus(serviceNode3.Address, 1, persistence.UnstakingStatus)
	require.NoError(t, err)

	// Check unstaking serviceNodes at height 0
	unstakingServiceNodes, err := db.GetServiceNodesReadyToUnstake(0, persistence.UnstakingStatus)
	require.NoError(t, err)
	require.Equal(t, 1, len(unstakingServiceNodes), "wrong number of actors ready to unstake at height 0")
	require.Equal(t, serviceNode.Address, unstakingServiceNodes[0].Address, "unexpected serviceNodelication actor returned")

	// Check unstaking serviceNodes at height 1
	unstakingServiceNodes, err = db.GetServiceNodesReadyToUnstake(1, persistence.UnstakingStatus)
	require.NoError(t, err)
	require.Equal(t, 2, len(unstakingServiceNodes), "wrong number of actors ready to unstake at height 1")
	require.ElementsMatch(t, [][]byte{serviceNode2.Address, serviceNode3.Address}, [][]byte{unstakingServiceNodes[0].Address, unstakingServiceNodes[1].Address})
}

func TestGetServiceNodeStatus(t *testing.T) {
	db := &persistence.PostgresContext{
		Height: 1, // intentionally set to a non-zero height
		DB:     *PostgresDB,
	}

	serviceNode, err := newTestServiceNode()
	require.NoError(t, err)

	err = db.InsertServiceNode(
		serviceNode.Address,
		serviceNode.PublicKey,
		serviceNode.Output,
		false,
		DefaultStakeStatus,
		DefaultMaxRelays,
		DefaultStake,
		DefaultChains,
		DefaultPauseHeight,
		DefaultUnstakingHeight)
	require.NoError(t, err)

	// Check status before the serviceNode exists
	status, err := db.GetServiceNodeStatus(serviceNode.Address, 0)
	require.Error(t, err)
	require.Equal(t, status, persistence.UndefinedStakingStatus, "unexpected status")

	// Check status after the serviceNode exists
	status, err = db.GetServiceNodeStatus(serviceNode.Address, 1)
	require.NoError(t, err)
	require.Equal(t, status, DefaultStakeStatus, "unexpected status")
}

func TestGetServiceNodePauseHeightIfExists(t *testing.T) {
	db := &persistence.PostgresContext{
		Height: 1, // intentionally set to a non-zero height
		DB:     *PostgresDB,
	}

	serviceNode, err := createAndInsertDefaultTestServiceNode(db)
	require.NoError(t, err)

	// Check pause height when serviceNode does not exist
	pauseHeight, err := db.GetServiceNodePauseHeightIfExists(serviceNode.Address, 0)
	require.Error(t, err)
	require.Equal(t, pauseHeight, DefaultPauseHeight, "unexpected pause height")

	// Check pause height when serviceNode does not exist
	pauseHeight, err = db.GetServiceNodePauseHeightIfExists(serviceNode.Address, 1)
	require.NoError(t, err)
	require.Equal(t, pauseHeight, DefaultPauseHeight, "unexpected pause height")
}

func TestSetServiceNodeStatusAndUnstakingHeightIfPausedBefore(t *testing.T) {
	db := &persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}

	serviceNode, err := newTestServiceNode()
	require.NoError(t, err)

	err = db.InsertServiceNode(
		serviceNode.Address,
		serviceNode.PublicKey,
		serviceNode.Output,
		false,
		DefaultStakeStatus,
		DefaultMaxRelays,
		DefaultStake,
		DefaultChains,
		0,
		DefaultUnstakingHeight)
	require.NoError(t, err)

	unstakingHeightSet := int64(0)
	err = db.SetServiceNodeStatusAndUnstakingHeightIfPausedBefore(1, unstakingHeightSet, -1)
	require.NoError(t, err)

	_, _, _, _, _, unstakingHeight, _, _, err := db.GetServiceNode(serviceNode.Address, db.Height)
	require.NoError(t, err)
	require.Equal(t, unstakingHeightSet, unstakingHeight, "unstaking height was not set correctly")
}

func TestSetServiceNodePauseHeightAndUnstake(t *testing.T) {
	db := &persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}

	serviceNode, err := createAndInsertDefaultTestServiceNode(db)
	require.NoError(t, err)

	pauseHeight := int64(1)
	unstakingHeight := pauseHeight + 10

	err = db.SetServiceNodePauseHeight(serviceNode.Address, pauseHeight)
	require.NoError(t, err)

	_, _, _, _, _, serviceNodePausedHeight, _, _, err := db.GetServiceNode(serviceNode.Address, db.Height)
	require.NoError(t, err)
	require.Equal(t, pauseHeight, serviceNodePausedHeight, "pause height not updated")

	err = db.SetServiceNodeStatusAndUnstakingHeightIfPausedBefore(pauseHeight+1, unstakingHeight, -1 /*unused*/)
	require.NoError(t, err)

	_, _, _, _, _, _, serviceNodeUnstakingHeight, _, err := db.GetServiceNode(serviceNode.Address, db.Height)
	require.NoError(t, err)
	require.Equal(t, unstakingHeight, serviceNodeUnstakingHeight, "unstaking height was not set correctly")
}

func TestGetServiceNodeOutputAddress(t *testing.T) {
	db := &persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}

	serviceNode, err := createAndInsertDefaultTestServiceNode(db)
	require.NoError(t, err)

	output, err := db.GetServiceNodeOutputAddress(serviceNode.Address, 0)
	require.NoError(t, err)
	require.Equal(t, output, serviceNode.Output, "unexpected output address")
}

func newTestServiceNode() (*typesGenesis.ServiceNode, error) {
	operatorKey, err := crypto.GeneratePublicKey()
	if err != nil {
		return nil, err
	}

	outputAddr, err := crypto.GenerateAddress()
	if err != nil {
		return nil, err
	}

	return &typesGenesis.ServiceNode{
		Address:         operatorKey.Address(),
		PublicKey:       operatorKey.Bytes(),
		Paused:          false,
		Status:          typesGenesis.DefaultStakeStatus,
		Chains:          typesGenesis.DefaultChains,
		ServiceUrl:      DefaultServiceUrl,
		StakedTokens:    typesGenesis.DefaultStake,
		PausedHeight:    uint64(DefaultPauseHeight),
		UnstakingHeight: DefaultUnstakingHeight,
		Output:          outputAddr,
	}, nil
}

func createAndInsertDefaultTestServiceNode(db *persistence.PostgresContext) (*typesGenesis.ServiceNode, error) {
	serviceNode, err := newTestServiceNode()
	if err != nil {
		return nil, err
	}

	return serviceNode, db.InsertServiceNode(
		serviceNode.Address,
		serviceNode.PublicKey,
		serviceNode.Output,
		false,
		DefaultStakeStatus,
		DefaultServiceUrl,
		DefaultStake,
		DefaultChains,
		DefaultPauseHeight,
		DefaultUnstakingHeight)
}

func GetTestServiceNode(db persistence.PostgresContext, address []byte) (*typesGenesis.ServiceNode, error) {
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

	return &typesGenesis.ServiceNode{
		Address:         operatorAddr,
		PublicKey:       operatorPubKey,
		Paused:          false,
		Status:          persistence.UnstakingHeightToStatus(unstakingHeight),
		Chains:          chains,
		ServiceUrl:      serviceURL,
		StakedTokens:    stakedTokens,
		PausedHeight:    uint64(pauseHeight),
		UnstakingHeight: unstakingHeight,
		Output:          outputAddr,
	}, nil
}
