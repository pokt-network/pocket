package utility_module

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/pokt-network/pocket/persistence/pre_persistence"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types"
	"github.com/pokt-network/pocket/utility"
)

func TestUtilityContext_AddAccountAmount(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	acc := GetAllTestingAccounts(t, ctx)[0]
	initialAmount, err := types.StringToBigInt(acc.Amount)
	if err != nil {
		t.Fatal(err)
	}
	addAmount := big.NewInt(1)
	if err := ctx.AddAccountAmount(acc.Address, addAmount); err != nil {
		t.Fatal(err)
	}
	afterAmount, err := ctx.GetAccountAmount(acc.Address)
	if err != nil {
		t.Fatal(err)
	}
	expected := initialAmount.Add(initialAmount, addAmount)
	if afterAmount.Cmp(expected) != 0 {
		t.Fatalf("amounts are not equal, expected %v, got %v", initialAmount, afterAmount)
	}
}

func TestUtilityContext_AddAccountAmountString(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	acc := GetAllTestingAccounts(t, ctx)[0]
	initialAmount, err := types.StringToBigInt(acc.Amount)
	if err != nil {
		t.Fatal(err)
	}
	addAmount := big.NewInt(1)
	addAmountString := types.BigIntToString(addAmount)
	if err := ctx.AddAccountAmountString(acc.Address, addAmountString); err != nil {
		t.Fatal(err)
	}
	afterAmount, err := ctx.GetAccountAmount(acc.Address)
	if err != nil {
		t.Fatal(err)
	}
	expected := initialAmount.Add(initialAmount, addAmount)
	if afterAmount.Cmp(expected) != 0 {
		t.Fatalf("amounts are not equal, expected %v, got %v", initialAmount, afterAmount)
	}
}

func TestUtilityContext_AddPoolAmount(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	pool := GetAllTestingPools(t, ctx)[0]
	initialAmount, err := types.StringToBigInt(pool.Account.Amount)
	if err != nil {
		t.Fatal(err)
	}
	addAmount := big.NewInt(1)
	if err := ctx.AddPoolAmount(pool.Name, addAmount); err != nil {
		t.Fatal(err)
	}
	afterAmount, err := ctx.GetPoolAmount(pool.Name)
	if err != nil {
		t.Fatal(err)
	}
	expected := initialAmount.Add(initialAmount, addAmount)
	if afterAmount.Cmp(expected) != 0 {
		t.Fatalf("amounts are not equal, expected %v, got %v", initialAmount, afterAmount)
	}
}

func TestUtilityContext_HandleMessageSend(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	accs := GetAllTestingAccounts(t, ctx)
	sendAmount := big.NewInt(1000000)
	sendAmountString := types.BigIntToString(sendAmount)
	senderBalanceBefore, err := types.StringToBigInt(accs[0].Amount)
	if err != nil {
		t.Fatal(err)
	}
	recipientBalanceBefore, err := types.StringToBigInt(accs[1].Amount)
	if err != nil {
		t.Fatal(err)
	}
	msg := NewTestingSendMessage(t, accs[0].Address, accs[1].Address, sendAmountString)
	if err := ctx.HandleMessageSend(&msg); err != nil {
		t.Fatal(err)
	}
	accs = GetAllTestingAccounts(t, ctx)
	senderBalanceAfter, err := types.StringToBigInt(accs[0].Amount)
	if err != nil {
		t.Fatal(err)
	}
	recipientBalanceAfter, err := types.StringToBigInt(accs[1].Amount)
	if err != nil {
		t.Fatal(err)
	}
	if big.NewInt(0).Sub(senderBalanceBefore, senderBalanceAfter).Cmp(sendAmount) != 0 {
		t.Fatal("unexpected sender balance")
	}
	if big.NewInt(0).Sub(recipientBalanceAfter, recipientBalanceBefore).Cmp(sendAmount) != 0 {
		t.Fatal("unexpected recipient balance")
	}
}

func TestUtilityContext_GetMessageSendSignerCandidates(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	accs := GetAllTestingAccounts(t, ctx)
	sendAmount := big.NewInt(1000000)
	sendAmountString := types.BigIntToString(sendAmount)
	msg := NewTestingSendMessage(t, accs[0].Address, accs[1].Address, sendAmountString)
	candidates, err := ctx.GetMessageSendSignerCandidates(&msg)
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) != 1 {
		t.Fatalf("wrong number of candidates, expected %d, got %d", 1, len(candidates))
	}
	if !bytes.Equal(candidates[0], accs[0].Address) {
		t.Fatal("unexpected signer candidate")
	}
}

