package ibc

import (
	"testing"
	"time"

	light_client_types "github.com/pokt-network/pocket/ibc/client/light_clients/types"
	client_types "github.com/pokt-network/pocket/ibc/client/types"
	"github.com/pokt-network/pocket/shared/codec"
	"github.com/stretchr/testify/require"
)

func TestHost_GetCurrentHeight(t *testing.T) {
	_, _, _, _, ibcMod := prepareEnvironment(t, 1, 0, 0, 0)
	cm := ibcMod.GetBus().GetClientManager()

	// get the current height
	height, err := cm.GetCurrentHeight()
	require.NoError(t, err)
	require.Equal(t, uint64(1), height.GetRevisionNumber())
	require.Equal(t, uint64(0), height.GetRevisionHeight())

	// increment the height
	publishNewHeightEvent(t, ibcMod.GetBus(), 1)

	height, err = cm.GetCurrentHeight()
	require.NoError(t, err)
	require.Equal(t, uint64(1), height.GetRevisionNumber())
	require.Equal(t, uint64(1), height.GetRevisionHeight())
}

func TestHost_GetHostConsensusState(t *testing.T) {
	_, _, _, _, ibcMod := prepareEnvironment(t, 1, 0, 0, 0)
	cm := ibcMod.GetBus().GetClientManager()

	consState, err := cm.GetHostConsensusState(&client_types.Height{RevisionNumber: 1, RevisionHeight: 0})
	require.NoError(t, err)

	require.Equal(t, "08-wasm", consState.ClientType())
	require.NoError(t, consState.ValidateBasic())
	require.Less(t, consState.GetTimestamp(), uint64(time.Now().UnixNano()))

	pocketConState := new(light_client_types.PocketConsensusState)
	err = codec.GetCodec().Unmarshal(consState.GetData(), pocketConState)
	require.NoError(t, err)

	blockstore := ibcMod.GetBus().GetPersistenceModule().GetBlockStore()
	block, err := blockstore.GetBlock(0)
	require.NoError(t, err)

	require.Equal(t, block.BlockHeader.Timestamp, pocketConState.Timestamp)
	require.Equal(t, block.BlockHeader.StateHash, pocketConState.StateHash)
	require.Equal(t, block.BlockHeader.StateTreeHashes, pocketConState.StateTreeHashes)
	require.Equal(t, block.BlockHeader.NextValSetHash, pocketConState.NextValSetHash)
}
