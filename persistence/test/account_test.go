package test

import (
	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types"
	typesGenesis "github.com/pokt-network/pocket/shared/types/genesis"
	"math/big"
	"testing"
)

func TestSetAccountAmount(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	acc := NewTestAccount()
	if err := db.SetAccountAmount(acc.Address, DefaultStake); err != nil {
		t.Fatal(err)
	}
	am, err := db.GetAccountAmount(acc.Address)
	if err != nil {
		t.Fatal(err)
	}
	if DefaultStake != am {
		t.Fatal("unexpected amount")
	}
	if err := db.SetAccountAmount(acc.Address, StakeToUpdate); err != nil {
		t.Fatal(err)
	}
	am, err = db.GetAccountAmount(acc.Address)
	if err != nil {
		t.Fatal(err)
	}
	if StakeToUpdate != am {
		t.Fatal("unexpected amount after second set")
	}
}

func TestAddAccountAmount(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	acc := NewTestAccount()
	if err := db.SetAccountAmount(acc.Address, DefaultStake); err != nil {
		t.Fatal(err)
	}
	amountToAddBig := big.NewInt(100)
	if err := db.AddAccountAmount(acc.Address, types.BigIntToString(amountToAddBig)); err != nil {
		t.Fatal(err)
	}
	am, err := db.GetAccountAmount(acc.Address)
	if err != nil {
		t.Fatal(err)
	}
	resultBig := (&big.Int{}).Add(DefaultStakeBig, amountToAddBig)
	expectedResult := types.BigIntToString(resultBig)
	if expectedResult != am {
		t.Fatal("unexpected amount after add")
	}
}

func TestSubAccountAmount(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	acc := NewTestAccount()
	if err := db.SetAccountAmount(acc.Address, DefaultStake); err != nil {
		t.Fatal(err)
	}
	amountToAddBig := big.NewInt(100)
	if err := db.SubtractAccountAmount(acc.Address, types.BigIntToString(amountToAddBig)); err != nil {
		t.Fatal(err)
	}
	am, err := db.GetAccountAmount(acc.Address)
	if err != nil {
		t.Fatal(err)
	}
	resultBig := (&big.Int{}).Sub(DefaultStakeBig, amountToAddBig)
	expectedResult := types.BigIntToString(resultBig)
	if expectedResult != am {
		t.Fatal("unexpected amount after sub")
	}
}

func TestSetPoolAmount(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	pool := NewTestPool()
	if err := db.SetPoolAmount(pool.Name, DefaultStake); err != nil {
		t.Fatal(err)
	}
	am, err := db.GetPoolAmount(pool.Name)
	if err != nil {
		t.Fatal(err)
	}
	if DefaultStake != am {
		t.Fatal("unexpected amount")
	}
	if err := db.SetPoolAmount(pool.Name, StakeToUpdate); err != nil {
		t.Fatal(err)
	}
	am, err = db.GetPoolAmount(pool.Name)
	if err != nil {
		t.Fatal(err)
	}
	if StakeToUpdate != am {
		t.Fatal("unexpected amount after second set")
	}
}

func TestAddPoolAmount(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	pool := NewTestPool()
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
	pool := NewTestPool()
	if err := db.SetPoolAmount(pool.Name, DefaultStake); err != nil {
		t.Fatal(err)
	}
	amountToAddBig := big.NewInt(100)
	if err := db.SubtractPoolAmount(pool.Name, types.BigIntToString(amountToAddBig)); err != nil {
		t.Fatal(err)
	}
	am, err := db.GetPoolAmount(pool.Name)
	if err != nil {
		t.Fatal(err)
	}
	resultBig := (&big.Int{}).Sub(DefaultStakeBig, amountToAddBig)
	expectedResult := types.BigIntToString(resultBig)
	if expectedResult != am {
		t.Fatal("unexpected amount after sub")
	}
}

func NewTestAccount() typesGenesis.Account {
	addr1, _ := crypto.GenerateAddress()
	defaultAmount := types.BigIntToString(big.NewInt(1000000))
	return typesGenesis.Account{
		Address: addr1,
		Amount:  defaultAmount,
	}
}

func NewTestPool() typesGenesis.Pool {
	addr1, _ := crypto.GenerateAddress()
	defaultAmount := types.BigIntToString(big.NewInt(1000000))
	return typesGenesis.Pool{
		Name: DefaultPoolName,
		Account: &typesGenesis.Account{
			Address: addr1, // TODO (Andrew) deprecate address in pool
			Amount:  defaultAmount,
		},
	}
}
