package ibc

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEmitMessage_MessageAddedToLocalMempool(t *testing.T) {
	_, _, utilityMod, _, ibcMod := prepareEnvironment(t, 1, 0, 0, 0)
	ibcHost := ibcMod.GetBus().GetIBCHost()

	// update store
	store, err := ibcHost.GetProvableStore("test")
	require.NoError(t, err)

	err = store.Set([]byte("key"), []byte("value"))
	require.NoError(t, err)

	mempool := utilityMod.GetMempool()
	require.Len(t, mempool.GetAll(), 1)
}
