package test

import (
	"encoding/hex"
	"testing"

	"github.com/pokt-network/pocket/persistence/savepoints"
	"github.com/stretchr/testify/require"
)

func TestSavepoint_GetAccountAmount(t *testing.T) {
	db := NewTestPostgresContext(t, 0)

	sF := savepoints.NewSavepointFactory(db)
	savepoint, err := sF.CreateSavepoint(0)
	require.NoError(t, err)

	accounts, err := savepoint.GetAllAccounts(0)
	require.NoError(t, err)
	require.Equal(t, 8, len(accounts))

	addrBz, err := hex.DecodeString(accounts[0].Address)
	require.NoError(t, err)

	accountAmount, err := savepoint.GetAccountAmount(addrBz, 0)
	require.NoError(t, err)
	require.Equal(t, accounts[0].Amount, accountAmount)
}
