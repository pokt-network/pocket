package test

import (
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"testing"

	"github.com/pokt-network/pocket/persistence/types"
	"github.com/pokt-network/pocket/shared/modules"

	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/stretchr/testify/require"
)

// TODO(andrew): Find all places where we import twice and update the imports appropriately.

func FuzzAccountAmount(f *testing.F) {
	db := NewTestPostgresContext(f, 0)
	operations := []string{
		"AddAmount",
		"SubAmount",
		"SetAmount",

		"IncrementHeight",
	}
	numOperationTypes := len(operations)

	account := newTestAccount(nil)
	addrBz, err := hex.DecodeString(account.Address)
	// TODO(andrew): All `log.Fatal` calls should be converted to `require.NoError` calls.
	if err != nil {
		log.Fatal(err)
	}
	db.SetAccountAmount(addrBz, DefaultAccountAmount)
	expectedAmount := big.NewInt(DefaultAccountBig.Int64())

	numDbOperations := 20
	for i := 0; i < numDbOperations; i++ {
		f.Add(operations[rand.Intn(numOperationTypes)])
	}

	f.Fuzz(func(t *testing.T, op string) {
		delta := big.NewInt(int64(rand.Intn(1000)))
		deltaString := types.BigIntToString(delta)

		switch op {
		case "AddAmount":
			originalAmountBig, err := db.GetAccountAmount(addrBz, db.Height)
			require.NoError(t, err)

			originalAmount, err := types.StringToBigInt(originalAmountBig)
			require.NoError(t, err)

			err = db.AddAccountAmount(addrBz, deltaString)
			require.NoError(t, err)

			expectedAmount.Add(originalAmount, delta)
		case "SubAmount":
			originalAmountBig, err := db.GetAccountAmount(addrBz, db.Height)
			require.NoError(t, err)

			originalAmount, err := types.StringToBigInt(originalAmountBig)
			require.NoError(t, err)

			err = db.SubtractAccountAmount(addrBz, deltaString)
			require.NoError(t, err)

			expectedAmount.Sub(originalAmount, delta)
		case "SetAmount":
			err := db.SetAccountAmount(addrBz, deltaString)
			require.NoError(t, err)

			expectedAmount = delta
		case "IncrementHeight":
			db.Height++
		default:
			t.Errorf("Unexpected operation fuzzing operation %s", op)
		}

		currentAmount, err := db.GetAccountAmount(addrBz, db.Height)
		require.NoError(t, err)
		require.Equal(t, types.BigIntToString(expectedAmount), currentAmount, fmt.Sprintf("unexpected amount after %s", op))
	})
}

func TestDefaultNonExistentAccountAmount(t *testing.T) {
	db := NewTestPostgresContext(t, 0)
	addr, err := crypto.GenerateAddress()
	require.NoError(t, err)

	accountAmount, err := db.GetAccountAmount(addr, db.Height)
	require.NoError(t, err)
	require.Equal(t, "0", accountAmount)
}

func TestSetAccountAmount(t *testing.T) {
	db := NewTestPostgresContext(t, 0)
	account := newTestAccount(t)
	addrBz, err := hex.DecodeString(account.Address)
	require.NoError(t, err)

	err = db.SetAccountAmount(addrBz, DefaultStake)
	require.NoError(t, err)

	accountAmount, err := db.GetAccountAmount(addrBz, db.Height)
	require.NoError(t, err)
	require.Equal(t, DefaultStake, accountAmount, "unexpected amount")

	err = db.SetAccountAmount(addrBz, StakeToUpdate)
	require.NoError(t, err)

	accountAmount, err = db.GetAccountAmount(addrBz, db.Height)
	require.NoError(t, err)
	require.Equal(t, StakeToUpdate, accountAmount, "unexpected amount after second set")
}

