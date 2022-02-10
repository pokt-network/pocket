package test

import (
	"github.com/pokt-network/utility-pre-prototype/utility/types"
	"github.com/syndtr/goleveldb/leveldb/comparer"
	"github.com/syndtr/goleveldb/leveldb/memdb"
	"math/big"
	"reflect"
	"testing"
)

func TestGenesisInit(t *testing.T) {
	utilityModule := NewMockUtilityModule(0, memdb.New(comparer.DefaultComparer, 1000000))
	genesisState, _, _, _, _, err := NewMockGenesisState(5, 1, 1, 5)
	if err != nil {
		t.Fatal(err)
	}
	err = InitGenesis(utilityModule, genesisState)
	if err != nil {
		t.Fatal(err)
	}
	db := utilityModule.Store()
	exported, err := db.PersistenceContext.(*MockPersistenceContext).ExportState()
	if err != nil {
		t.Fatal(err)
	}
	// TODO better validation
	if len(genesisState.Apps) != len(exported.Apps) {
		t.Fatalf("incorrect number of apps: got %d expected %d\n", len(exported.Apps), len(genesisState.Apps))
	}
	if len(genesisState.Validators) != len(exported.Validators) {
		t.Fatalf("incorrect number of vals: got %d expected %d\n", len(exported.Validators), len(genesisState.Validators))
	}
	if len(genesisState.Fishermen) != len(exported.Fishermen) {
		t.Fatalf("incorrect number of fish: got %d expected %d\n", len(exported.Fishermen), len(genesisState.Fishermen))
	}
	if len(genesisState.ServiceNodes) != len(exported.ServiceNodes) {
		t.Fatalf("incorrect number of service nodes: got %d expected %d\n", len(exported.ServiceNodes), len(genesisState.ServiceNodes))
	}
	if len(genesisState.Accounts) != len(exported.Accounts) {
		t.Fatalf("incorrect number of accounts: got %d expected %d\n", len(exported.Accounts), len(genesisState.Accounts))
	}
	if len(genesisState.Pools) != len(exported.Pools) {
		t.Fatalf("incorrect number of pools: got %d expected %d\n", len(exported.Pools), len(genesisState.Pools))
	}
	if !reflect.DeepEqual(genesisState.Params.MessageChangeParameterFee, exported.Params.MessageChangeParameterFee) {
		t.Fatalf("incorrect params: \ngot\n%v \nexpected\n%v\n", exported.Params, genesisState.Params)
	}
}

func TestCheckTx(t *testing.T) {
	cdc := Codec()
	utilityModule := NewMockUtilityModule(0, memdb.New(comparer.DefaultComparer, 1000000))
	// init genesis
	genesisState, validatorKeys, _, _, _, err := NewMockGenesisState(5, 1, 1, 5)
	if err != nil {
		t.Fatal(err)
	}
	err = InitGenesis(utilityModule, genesisState)
	if err != nil {
		t.Fatal(err)
	}
	// build transaction
	privateKeyA := validatorKeys[0]
	privateKeyB := validatorKeys[1]
	addrA := privateKeyA.Address()
	addrB := privateKeyB.Address()
	amount := types.BigIntToString(big.NewInt(1000000))
	fee := types.BigIntToString(big.NewInt(10000))
	tx, err := NewTransaction(&types.MessageSend{
		FromAddress: addrA,
		ToAddress:   addrB,
		Amount:      amount,
	}, fee)
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.Sign(privateKeyA); err != nil {
		t.Fatal(err)
	}
	transactionBytes, err := cdc.Marshal(tx)
	if err != nil {
		t.Fatal(err)
	}
	if err := utilityModule.CheckTransaction(transactionBytes); err != nil {
		t.Fatal(err)
	}
	if utilityModule.Mempool.Size() != 1 {
		t.Fatal("incorrect mempool size")
	}
}

