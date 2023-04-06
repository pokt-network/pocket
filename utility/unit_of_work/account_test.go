package unit_of_work

import (
	"encoding/hex"
	"math/big"
	"sort"
	"testing"

	"github.com/pokt-network/pocket/runtime/test_artifacts/keygen"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/utils"
	"github.com/stretchr/testify/require"
)

func TestUtilityUnitOfWork_AddAccountAmount(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	acc := getFirstTestingAccount(t, uow)

	initialAmount, err := utils.StringToBigInt(acc.GetAmount())
	require.NoError(t, err)

	addAmount := big.NewInt(1)
	addrBz, err := hex.DecodeString(acc.GetAddress())
	require.NoError(t, err)
	require.NoError(t, uow.addAccountAmount(addrBz, addAmount), "add account amount")

	afterAmount, err := uow.getAccountAmount(addrBz)
	require.NoError(t, err)

	expected := initialAmount.Add(initialAmount, addAmount)
	require.Equal(t, expected, afterAmount)
}

func TestUtilityUnitOfWork_SubtractAccountAmount(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	acc := getFirstTestingAccount(t, uow)

	beforeAmount, err := utils.StringToBigInt(acc.GetAmount())
	require.NoError(t, err)

	subAmount := big.NewInt(100)
	addrBz, er := hex.DecodeString(acc.GetAddress())
	require.NoError(t, er)
	require.NoError(t, uow.subtractAccountAmount(addrBz, subAmount), "sub account amount")

	amount, err := uow.getAccountAmount(addrBz)
	require.NoError(t, err)
	require.NotEqual(t, beforeAmount, amount)

	expected := beforeAmount.Sub(beforeAmount, subAmount)
	require.Equal(t, expected, amount)
}

func TestUtilityUnitOfWork_SetAccountAmount(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)

	addr, err := crypto.GenerateAddress()
	require.NoError(t, err)

	amount := big.NewInt(100)
	require.NoError(t, uow.setAccountAmount(addr, amount), "set account amount")

	gotAmount, err := uow.getAccountAmount(addr)
	require.NoError(t, err)
	require.Equal(t, amount, gotAmount)
}

func TestUtilityUnitOfWork_AddPoolAmount(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	pool := getFirstTestingPool(t, uow)
	addrBz, err := hex.DecodeString(pool.Address)
	require.NoError(t, err)

	initialAmount, err := utils.StringToBigInt(pool.GetAmount())
	require.NoError(t, err)

	addAmount := big.NewInt(1)
	require.NoError(t, uow.addPoolAmount(addrBz, addAmount), "add pool amount")

	afterAmount, err := uow.getPoolAmount(addrBz)
	require.NoError(t, err)

	expected := initialAmount.Add(initialAmount, addAmount)
	require.Equal(t, expected, afterAmount)
}

func TestUtilityUnitOfWork_InsertPool(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)

	_, _, poolAddr := keygen.GetInstance().Next()
	addrBz, err := hex.DecodeString(poolAddr)
	require.NoError(t, err)

	amount := big.NewInt(1000)
	err = uow.insertPool(addrBz, amount)
	require.NoError(t, err, "insert pool")

	poolAmount, err := uow.getPoolAmount(addrBz)
	require.NoError(t, err)
	require.Equal(t, amount, poolAmount)
}

func TestUtilityUnitOfWork_SetPoolAmount(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	pool := getFirstTestingPool(t, uow)
	addrBz, err := hex.DecodeString(pool.GetAddress())
	require.NoError(t, err)

	beforeAmount, err := utils.StringToBigInt(pool.GetAmount())
	require.NoError(t, err)

	expectedAfterAmount := big.NewInt(100)
	require.NoError(t, uow.setPoolAmount(addrBz, expectedAfterAmount), "set pool amount")

	amount, err := uow.getPoolAmount(addrBz)
	require.NoError(t, err)
	require.NotEqual(t, beforeAmount, amount)
	require.Equal(t, amount, expectedAfterAmount)
}

func TestUtilityUnitOfWork_SubPoolAmount(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	pool := getFirstTestingPool(t, uow)
	addrBz, err := hex.DecodeString(pool.GetAddress())
	require.NoError(t, err)

	beforeAmountBig := big.NewInt(1000000000000000)
	require.NoError(t, uow.setPoolAmount(addrBz, beforeAmountBig))

	subAmount := big.NewInt(100)
	require.NoError(t, uow.subPoolAmount(addrBz, subAmount), "sub pool amount")

	amount, err := uow.getPoolAmount(addrBz)
	require.NoError(t, err)
	require.NotEqual(t, beforeAmountBig, amount)

	expected := beforeAmountBig.Sub(beforeAmountBig, subAmount)
	require.Equal(t, expected, amount)
}

func getAllTestingAccounts(t *testing.T, uow *baseUtilityUnitOfWork) []*coreTypes.Account {
	accs, err := uow.persistenceReadContext.GetAllAccounts(0)
	require.NoError(t, err)

	sort.Slice(accs, func(i, j int) bool {
		return accs[i].GetAddress() < accs[j].GetAddress()
	})
	return accs
}

func getFirstTestingAccount(t *testing.T, uow *baseUtilityUnitOfWork) *coreTypes.Account {
	return getAllTestingAccounts(t, uow)[0]
}

func getAllTestingPools(t *testing.T, uow *baseUtilityUnitOfWork) []*coreTypes.Account {
	pools, err := uow.persistenceReadContext.GetAllPools(0)
	require.NoError(t, err)

	sort.Slice(pools, func(i, j int) bool {
		return pools[i].GetAddress() < pools[j].GetAddress()
	})
	return pools
}

func getFirstTestingPool(t *testing.T, uow *baseUtilityUnitOfWork) *coreTypes.Account {
	return getAllTestingPools(t, uow)[0]
}