func TestAddAccountAmount(t *testing.T) {
	db := NewTestPostgresContext(t, 0)
	account := newTestAccount(t)

	addrBz, err := hex.DecodeString(account.Address)
	require.NoError(t, err)

	err = db.SetAccountAmount(addrBz, DefaultStake)
	require.NoError(t, err)

	amountToAddBig := big.NewInt(100)
	err = db.AddAccountAmount(addrBz, types.BigIntToString(amountToAddBig))
	require.NoError(t, err)

	accountAmount, err := db.GetAccountAmount(addrBz, db.Height)
	require.NoError(t, err)

	accountAmountBig := (&big.Int{}).Add(DefaultStakeBig, amountToAddBig)
	expectedAccountAmount := types.BigIntToString(accountAmountBig)

	require.Equal(t, expectedAccountAmount, accountAmount, "unexpected amount after add")
}

func TestSubAccountAmount(t *testing.T) {
	db := NewTestPostgresContext(t, 0)
	account := newTestAccount(t)

	addrBz, err := hex.DecodeString(account.Address)
	require.NoError(t, err)

	err = db.SetAccountAmount(addrBz, DefaultStake)
	require.NoError(t, err)

	amountToSubBig := big.NewInt(100)
	err = db.SubtractAccountAmount(addrBz, types.BigIntToString(amountToSubBig))
	require.NoError(t, err)

	accountAmount, err := db.GetAccountAmount(addrBz, db.Height)
	require.NoError(t, err)

	accountAmountBig := (&big.Int{}).Sub(DefaultStakeBig, amountToSubBig)
	expectedAccountAmount := types.BigIntToString(accountAmountBig)
	require.Equal(t, expectedAccountAmount, accountAmount, "unexpected amount after sub")
}

func FuzzPoolAmount(f *testing.F) {
	db := NewTestPostgresContext(f, 0)
	operations := []string{
		"AddAmount",
		"SubAmount",
		"SetAmount",

		"IncrementHeight",
	}
	numOperationTypes := len(operations)

	pool := newTestPool(nil)
	db.SetPoolAmount(pool.Address, DefaultAccountAmount)
	expectedAmount := big.NewInt(DefaultAccountBig.Int64())

	numDbOperations := 20
	for i := 0; i < numDbOperations; i++ {
		f.Add(operations[rand.Intn(numOperationTypes)])
	}

	f.Fuzz(func(t *testing.T, op string) {
		delta := big.NewInt(int64(rand.Intn(1000)))
		deltaString := types.BigIntToString(delta)

		switch op {
		case "AddAmount":
			originalAmountBig, err := db.GetPoolAmount(pool.Address, db.Height)
			require.NoError(t, err)

			originalAmount, err := types.StringToBigInt(originalAmountBig)
			require.NoError(t, err)

			err = db.AddPoolAmount(pool.Address, deltaString)
			require.NoError(t, err)

			expectedAmount.Add(originalAmount, delta)
		case "SubAmount":
			originalAmountBig, err := db.GetPoolAmount(pool.Address, db.Height)
			require.NoError(t, err)

			originalAmount, err := types.StringToBigInt(originalAmountBig)
			require.NoError(t, err)

			err = db.SubtractPoolAmount(pool.Address, deltaString)
			require.NoError(t, err)

			expectedAmount.Sub(originalAmount, delta)
		case "SetAmount":
			err := db.SetPoolAmount(pool.Address, deltaString)
			require.NoError(t, err)

			expectedAmount = delta
		case "IncrementHeight":
			db.Height++
		default:
			t.Errorf("Unexpected operation fuzzing operation %s", op)
		}

		currentAmount, err := db.GetPoolAmount(pool.Address, db.Height)
		require.NoError(t, err)
		require.Equal(t, types.BigIntToString(expectedAmount), currentAmount, fmt.Sprintf("unexpected amount after %s", op))
	})
}

func TestDefaultNonExistentPoolAmount(t *testing.T) {
	db := NewTestPostgresContext(t, 0)

	poolAmount, err := db.GetPoolAmount("some_pool_name", db.Height)
	require.NoError(t, err)
	require.Equal(t, "0", poolAmount)
}

