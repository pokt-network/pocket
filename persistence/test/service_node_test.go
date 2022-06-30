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

func FuzzServiceNode(f *testing.F) {
	fuzzSingleProtocolActor(f,
		NewTestGenericActor(schema.ServiceNodeActor, newTestServiceNode),
		GetGenericActor(schema.ServiceNodeActor, GetTestServiceNode),
		schema.ServiceNodeActor)
}

func TestInsertServiceNodeAndExists(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	serviceNode := NewTestServiceNode(t)
	serviceNode2 := NewTestServiceNode(t)
	err := db.InsertServiceNode(serviceNode.Address, serviceNode.PublicKey, serviceNode.Output, false, DefaultStakeStatus, DefaultStake, DefaultStake, DefaultChains, DefaultPauseHeight, DefaultUnstakingHeight)
	require.NoError(t, err)
	exists, err := db.GetServiceNodeExists(serviceNode.Address, db.Height)
	require.NoError(t, err)
	if !exists {
		t.Fatal("actor that should exist does not")
	}
	exists, err = db.GetServiceNodeExists(serviceNode2.Address, db.Height)
	require.NoError(t, err)
	if exists {
		t.Fatal("actor that should not exist, appears to")
	}
}

func TestUpdateServiceNode(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	serviceNode := NewTestServiceNode(t)
	err := db.InsertServiceNode(serviceNode.Address, serviceNode.PublicKey, serviceNode.Output, false, DefaultStakeStatus, DefaultStake, DefaultStake, DefaultChains, DefaultPauseHeight, DefaultUnstakingHeight)
	require.NoError(t, err)
	_, _, stakedTokens, _, _, _, _, chains, err := db.GetServiceNode(serviceNode.Address, db.Height)
	require.NoError(t, err)
	err = db.UpdateServiceNode(serviceNode.Address, serviceNode.ServiceUrl, StakeToUpdate, ChainsToUpdate)
	require.NoError(t, err)
	_, _, stakedTokens, _, _, _, _, chains, err = db.GetServiceNode(serviceNode.Address, db.Height)
	require.NoError(t, err)
	if chains[0] != ChainsToUpdate[0] {
		t.Fatal("chains not updated")
	}
	if stakedTokens != StakeToUpdate {
		t.Fatal("stake not updated")
	}
}

func TestGetServiceNodesReadyToUnstake(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	db.ClearAllDebug()
	serviceNode := NewTestServiceNode(t)
	err := db.InsertServiceNode(serviceNode.Address, serviceNode.PublicKey, serviceNode.Output, false, 1, DefaultStake, DefaultStake, DefaultChains, DefaultPauseHeight, DefaultUnstakingHeight)
	require.NoError(t, err)
	// test SetServiceNodeUnstakingHeightAndStatus
	err = db.SetServiceNodeUnstakingHeightAndStatus(serviceNode.Address, 0, 1)
	require.NoError(t, err)
	// test GetServiceNodesReadyToUnstake
	serviceNodes, err := db.GetServiceNodesReadyToUnstake(0, 1)
	require.NoError(t, err)
	if len(serviceNodes) != 1 {
		t.Fatal("wrong number of actors")
	}
	if !bytes.Equal(serviceNode.Address, serviceNodes[0].Address) {
		t.Fatal("unexpected actor returned")
	}
}

func TestGetServiceNodeStatus(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	serviceNode := NewTestServiceNode(t)
	err := db.InsertServiceNode(serviceNode.Address, serviceNode.PublicKey, serviceNode.Output, false, 1, DefaultStake, DefaultStake, DefaultChains, DefaultPauseHeight, DefaultUnstakingHeight)
	require.NoError(t, err)
	status, err := db.GetServiceNodeStatus(serviceNode.Address, db.Height)
	require.NoError(t, err)
	if status != DefaultStakeStatus {
		t.Fatalf("unexpected status: got %d expected %d", status, DefaultStakeStatus)
	}
}

func TestGetServiceNodePauseHeightIfExists(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	serviceNode := NewTestServiceNode(t)
	err := db.InsertServiceNode(serviceNode.Address, serviceNode.PublicKey, serviceNode.Output, false, 1, DefaultStake, DefaultStake, DefaultChains, DefaultPauseHeight, DefaultUnstakingHeight)
	require.NoError(t, err)
	height, err := db.GetServiceNodePauseHeightIfExists(serviceNode.Address, db.Height)
	require.NoError(t, err)
	if height != DefaultPauseHeight {
		t.Fatalf("unexpected pausedHeight: got %d expected %d", DefaultPauseHeight, DefaultStakeStatus)
	}
}

