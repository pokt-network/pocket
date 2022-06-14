package test

import (
	"fmt"
	"math/big"
	"math/rand"
	"testing"

	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types"
	typesGenesis "github.com/pokt-network/pocket/shared/types/genesis"
	"github.com/stretchr/testify/require"
)

// --- Account Tests ---

func FuzzAccountAmount(f *testing.F) {
	// Setup
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}

	// TODO(team): We can create a map of functions once the following is answered:
	// https://stackoverflow.com/questions/72496074/golang-1-18-map-of-functions is answered
	ops := []string{"Get", "Add", "Sub", "Set"}

	// IMPROVE(team): See the documentation in `persistence/README.md` for more information on
	// why we need to call this initialization.
	acc := NewTestAccount(nil)
	db.SetAccountAmount(acc.Address, DefaultAccountAmount)
	expectedAmount := big.NewInt(DefaultAccountBig.Int64())

	numOptions := len(ops)
	numOperations := 20
	for i := 0; i < numOperations; i++ {
		f.Add(ops[rand.Intn(numOptions)])
	}

	// IMPROVE(team): Randomize the amounts
	// IMPROVE(team): Assert negative balances never happen
	f.Fuzz(func(t *testing.T, op string) {
		switch op {
		case "Get":
			amount, err := db.GetAccountAmount(acc.Address, db.Height)
			require.NoError(t, err)
			require.Equal(t, types.BigIntToString(expectedAmount), amount, "unexpected retrieved amount")
		case "Add":
			err := db.AddAccountAmount(acc.Address, DefaultDeltaAmount)
			require.NoError(t, err)
			expectedAmount.Add(expectedAmount, DefaultDeltaBig)
		case "Sub":
			err := db.SubtractAccountAmount(acc.Address, DefaultDeltaAmount)
			require.NoError(t, err)
			expectedAmount.Sub(expectedAmount, DefaultDeltaBig)
		case "Set":
			err := db.SetAccountAmount(acc.Address, DefaultAccountAmount)
			require.NoError(t, err)
			expectedAmount = big.NewInt(DefaultAccountBig.Int64())
		}
		currentAmount, err := db.GetAccountAmount(acc.Address, db.Height)
		require.NoError(t, err)
		require.Equal(t, types.BigIntToString(expectedAmount), currentAmount, fmt.Sprintf("unexpected amount after %s", op))
	})
}

// 1) Fuzz tests - takes all the CRU operations that any actor may have and fuzz's all of them and verifying functionality
// Change height in operations for each fuzz test
// 2) User story testing - create a few examples per function

func TestSetAccountAmount(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	acc := NewTestAccount(t)

	err := db.SetAccountAmount(acc.Address, DefaultStake)
	require.NoError(t, err)

	am, err := db.GetAccountAmount(acc.Address, db.Height)
	require.NoError(t, err)
	require.Equal(t, DefaultStake, am, "unexpected amount")

	db.SetAccountAmount(acc.Address, StakeToUpdate)
	require.NoError(t, err)

	am, err = db.GetAccountAmount(acc.Address, db.Height)
	require.NoError(t, err)
	require.Equal(t, StakeToUpdate, am, "unexpected amount after second set")
}

func TestAddAccountAmount(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	acc := NewTestAccount(t)

	err := db.SetAccountAmount(acc.Address, DefaultStake)
	require.NoError(t, err)

	amountToAddBig := big.NewInt(100)
	err = db.AddAccountAmount(acc.Address, types.BigIntToString(amountToAddBig))
	require.NoError(t, err)

	am, err := db.GetAccountAmount(acc.Address, db.Height)
	require.NoError(t, err)

	resultBig := (&big.Int{}).Add(DefaultStakeBig, amountToAddBig)
	expectedResult := types.BigIntToString(resultBig)
	require.Equal(t, expectedResult, am, "unexpected amount after add")
}

func TestSubAccountAmount(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	acc := NewTestAccount(t)

	err := db.SetAccountAmount(acc.Address, DefaultStake)
	require.NoError(t, err)

	amountToSubBig := big.NewInt(100)
	db.SubtractAccountAmount(acc.Address, types.BigIntToString(amountToSubBig))
	require.NoError(t, err)

	am, err := db.GetAccountAmount(acc.Address, db.Height)
	require.NoError(t, err)

	resultBig := (&big.Int{}).Sub(DefaultStakeBig, amountToSubBig)
	expectedResult := types.BigIntToString(resultBig)
	require.Equal(t, expectedResult, am, "unexpected amount after sub")
}

// --- Pool Tests ---

