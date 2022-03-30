package pre_persistence

import (
	"math/big"
	"testing"
)

const (
	testPoolName  = "TEST_POOL"
	testPoolName2 = "TEST_POOL2"
)

// NOTE: Pools encapsulate `accounts` so the functionality is tested
func TestAddPoolAmount(t *testing.T) {
	ctx := NewTestingPrePersistenceContext(t)
	initialBalanceBig := big.NewInt(0)
	initialBalance := BigIntToString(initialBalanceBig)
	addedBalanceBig := big.NewInt(1)
	addedBalance := BigIntToString(addedBalanceBig)
	expectedBalanceBig := initialBalanceBig.Add(initialBalanceBig, addedBalanceBig)
	expectedBalance := BigIntToString(expectedBalanceBig)
	if err := ctx.InsertPool(testPoolName, nil, initialBalance); err != nil {
		t.Fatal(err)
	}
	if err := ctx.AddPoolAmount(testPoolName, addedBalance); err != nil {
		t.Fatal(err)
	}
	actualBalance, err := ctx.GetPoolAmount(testPoolName)
	if err != nil {
		t.Fatal(err)
	}
	if actualBalance != expectedBalance {
		t.Fatalf("not equal balances, expected: %s got %s", expectedBalance, actualBalance)
	}
}

func TestSubtractPoolAmount(t *testing.T) {
	ctx := NewTestingPrePersistenceContext(t)
	initialBalanceBig := big.NewInt(2)
	initialBalance := BigIntToString(initialBalanceBig)
	subBalanceBig := big.NewInt(1)
	subBalance := BigIntToString(subBalanceBig)
	expectedBalanceBig := initialBalanceBig.Sub(initialBalanceBig, subBalanceBig)
	expectedBalance := BigIntToString(expectedBalanceBig)
	if err := ctx.InsertPool(testPoolName, nil, initialBalance); err != nil {
		t.Fatal(err)
	}
	if err := ctx.SubtractPoolAmount(testPoolName, subBalance); err != nil {
		t.Fatal(err)
	}
	actualBalance, err := ctx.GetPoolAmount(testPoolName)
	if err != nil {
		t.Fatal(err)
	}
	if actualBalance != expectedBalance {
		t.Fatalf("not equal balances, expected: %s got %s", expectedBalance, actualBalance)
	}
}

func TestSetPoolAmount(t *testing.T) {
	ctx := NewTestingPrePersistenceContext(t)
	initialBalanceBig := big.NewInt(2)
	initialBalance := BigIntToString(initialBalanceBig)
	setBalanceBig := big.NewInt(1)
	setBalance := BigIntToString(setBalanceBig)
	if err := ctx.InsertPool(testPoolName, nil, initialBalance); err != nil {
		t.Fatal(err)
	}
	if err := ctx.SetPoolAmount(testPoolName, setBalance); err != nil {
		t.Fatal(err)
	}
	actualBalance, err := ctx.GetPoolAmount(testPoolName)
	if err != nil {
		t.Fatal(err)
	}
	if actualBalance != setBalance {
		t.Fatalf("not equal balances, expected: %s got %s", setBalance, actualBalance)
	}
}

func TestGetAllPoolsAmount(t *testing.T) {
	ctx := NewTestingPrePersistenceContext(t)
	initialBalanceBig := big.NewInt(2)
	initialBalance := BigIntToString(initialBalanceBig)
	if err := ctx.InsertPool(testPoolName, nil, initialBalance); err != nil {
		t.Fatal(err)
	}
	if err := ctx.InsertPool(testPoolName2, nil, initialBalance); err != nil {
		t.Fatal(err)
	}
	pools, err := ctx.(*PrePersistenceContext).GetAllPools(0)
	if err != nil {
		t.Fatal(err)
	}
	got1, got2 := false, false
	for _, pool := range pools {
		if pool.Name == testPoolName {
			got1 = true
		}
		if pool.Name == testPoolName2 {
			got2 = true
		}
	}
	if !got1 || !got2 {
		t.Fatal("not all pools returned")
	}
}
