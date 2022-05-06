package test

import (
	"math/big"
	"testing"

	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types"
	typesGenesis "github.com/pokt-network/pocket/shared/types/genesis"
	"github.com/stretchr/testify/require"
)

var (
	defaultAmount = types.BigIntToString(big.NewInt(1000000))
)

func TestSetAccountAmount(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	acc := NewTestAccount(t)

	err := db.SetAccountAmount(acc.Address, DefaultStake)
	require.NoError(t, err)

	am, err := db.GetAccountAmount(acc.Address)
	require.NoError(t, err)
	require.Equal(t, DefaultStake, am, "unexpected amount")

	db.SetAccountAmount(acc.Address, StakeToUpdate)
	require.NoError(t, err)

	am, err = db.GetAccountAmount(acc.Address)
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

	// am, err := db.GetAccountAmount(acc.Address)
	// require.NoError(t, err)

	// resultBig := (&big.Int{}).Add(DefaultStakeBig, amountToAddBig)
	// expectedResult := types.BigIntToString(resultBig)
	// require.Equal(t, expectedResult, am, "unexpected amount after add")
}

func TestSubAccountAmount(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	acc := NewTestAccount(t)

	err := db.SetAccountAmount(acc.Address, DefaultStake)
	require.NoError(t, err)

	amountToAddBig := big.NewInt(100)
	db.SubtractAccountAmount(acc.Address, types.BigIntToString(amountToAddBig))
	require.NoError(t, err)

	am, err := db.GetAccountAmount(acc.Address)
	require.NoError(t, err)

	resultBig := (&big.Int{}).Sub(DefaultStakeBig, amountToAddBig)
	expectedResult := types.BigIntToString(resultBig)
	require.Equal(t, expectedResult, am, "unexpected amount after sub")
}

func TestSetPoolAmount(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	pool := NewTestPool(t)

	err := db.SetPoolAmount(pool.Name, DefaultStake)
	require.NoError(t, err)

	am, err := db.GetPoolAmount(pool.Name)
	require.NoError(t, err)
	require.Equal(t, DefaultStake, am, "unexpected amount")

	err = db.SetPoolAmount(pool.Name, StakeToUpdate)
	require.NoError(t, err)

	am, err = db.GetPoolAmount(pool.Name)
	require.NoError(t, err)
	require.Equal(t, StakeToUpdate, am, "unexpected amount after second set")
}

func TestAddPoolAmount(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	pool := NewTestPool(t)
	if err := db.SetPoolAmount(pool.Name, DefaultStake); err != nil {
		t.Fatal(err)
	}
	amountToAddBig := big.NewInt(100)
	if err := db.AddPoolAmount(pool.Name, types.BigIntToString(amountToAddBig)); err != nil {
		t.Fatal(err)
	}
	am, err := db.GetPoolAmount(pool.Name)
	if err != nil {
		t.Fatal(err)
	}
	resultBig := (&big.Int{}).Add(DefaultStakeBig, amountToAddBig)
	expectedResult := types.BigIntToString(resultBig)
	if expectedResult != am {
		t.Fatal("unexpected amount after add")
	}
}

func TestSubPoolAmount(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	pool := NewTestPool(t)

	err := db.SetPoolAmount(pool.Name, DefaultStake)
	require.NoError(t, err)

	amountToAddBig := big.NewInt(100)
	err = db.SubtractPoolAmount(pool.Name, types.BigIntToString(amountToAddBig))
	require.NoError(t, err)

	am, err := db.GetPoolAmount(pool.Name)
	require.NoError(t, err)

	resultBig := (&big.Int{}).Sub(DefaultStakeBig, amountToAddBig)
	expectedResult := types.BigIntToString(resultBig)
	require.Equal(t, expectedResult, am, "unexpected amount after sub")
}

func NewTestAccount(t *testing.T) typesGenesis.Account {
	addr, err := crypto.GenerateAddress()
	require.NoError(t, err)

	return typesGenesis.Account{
		Address: addr,
		Amount:  defaultAmount,
	}
}

func NewTestPool(t *testing.T) typesGenesis.Pool {
	addr, err := crypto.GenerateAddress()
	require.NoError(t, err)

	return typesGenesis.Pool{
		Name: DefaultPoolName,
		Account: &typesGenesis.Account{
			Address: addr, // TODO(Andrew): deprecate address in pool
			Amount:  defaultAmount,
		},
	}
}