func TestSetPoolAmount(t *testing.T) {
	db := NewTestPostgresContext(t, 0)
	pool := newTestPool(t)

	err := db.SetPoolAmount(pool.Address, DefaultStake)
	require.NoError(t, err)

	poolAmount, err := db.GetPoolAmount(pool.Address, db.Height)
	require.NoError(t, err)
	require.Equal(t, DefaultStake, poolAmount, "unexpected amount")

	err = db.SetPoolAmount(pool.Address, StakeToUpdate)
	require.NoError(t, err)

	poolAmount, err = db.GetPoolAmount(pool.Address, db.Height)
	require.NoError(t, err)
	require.Equal(t, StakeToUpdate, poolAmount, "unexpected amount after second set")
}

func TestAddPoolAmount(t *testing.T) {
	db := NewTestPostgresContext(t, 0)
	pool := newTestPool(t)

	err := db.SetPoolAmount(pool.Address, DefaultStake)
	require.NoError(t, err)

	amountToAddBig := big.NewInt(100)
	err = db.AddPoolAmount(pool.Address, types.BigIntToString(amountToAddBig))
	require.NoError(t, err)

	poolAmount, err := db.GetPoolAmount(pool.Address, db.Height)
	require.NoError(t, err)

	poolAmountBig := (&big.Int{}).Add(DefaultStakeBig, amountToAddBig)
	expectedPoolAmount := types.BigIntToString(poolAmountBig)

	require.Equal(t, expectedPoolAmount, poolAmount, "unexpected amount after add")
}

func TestSubPoolAmount(t *testing.T) {
	db := NewTestPostgresContext(t, 0)
	pool := newTestPool(t)
	err := db.SetPoolAmount(pool.Address, DefaultStake)
	require.NoError(t, err)

	amountToSubBig := big.NewInt(100)
	err = db.SubtractPoolAmount(pool.Address, types.BigIntToString(amountToSubBig))
	require.NoError(t, err)

	poolAmount, err := db.GetPoolAmount(pool.Address, db.Height)
	require.NoError(t, err)

	poolAmountBig := (&big.Int{}).Sub(DefaultStakeBig, amountToSubBig)
	expectedPoolAmount := types.BigIntToString(poolAmountBig)
	require.Equal(t, expectedPoolAmount, poolAmount, "unexpected amount after sub")
}

func TestGetAllAccounts(t *testing.T) {
	db := NewTestPostgresContext(t, 0)
	updateAccount := func(db *persistence.PostgresContext, acc modules.Account) error {
		if addr, err := hex.DecodeString(acc.GetAddress()); err == nil {
			return nil
		} else {
			return db.AddAccountAmount(addr, "10")
		}
	}

	getAllActorsTest(t, db, db.GetAllAccounts, createAndInsertNewAccount, updateAccount, 8)
}

func TestGetAllPools(t *testing.T) {
	db := NewTestPostgresContext(t, 0)

	updatePool := func(db *persistence.PostgresContext, pool modules.Account) error {
		return db.AddPoolAmount(pool.GetAddress(), "10")
	}

	getAllActorsTest(t, db, db.GetAllPools, createAndInsertNewPool, updatePool, 7)
}

// --- Helpers ---

func createAndInsertNewAccount(db *persistence.PostgresContext) (modules.Account, error) {
	account := newTestAccount(nil)
	addr, err := hex.DecodeString(account.Address)
	if err != nil {
		return nil, err
	}
	return &account, db.SetAccountAmount(addr, DefaultAccountAmount)
}

func createAndInsertNewPool(db *persistence.PostgresContext) (modules.Account, error) {
	pool := newTestPool(nil)
	return &pool, db.SetPoolAmount(pool.Address, DefaultAccountAmount)
}

// TODO(olshansky): consolidate newTestAccount and newTestPool into one function

// Note to the reader: lack of consistency between []byte and string in addresses will be consolidated.
func newTestAccount(t *testing.T) types.Account {
	addr, err := crypto.GenerateAddress()
	if t != nil {
		require.NoError(t, err)
	}
	return types.Account{
		Address: hex.EncodeToString(addr),
		Amount:  DefaultAccountAmount,
	}
}

func newTestPool(t *testing.T) types.Account {
	addr, err := crypto.GenerateAddress()
	if t != nil {
		require.NoError(t, err)
	}
	return types.Account{
		Address: hex.EncodeToString(addr),
		Amount:  DefaultAccountAmount,
	}
}