func TestSetServiceNodeStatusAndUnstakingHeightIfPausedBefore(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	serviceNode := NewTestServiceNode(t)
	err := db.InsertServiceNode(serviceNode.Address, serviceNode.PublicKey, serviceNode.Output, false, 1, DefaultStake, DefaultStake, DefaultChains, 0, DefaultUnstakingHeight)
	require.NoError(t, err)
	err = db.SetServiceNodeStatusAndUnstakingHeightIfPausedBefore(1, 0, 1)
	require.NoError(t, err)
	_, _, _, _, _, _, unstakingHeight, _, err := db.GetServiceNode(serviceNode.Address, db.Height)
	require.NoError(t, err)
	if unstakingHeight != 0 {
		t.Fatal("unexpected unstaking height")
	}
}

func TestSetServiceNodePauseHeight(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	serviceNode := NewTestServiceNode(t)
	err := db.InsertServiceNode(serviceNode.Address, serviceNode.PublicKey, serviceNode.Output, false, 1, DefaultStake, DefaultStake, DefaultChains, 0, DefaultUnstakingHeight)
	require.NoError(t, err)
	err = db.SetServiceNodePauseHeight(serviceNode.Address, 1)
	require.NoError(t, err)
	_, _, _, _, _, pausedHeight, _, _, err := db.GetServiceNode(serviceNode.Address, db.Height)
	require.NoError(t, err)
	if pausedHeight != 1 {
		t.Fatal("unexpected pause height")
	}
}

func TestGetServiceNodeOutputAddress(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	serviceNode := NewTestServiceNode(t)
	err := db.InsertServiceNode(serviceNode.Address, serviceNode.PublicKey, serviceNode.Output, false, 1, DefaultStake, DefaultStake, DefaultChains, 0, DefaultUnstakingHeight)
	require.NoError(t, err)
	output, err := db.GetServiceNodeOutputAddress(serviceNode.Address, db.Height)
	require.NoError(t, err)
	if !bytes.Equal(output, serviceNode.Output) {
		t.Fatal("unexpected output address")
	}
}

func TestServiceNodeCount(t *testing.T) {
	//db := persistence.PostgresContext{ TODO implement
	//	Height: 0,
	//	DB:     *PostgresDB,
	//}
	//err := db.ClearAllDebug()
	//require.NoError(t, err)
	//count, err := db.GetServiceNodeCount(DefaultChains[0], 0)
	//require.NoError(t, err)
	//if count != 0 {
	//	t.Fatal("unexpected service node count")
	//}
	//serviceNode := NewTestServiceNode(t)
	//err = db.InsertServiceNode(serviceNode.Address, serviceNode.PublicKey, serviceNode.Output, false, 1, DefaultStake, DefaultStake, DefaultChains, -1, DefaultUnstakingHeight)
	//require.NoError(t, err)
	//count, err = db.GetServiceNodeCount(DefaultChains[0], 0)
	//require.NoError(t, err)
	//if count != 1 {
	//	t.Fatal("unexpected service node count")
	//}
}

func NewTestServiceNode(t *testing.T) typesGenesis.ServiceNode {
	sn, err := newTestServiceNode()
	require.NoError(t, err)
	return sn
}

func newTestServiceNode() (typesGenesis.ServiceNode, error) {
	pub1, err := crypto.GeneratePublicKey()
	if err != nil {
		return typesGenesis.ServiceNode{}, nil
	}
	addr1 := pub1.Address()
	addr2, err := crypto.GenerateAddress()
	if err != nil {
		return typesGenesis.ServiceNode{}, nil
	}
	return typesGenesis.ServiceNode{
		Address:         addr1,
		PublicKey:       pub1.Bytes(),
		Paused:          false,
		Status:          typesGenesis.DefaultStakeStatus,
		Chains:          typesGenesis.DefaultChains,
		ServiceUrl:      DefaultServiceUrl,
		StakedTokens:    typesGenesis.DefaultStake,
		PausedHeight:    uint64(DefaultPauseHeight),
		UnstakingHeight: DefaultUnstakingHeight,
		Output:          addr2,
	}, nil
}

func GetTestServiceNode(db persistence.PostgresContext, address []byte) (*typesGenesis.ServiceNode, error) {
	operator, publicKey, stakedTokens, serviceURL, outputAddress, pauseHeight, unstakingHeight, chains, err := db.GetServiceNode(address, db.Height)
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
	return &typesGenesis.ServiceNode{
		Address:         addr,
		PublicKey:       pubKey,
		Paused:          false,
		Status:          int32(status),
		Chains:          chains,
		ServiceUrl:      serviceURL,
		StakedTokens:    stakedTokens,
		PausedHeight:    uint64(pauseHeight),
		UnstakingHeight: unstakingHeight,
		Output:          outputAddr,
	}, nil
}