func TestGetTransactionForProposal(t *testing.T) {
	cdc := Codec()
	utilityModule := NewMockUtilityModule(0, memdb.New(comparer.DefaultComparer, 1000000))
	// init genesis
	genesisState, validatorKeys, _, _, _, err := NewMockGenesisState(5, 1, 1, 5)
	if err != nil {
		t.Fatal(err)
	}
	err = InitGenesis(utilityModule, genesisState)
	if err != nil {
		t.Fatal(err)
	}
	// build transaction
	privateKeyA := validatorKeys[0]
	privateKeyB := validatorKeys[1]
	addrA := privateKeyA.Address()
	addrB := privateKeyB.Address()
	amount := types.BigIntToString(big.NewInt(1000000))
	fee := types.BigIntToString(big.NewInt(10000))
	tx, err := NewTransaction(&types.MessageSend{
		FromAddress: addrA,
		ToAddress:   addrB,
		Amount:      amount,
	}, fee)
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.Sign(privateKeyA); err != nil {
		t.Fatal(err)
	}
	transactionBytes, err := cdc.Marshal(tx)
	if err != nil {
		t.Fatal(err)
	}
	err = utilityModule.CheckTransaction(transactionBytes)
	if err != nil {
		t.Fatal(err)
	}
	txs, err := utilityModule.GetTransactionsForProposal(addrA, 1000, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(txs) != 1 {
		t.Fatal("wrong number of transactions")
	}
	if !reflect.DeepEqual(txs[0], transactionBytes) {
		t.Fatal("not the expected transaction...")
	}
}

func TestApplyBlock(t *testing.T) {
	cdc := Codec()
	utilityModule := NewMockUtilityModule(0, memdb.New(comparer.DefaultComparer, 1000000))
	// init genesis
	genesisState, validatorKeys, _, _, _, err := NewMockGenesisState(5, 1, 1, 5)
	if err != nil {
		t.Fatal(err)
	}
	err = InitGenesis(utilityModule, genesisState)
	if err != nil {
		t.Fatal(err)
	}
	feeBig, err := utilityModule.GetMessageSendFee()
	if err != nil {
		t.Fatal(err)
	}
	amountBig := big.NewInt(100000)
	// build transaction
	privateKeyA := validatorKeys[0]
	privateKeyB := validatorKeys[1]
	privateKeyC := validatorKeys[2]
	addrA := privateKeyA.Address()
	addrB := privateKeyB.Address()
	addrC := privateKeyC.Address()
	initialA, err := utilityModule.GetAccountAmount(addrA)
	if err != nil {
		panic(err)
	}
	initialB, err := utilityModule.GetAccountAmount(addrB)
	if err != nil {
		panic(err)
	}
	initialC, err := utilityModule.GetAccountAmount(addrC)
	if err != nil {
		panic(err)
	}
	amount := types.BigIntToString(amountBig)
	fee := types.BigIntToString(feeBig)
	tx, err := NewTransaction(&types.MessageSend{
		FromAddress: addrA,
		ToAddress:   addrB,
		Amount:      amount,
	}, fee)
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.Sign(privateKeyA); err != nil {
		t.Fatal(err)
	}
	transactionBytes, err := cdc.Marshal(tx)
	if err != nil {
		t.Fatal(err)
	}
	var txs [][]byte
	txs = append(txs, transactionBytes)
	hash, err := utilityModule.ApplyBlock(0, addrC, txs, nil)
	if err != nil {
		t.Fatal(err)
	}
	if hash == nil {
		t.Fatal(err)
	}
	afterA, err := utilityModule.GetAccountAmount(addrA)
	if err != nil {
		t.Fatal(err)
	}
	afterB, err := utilityModule.GetAccountAmount(addrB)
	if err != nil {
		t.Fatal(err)
	}
	afterC, err := utilityModule.GetAccountAmount(addrC)
	if err != nil {
		t.Fatal(err)
	}
	withdrawalAmount := new(big.Int).Add(amountBig, feeBig)
	if initialA.Sub(initialA, afterA).Cmp(withdrawalAmount) != 0 {
		t.Fatal("not correct account withdrawal")
	}
	if initialB.Sub(afterB, initialB).Cmp(amountBig) != 0 {
		t.Fatal("not correct account deposit")
	}
	ppf, err := utilityModule.GetProposerPercentageOfFees()
	if err != nil {
		t.Fatal(err)
	}
	proposerAmount := &big.Int{}
	proposerAmount.Mul(feeBig, big.NewInt(int64(ppf)))
	proposerAmount.Quo(proposerAmount, big.NewInt(100))
	if initialC.Sub(afterC, initialC).Cmp(proposerAmount) != 0 {
		t.Fatal("not correct proposer balance")
	}
}

func Codec() types.Codec {
	return types.UtilityCodec()
}