func FuzzPoolAmount(f *testing.F) {
	// Setup
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}

	// TODO(team): We can create a map of functions once the following is answered:
	// https://stackoverflow.com/questions/72496074/golang-1-18-map-of-functions is answered
	ops := []string{"Get", "Add", "Sub", "Set"}

	// IMPROVE(team): See the documentation in `persistence/README.md` for more information on
	// why we need to call this initialization.
	pool := NewTestPool(nil)
	db.SetPoolAmount(pool.Name, DefaultAccountAmount)
	expectedAmount := big.NewInt(DefaultAccountBig.Int64())

	numOptions := len(ops)
	numOperations := 20
	for i := 0; i < numOperations; i++ {
		f.Add(ops[rand.Intn(numOptions)])
	}

	// IMPROVE(team): Randomize the amounts
	// IMPROVE(team): Assert negative balances never happen
	f.Fuzz(func(t *testing.T, op string) {
		switch op {
		case "Get":
			amount, err := db.GetPoolAmount(pool.Name, db.Height)
			require.NoError(t, err)
			require.Equal(t, types.BigIntToString(expectedAmount), amount, "unexpected retrieved amount")
		case "Add":
			err := db.AddPoolAmount(pool.Name, DefaultDeltaAmount)
			require.NoError(t, err)
			expectedAmount.Add(expectedAmount, DefaultDeltaBig)
		case "Sub":
			err := db.SubtractPoolAmount(pool.Name, DefaultDeltaAmount)
			require.NoError(t, err)
			expectedAmount.Sub(expectedAmount, DefaultDeltaBig)
		case "Set":
			err := db.SetPoolAmount(pool.Name, DefaultAccountAmount)
			require.NoError(t, err)
			expectedAmount = big.NewInt(DefaultAccountBig.Int64())
		}
		currentAmount, err := db.GetPoolAmount(pool.Name, db.Height)
		require.NoError(t, err)
		require.Equal(t, types.BigIntToString(expectedAmount), currentAmount, fmt.Sprintf("unexpected amount after %s", op))
	})
}
func TestSetPoolAmount(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	pool := NewTestPool(t)

	err := db.SetPoolAmount(pool.Name, DefaultStake)
	require.NoError(t, err)

	am, err := db.GetPoolAmount(pool.Name, db.Height)
	require.NoError(t, err)
	require.Equal(t, DefaultStake, am, "unexpected amount")

	err = db.SetPoolAmount(pool.Name, StakeToUpdate)
	require.NoError(t, err)

	am, err = db.GetPoolAmount(pool.Name, db.Height)
	require.NoError(t, err)
	require.Equal(t, StakeToUpdate, am, "unexpected amount after second set")
}

func TestAddPoolAmount(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	pool := NewTestPool(t)

	err := db.SetPoolAmount(pool.Name, DefaultStake)
	require.NoError(t, err)

	amountToAddBig := big.NewInt(100)
	err = db.AddPoolAmount(pool.Name, types.BigIntToString(amountToAddBig))
	require.NoError(t, err)

	am, err := db.GetPoolAmount(pool.Name, db.Height)
	require.NoError(t, err)

	resultBig := (&big.Int{}).Add(DefaultStakeBig, amountToAddBig)
	expectedResult := types.BigIntToString(resultBig)
	require.Equal(t, expectedResult, am, "unexpected amount after add")
}

func TestSubPoolAmount(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	pool := NewTestPool(t)

	err := db.SetPoolAmount(pool.Name, DefaultStake)
	require.NoError(t, err)

	amountToSubBig := big.NewInt(100)
	err = db.SubtractPoolAmount(pool.Name, types.BigIntToString(amountToSubBig))
	require.NoError(t, err)

	am, err := db.GetPoolAmount(pool.Name, db.Height)
	require.NoError(t, err)

	resultBig := (&big.Int{}).Sub(DefaultStakeBig, amountToSubBig)
	expectedResult := types.BigIntToString(resultBig)
	require.Equal(t, expectedResult, am, "unexpected amount after sub")
}

// --- Helpers ---

func NewTestAccount(t *testing.T) typesGenesis.Account {
	addr, err := crypto.GenerateAddress()
	if t != nil {
		require.NoError(t, err)
	}

	return typesGenesis.Account{
		Address: addr,
		Amount:  DefaultAccountAmount,
	}
}

func NewTestPool(t *testing.T) typesGenesis.Pool {
	_, err := crypto.GenerateAddress()
	if t != nil {
		require.NoError(t, err)
	}

	return typesGenesis.Pool{
		Name: DefaultPoolName,
		Account: &typesGenesis.Account{
			Amount: DefaultAccountAmount,
		},
	}
}