func TestUtilityContext_InsertPool(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	testPoolName := "TEST_POOL"
	addr, _ := crypto.GenerateAddress()
	amount := types.BigIntToString(big.NewInt(1000))
	if err := ctx.InsertPool(testPoolName, addr, amount); err != nil {
		t.Fatal(err)
	}
	gotAmount, err := ctx.GetPoolAmount(testPoolName)
	if err != nil {
		t.Fatal(err)
	}
	gotAmountString := types.BigIntToString(gotAmount)
	if amount != gotAmountString {
		t.Fatalf("unexpected amount, expected %s got %s", amount, gotAmountString)
	}
}

func TestUtilityContext_SetAccountAmount(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	addr, _ := crypto.GenerateAddress()
	amount := big.NewInt(100)
	if err := ctx.SetAccountAmount(addr, amount); err != nil {
		t.Fatal(err)
	}
	gotAmount, err := ctx.GetAccountAmount(addr)
	if err != nil {
		t.Fatal(err)
	}
	if gotAmount.Cmp(amount) != 0 {
		t.Fatalf("unexpected amounts: expected %v, got %v", amount, gotAmount)
	}
}

func TestUtilityContext_SetAccountWithAmountString(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	addr, _ := crypto.GenerateAddress()
	amount := big.NewInt(100)
	amountString := types.BigIntToString(amount)
	if err := ctx.SetAccountWithAmountString(addr, amountString); err != nil {
		t.Fatal(err)
	}
	gotAmount, err := ctx.GetAccountAmount(addr)
	if err != nil {
		t.Fatal(err)
	}
	if gotAmount.Cmp(amount) != 0 {
		t.Fatalf("unexpected amounts: expected %v, got %v", amount, gotAmount)
	}
}

func TestUtilityContext_SetPoolAmount(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	pool := GetAllTestingPools(t, ctx)[0]
	beforeAmount := pool.Account.Amount
	beforeAmountBig, err := types.StringToBigInt(beforeAmount)
	if err != nil {
		t.Fatal(err)
	}
	expectedAfterAmount := big.NewInt(100)
	if err := ctx.SetPoolAmount(pool.Name, expectedAfterAmount); err != nil {
		t.Fatal(err)
	}
	amount, err := ctx.GetPoolAmount(pool.Name)
	if err != nil {
		t.Fatal(err)
	}
	if beforeAmountBig.Cmp(amount) == 0 {
		t.Fatal("no amount change in pool")
	}
	if expectedAfterAmount.Cmp(amount) != 0 {
		t.Fatalf("unexpected pool amount; expected %v got %v", expectedAfterAmount, amount)
	}
}

func TestUtilityContext_SubPoolAmount(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	pool := GetAllTestingPools(t, ctx)[0]
	beforeAmountBig := big.NewInt(1000000000000000)
	ctx.SetPoolAmount(pool.Name, beforeAmountBig)
	subAmountBig := big.NewInt(100)
	subAmount := types.BigIntToString(subAmountBig)
	if err := ctx.SubPoolAmount(pool.Name, subAmount); err != nil {
		t.Fatal(err)
	}
	amount, err := ctx.GetPoolAmount(pool.Name)
	if err != nil {
		t.Fatal(err)
	}
	if beforeAmountBig.Cmp(amount) == 0 {
		t.Fatal("no amount change in pool")
	}
	expected := beforeAmountBig.Sub(beforeAmountBig, subAmountBig)
	if expected.Cmp(amount) != 0 {
		t.Fatalf("unexpected pool amount; expected %v got %v", expected, amount)
	}
}

func TestUtilityContext_SubtractAccountAmount(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	acc := GetAllTestingAccounts(t, ctx)[0]
	beforeAmount := acc.Amount
	beforeAmountBig, err := types.StringToBigInt(beforeAmount)
	if err != nil {
		t.Fatal(err)
	}
	subAmountBig := big.NewInt(100)
	if err := ctx.SubtractAccountAmount(acc.Address, subAmountBig); err != nil {
		t.Fatal(err)
	}
	amount, err := ctx.GetAccountAmount(acc.Address)
	if err != nil {
		t.Fatal(err)
	}
	if beforeAmountBig.Cmp(amount) == 0 {
		t.Fatal("no amount change in pool")
	}
	expected := beforeAmountBig.Sub(beforeAmountBig, subAmountBig)
	if expected.Cmp(amount) != 0 {
		t.Fatalf("unexpected acc amount; expected %v got %v", expected, amount)
	}
}

func GetAllTestingAccounts(t *testing.T, ctx utility.UtilityContext) []*pre_persistence.Account {
	accs, err := (ctx.Context.PersistenceContext).(*pre_persistence.PrePersistenceContext).GetAllAccounts(0)
	if err != nil {
		t.Fatal(err)
	}
	return accs
}

func GetAllTestingPools(t *testing.T, ctx utility.UtilityContext) []*pre_persistence.Pool {
	accs, err := (ctx.Context.PersistenceContext).(*pre_persistence.PrePersistenceContext).GetAllPools(0)
	if err != nil {
		t.Fatal(err)
	}
	return accs
}
