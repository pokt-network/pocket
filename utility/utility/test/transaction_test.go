package test

import (
	"bytes"
	"pocket/utility/utility/types"
	"github.com/syndtr/goleveldb/leveldb/comparer"
	"github.com/syndtr/goleveldb/leveldb/memdb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"math/big"
	"testing"
)

func TestStakeValidator(t *testing.T) {
	numOfValidators := 5
	//cdc := Codec()
	utilityModule := NewMockUtilityModule(0, memdb.New(comparer.DefaultComparer, 1000000))
	// init genesis
	genesisState, valKeys, appsKeys, _, _, err := NewMockGenesisState(numOfValidators, 1, 1, 5)
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
	amountBig := big.NewInt(15000000000)
	// build transaction
	privateKeyA := appsKeys[0]
	privateKeyB := valKeys[0]
	pubKeyA := privateKeyA.PublicKey()
	addrA := pubKeyA.Address()
	tx, err := NewTransaction(&types.MessageStakeValidator{
		PublicKey:     pubKeyA.Bytes(),
		Amount:        types.BigIntToString(amountBig),
		ServiceURL:    defaultServiceURL,
		OutputAddress: addrA,
	}, types.BigIntToString(feeBig))
	if err := tx.Sign(privateKeyA); err != nil {
		t.Fatal(err)
	}
	transactionBytes, err := tx.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if err := utilityModule.CheckTransaction(transactionBytes); err != nil {
		t.Fatal(err)
	}
	if _, err := utilityModule.ApplyBlock(0, privateKeyB.Address(), [][]byte{transactionBytes}, nil); err != nil {
		t.Fatal(err)
	}
	state, err := ExportState(utilityModule)
	if err != nil {
		t.Fatal(err)
	}
	if len(state.Validators) != numOfValidators+1 {
		t.Fatal("wrong validator count")
	}
	if ok, _ := utilityModule.GetValidatorExists(addrA); !ok {
		t.Fatal("newly created validator doesn't exist")
	}
	for _, v := range state.Validators {
		if bytes.Equal(v.Address, addrA) {
			if v.StakedTokens != types.BigIntToString(amountBig) {
				t.Fatalf("wrong staked tokens: got: %v, expected: %v", v.StakedTokens, types.BigIntToString(amountBig))
			}
			if v.Status != defaultStakeStatus {
				t.Fatalf("wrong status; got: %v, expected %v", v.Status, defaultStakeStatus)
			}
			if v.ServiceURL != defaultServiceURL {
				t.Fatalf("wrong service url; got: %v, expected: %v", v.ServiceURL, defaultServiceURL)
			}
		}
	}
	return
}

func TestEditStakeValidator(t *testing.T) {
	//cdc := Codec()
	utilityModule := NewMockUtilityModule(0, memdb.New(comparer.DefaultComparer, 1000000))
	// init genesis
	genesisState, valKeys, _, _, _, err := NewMockGenesisState(5, 1, 1, 5)
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
	amountToAddBig := big.NewInt(1)
	// build transaction
	privateKeyA := valKeys[1]
	privateKeyB := valKeys[0]
	pubKeyA := privateKeyA.PublicKey()
	addrA := pubKeyA.Address()
	tx, err := NewTransaction(&types.MessageEditStakeValidator{
		Address:     addrA,
		AmountToAdd: types.BigIntToString(amountToAddBig),
		ServiceURL:  defaultServiceURLEdited,
	}, types.BigIntToString(feeBig))
	if err := tx.Sign(privateKeyA); err != nil {
		t.Fatal(err)
	}
	transactionBytes, err := tx.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if err := utilityModule.CheckTransaction(transactionBytes); err != nil {
		t.Fatal(err)
	}
	if _, err := utilityModule.ApplyBlock(0, privateKeyB.Address(), [][]byte{transactionBytes}, nil); err != nil {
		t.Fatal(err)
	}
	state, err := ExportState(utilityModule)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range state.Validators {
		if bytes.Equal(v.Address, addrA) {
			amount := &big.Int{}
			amount.Add(defaultStakeBig, amountToAddBig)
			if v.StakedTokens != types.BigIntToString(amount) {
				t.Fatalf("wrong staked tokens: got: %v, expected: %v", v.StakedTokens, types.BigIntToString(amount))
			}
			if v.Status != defaultStakeStatus {
				t.Fatalf("wrong status; got: %v, expected %v", v.Status, defaultStakeStatus)
			}
			if v.ServiceURL != defaultServiceURLEdited {
				t.Fatalf("wrong service url; got: %v, expected: %v", v.ServiceURL, defaultServiceURL)
			}
		}
	}
	return
}

func TestPauseValidator(t *testing.T) {
	//cdc := Codec()
	utilityModule := NewMockUtilityModule(0, memdb.New(comparer.DefaultComparer, 1000000))
	// init genesis
	genesisState, valKeys, _, _, _, err := NewMockGenesisState(5, 1, 1, 5)
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
	// build transaction
	privateKeyA := valKeys[1]
	privateKeyB := valKeys[0]
	pubKeyA := privateKeyA.PublicKey()
	addrA := pubKeyA.Address()
	tx, err := NewTransaction(&types.MessagePauseValidator{
		Address: addrA,
	}, types.BigIntToString(feeBig))
	if err := tx.Sign(privateKeyA); err != nil {
		t.Fatal(err)
	}
	transactionBytes, err := tx.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if err := utilityModule.CheckTransaction(transactionBytes); err != nil {
		t.Fatal(err)
	}
	if _, err := utilityModule.ApplyBlock(0, privateKeyB.Address(), [][]byte{transactionBytes}, nil); err != nil {
		t.Fatal(err)
	}
	state, err := ExportState(utilityModule)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range state.Validators {
		if bytes.Equal(v.Address, addrA) {
			if v.Paused != true {
				t.Fatalf("wrong paused status; got: %v, expected %v", v.Paused, true)
			}
		}
	}
	return
}

func TestUnpauseValidator(t *testing.T) {
	//cdc := Codec()
	utilityModule := NewMockUtilityModule(0, memdb.New(comparer.DefaultComparer, 1000000))
	// init genesis
	genesisState, valKeys, _, _, _, err := NewMockGenesisState(5, 1, 1, 5)
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
	// build transaction
	privateKeyA := valKeys[1]
	privateKeyB := valKeys[0]
	pubKeyA := privateKeyA.PublicKey()
	addrA := pubKeyA.Address()
	pausedBlocks, err := utilityModule.GetValidatorMinimumPauseBlocks()
	if err != nil {
		t.Fatal(err)
	}
	pauseHeight := int64(1)
	utilityModule.LatestHeight = int64(pausedBlocks) + pauseHeight
	if err := utilityModule.SetValidatorPauseHeight(addrA, pauseHeight); err != nil {
		t.Fatal(err)
	}
	tx, err := NewTransaction(&types.MessageUnpauseValidator{
		Address: addrA,
	}, types.BigIntToString(feeBig))
	if err := tx.Sign(privateKeyA); err != nil {
		t.Fatal(err)
	}
	transactionBytes, err := tx.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if err := utilityModule.CheckTransaction(transactionBytes); err != nil {
		t.Fatal(err)
	}
	if _, err := utilityModule.ApplyBlock(utilityModule.LatestHeight, privateKeyB.Address(), [][]byte{transactionBytes}, nil); err != nil {
		t.Fatal(err)
	}
	state, err := ExportState(utilityModule)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range state.Validators {
		if bytes.Equal(v.Address, addrA) {
			if v.Paused != true {
				t.Fatalf("wrong paused status; got: %v, expected %v", v.Paused, false)
			}
		}
	}
	return
}

func TestUnstakeValidator(t *testing.T) {
	numOfValidators := 5
	//cdc := Codec()
	utilityModule := NewMockUtilityModule(0, memdb.New(comparer.DefaultComparer, 1000000))
	// init genesis
	genesisState, valKeys, _, _, _, err := NewMockGenesisState(numOfValidators, 1, 1, 5)
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
	// build transaction
	privateKeyA := valKeys[1]
	privateKeyB := valKeys[0]
	pubKeyA := privateKeyA.PublicKey()
	addrA := pubKeyA.Address()
	tx, err := NewTransaction(&types.MessageUnstakeValidator{
		Address: addrA,
	}, types.BigIntToString(feeBig))
	if err := tx.Sign(privateKeyA); err != nil {
		t.Fatal(err)
	}
	transactionBytes, err := tx.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if err := utilityModule.CheckTransaction(transactionBytes); err != nil {
		t.Fatal(err)
	}
	if _, err := utilityModule.ApplyBlock(0, privateKeyB.Address(), [][]byte{transactionBytes}, nil); err != nil {
		t.Fatal(err)
	}
	state, err := ExportState(utilityModule)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range state.Validators {
		if bytes.Equal(v.Address, addrA) {
			if v.Status != types.UnstakingStatus {
				t.Fatalf("wrong status; got: %v, expected %v", v.Status, defaultStakeStatus)
			}
			blocks, err := utilityModule.GetValidatorUnstakingBlocks()
			if err != nil {
				t.Fatal(err)
			}
			utilityModule.LatestHeight = blocks
			readyToUnstake, err := utilityModule.GetValidatorsReadyToUnstake()
			if err != nil {
				t.Fatal(err)
			}
			if len(readyToUnstake) != 1 {
				t.Fatalf("should have 1 ready to unstake")
			}
		}
	}
	return
}

func TestStakeFisherman(t *testing.T) {
	numOfFisherman := 1
	//cdc := Codec()
	utilityModule := NewMockUtilityModule(0, memdb.New(comparer.DefaultComparer, 1000000))
	// init genesis
	genesisState, valKeys, appsKeys, _, _, err := NewMockGenesisState(5, 1, numOfFisherman, 5)
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
	amountBig := big.NewInt(15000000000)
	// build transaction
	privateKeyA := appsKeys[0]
	privateKeyB := valKeys[0]
	pubKeyA := privateKeyA.PublicKey()
	addrA := pubKeyA.Address()
	tx, err := NewTransaction(&types.MessageStakeFisherman{
		PublicKey:     pubKeyA.Bytes(),
		Chains:        defaultChains,
		Amount:        types.BigIntToString(amountBig),
		ServiceURL:    defaultServiceURL,
		OutputAddress: addrA,
	}, types.BigIntToString(feeBig))
	if err := tx.Sign(privateKeyA); err != nil {
		t.Fatal(err)
	}
	transactionBytes, err := tx.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if err := utilityModule.CheckTransaction(transactionBytes); err != nil {
		t.Fatal(err)
	}
	if _, err := utilityModule.ApplyBlock(0, privateKeyB.Address(), [][]byte{transactionBytes}, nil); err != nil {
		t.Fatal(err)
	}
	state, err := ExportState(utilityModule)
	if err != nil {
		t.Fatal(err)
	}
	if len(state.Fishermen) != numOfFisherman+1 {
		t.Fatal("wrong fish count")
	}
	if ok, _ := utilityModule.GetFishermanExists(addrA); !ok {
		t.Fatal("newly created fish doesn't exist")
	}
	for _, v := range state.Fishermen {
		if bytes.Equal(v.Address, addrA) {
			if v.StakedTokens != types.BigIntToString(amountBig) {
				t.Fatalf("wrong staked tokens: got: %v, expected: %v", v.StakedTokens, types.BigIntToString(amountBig))
			}
			if v.Status != defaultStakeStatus {
				t.Fatalf("wrong status; got: %v, expected %v", v.Status, defaultStakeStatus)
			}
			if !ChainsEquality(v.Chains, defaultChains) {
				t.Fatalf("wrong chains; got: %v, expected: %v", v.Chains, defaultChains)
			}
			if v.ServiceURL != defaultServiceURL {
				t.Fatalf("wrong service url; got: %v, expected: %v", v.ServiceURL, defaultServiceURL)
			}
		}
	}
}

func TestEditStakeFisherman(t *testing.T) {
	//cdc := Codec()
	utilityModule := NewMockUtilityModule(0, memdb.New(comparer.DefaultComparer, 1000000))
	// init genesis
	genesisState, valKeys, _, _, fishKeys, err := NewMockGenesisState(5, 1, 1, 5)
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
	amountToAddBig := big.NewInt(1)
	// build transaction
	privateKeyA := fishKeys[0]
	privateKeyB := valKeys[0]
	pubKeyA := privateKeyA.PublicKey()
	addrA := pubKeyA.Address()
	tx, err := NewTransaction(&types.MessageEditStakeFisherman{
		Address:     addrA,
		Chains:      defaultChainsEdited,
		AmountToAdd: types.BigIntToString(amountToAddBig),
		ServiceURL:  defaultServiceURLEdited,
	}, types.BigIntToString(feeBig))
	if err := tx.Sign(privateKeyA); err != nil {
		t.Fatal(err)
	}
	transactionBytes, err := tx.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if err := utilityModule.CheckTransaction(transactionBytes); err != nil {
		t.Fatal(err)
	}
	if _, err := utilityModule.ApplyBlock(0, privateKeyB.Address(), [][]byte{transactionBytes}, nil); err != nil {
		t.Fatal(err)
	}
	state, err := ExportState(utilityModule)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range state.Fishermen {
		if bytes.Equal(v.Address, addrA) {
			amount := &big.Int{}
			amount.Add(defaultStakeBig, amountToAddBig)
			if v.StakedTokens != types.BigIntToString(amount) {
				t.Fatalf("wrong staked tokens: got: %v, expected: %v", v.StakedTokens, types.BigIntToString(amount))
			}
			if v.Status != defaultStakeStatus {
				t.Fatalf("wrong status; got: %v, expected %v", v.Status, defaultStakeStatus)
			}
			if !ChainsEquality(v.Chains, defaultChainsEdited) {
				t.Fatalf("wrong chains; got: %v, expected: %v", v.Chains, defaultChains)
			}
			if v.ServiceURL != defaultServiceURLEdited {
				t.Fatalf("wrong service url; got: %v, expected: %v", v.ServiceURL, defaultServiceURL)
			}
		}
	}
	return
}

func TestPauseFisherman(t *testing.T) {
	//cdc := Codec()
	utilityModule := NewMockUtilityModule(0, memdb.New(comparer.DefaultComparer, 1000000))
	// init genesis
	genesisState, valKeys, _, _, fishKeys, err := NewMockGenesisState(5, 1, 1, 5)
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
	// build transaction
	privateKeyA := fishKeys[0]
	privateKeyB := valKeys[0]
	pubKeyA := privateKeyA.PublicKey()
	addrA := pubKeyA.Address()
	tx, err := NewTransaction(&types.MessagePauseFisherman{
		Address: addrA,
	}, types.BigIntToString(feeBig))
	if err := tx.Sign(privateKeyA); err != nil {
		t.Fatal(err)
	}
	transactionBytes, err := tx.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if err := utilityModule.CheckTransaction(transactionBytes); err != nil {
		t.Fatal(err)
	}
	if _, err := utilityModule.ApplyBlock(0, privateKeyB.Address(), [][]byte{transactionBytes}, nil); err != nil {
		t.Fatal(err)
	}
	state, err := ExportState(utilityModule)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range state.Fishermen {
		if bytes.Equal(v.Address, addrA) {
			if v.Paused != true {
				t.Fatalf("wrong paused status; got: %v, expected %v", v.Paused, true)
			}
		}
	}
	return
}

func TestUnpauseFisherman(t *testing.T) {
	//cdc := Codec()
	utilityModule := NewMockUtilityModule(0, memdb.New(comparer.DefaultComparer, 1000000))
	// init genesis
	genesisState, valKeys, _, _, fishKeys, err := NewMockGenesisState(5, 1, 1, 5)
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
	// build transaction
	privateKeyA := fishKeys[0]
	privateKeyB := valKeys[0]
	pubKeyA := privateKeyA.PublicKey()
	addrA := pubKeyA.Address()
	pausedBlocks, err := utilityModule.GetFishermanMinimumPauseBlocks()
	if err != nil {
		t.Fatal(err)
	}
	pauseHeight := int64(1)
	utilityModule.LatestHeight = int64(pausedBlocks) + pauseHeight
	if err := utilityModule.SetFishermanPauseHeight(addrA, pauseHeight); err != nil {
		t.Fatal(err)
	}
	tx, err := NewTransaction(&types.MessageUnpauseFisherman{
		Address: addrA,
	}, types.BigIntToString(feeBig))
	if err := tx.Sign(privateKeyA); err != nil {
		t.Fatal(err)
	}
	transactionBytes, err := tx.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if err := utilityModule.CheckTransaction(transactionBytes); err != nil {
		t.Fatal(err)
	}
	if _, err := utilityModule.ApplyBlock(utilityModule.LatestHeight, privateKeyB.Address(), [][]byte{transactionBytes}, nil); err != nil {
		t.Fatal(err)
	}
	state, err := ExportState(utilityModule)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range state.Fishermen {
		if bytes.Equal(v.Address, addrA) {
			if v.Paused != true {
				t.Fatalf("wrong paused status; got: %v, expected %v", v.Paused, false)
			}
		}
	}
	return
}

func TestFishermanPauseServiceNode(t *testing.T) {
	//cdc := Codec()
	utilityModule := NewMockUtilityModule(0, memdb.New(comparer.DefaultComparer, 1000000))
	// init genesis
	genesisState, valKeys, _, snKeys, fishKeys, err := NewMockGenesisState(5, 1, 1, 5)
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
	// build transaction
	privateKeyA := fishKeys[0]
	privateKeyB := valKeys[0]
	privateKeyC := snKeys[0]
	addrA := privateKeyA.Address()
	addrC := privateKeyC.Address()
	tx, err := NewTransaction(&types.MessageFishermanPauseServiceNode{
		Address:  addrC,
		Reporter: addrA,
	}, types.BigIntToString(feeBig))
	if err := tx.Sign(privateKeyA); err != nil {
		t.Fatal(err)
	}
	transactionBytes, err := tx.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if err := utilityModule.CheckTransaction(transactionBytes); err != nil {
		t.Fatal(err)
	}
	if _, err := utilityModule.ApplyBlock(0, privateKeyB.Address(), [][]byte{transactionBytes}, nil); err != nil {
		t.Fatal(err)
	}
	state, err := ExportState(utilityModule)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range state.ServiceNodes {
		if bytes.Equal(v.Address, addrC) {
			if v.Paused != true {
				t.Fatalf("wrong paused status; got: %v, expected %v", v.Paused, true)
			}
		}
	}
	return
}

func TestUnstakeFisherman(t *testing.T) {
	//cdc := Codec()
	utilityModule := NewMockUtilityModule(0, memdb.New(comparer.DefaultComparer, 1000000))
	// init genesis
	genesisState, valKeys, _, _, fishKeys, err := NewMockGenesisState(5, 1, 1, 5)
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
	// build transaction
	privateKeyA := fishKeys[0]
	privateKeyB := valKeys[0]
	pubKeyA := privateKeyA.PublicKey()
	addrA := pubKeyA.Address()
	tx, err := NewTransaction(&types.MessageUnstakeFisherman{
		Address: addrA,
	}, types.BigIntToString(feeBig))
	if err := tx.Sign(privateKeyA); err != nil {
		t.Fatal(err)
	}
	transactionBytes, err := tx.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if err := utilityModule.CheckTransaction(transactionBytes); err != nil {
		t.Fatal(err)
	}
	if _, err := utilityModule.ApplyBlock(0, privateKeyB.Address(), [][]byte{transactionBytes}, nil); err != nil {
		t.Fatal(err)
	}
	state, err := ExportState(utilityModule)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range state.Fishermen {
		if bytes.Equal(v.Address, addrA) {
			if v.Status != types.UnstakingStatus {
				t.Fatalf("wrong status; got: %v, expected %v", v.Status, defaultStakeStatus)
			}
			blocks, err := utilityModule.GetFishermanUnstakingBlocks()
			if err != nil {
				t.Fatal(err)
			}
			utilityModule.LatestHeight = blocks
			readyToUnstake, err := utilityModule.GetFishermenReadyToUnstake()
			if err != nil {
				t.Fatal(err)
			}
			if len(readyToUnstake) != 1 {
				t.Fatalf("should have 1 ready to unstake")
			}
		}
	}
	return
}

func TestStakeApp(t *testing.T) {
	numOfApps := 1
	//cdc := Codec()
	utilityModule := NewMockUtilityModule(0, memdb.New(comparer.DefaultComparer, 1000000))
	// init genesis
	genesisState, valKeys, _, _, _, err := NewMockGenesisState(5, numOfApps, 1, 5)
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
	amountBig := big.NewInt(15000000000)
	// build transaction
	privateKeyA := valKeys[1]
	privateKeyB := valKeys[0]
	pubKeyA := privateKeyA.PublicKey()
	addrA := pubKeyA.Address()
	tx, err := NewTransaction(&types.MessageStakeApp{
		PublicKey:     pubKeyA.Bytes(),
		Chains:        defaultChains,
		Amount:        types.BigIntToString(amountBig),
		OutputAddress: addrA,
	}, types.BigIntToString(feeBig))
	if err := tx.Sign(privateKeyA); err != nil {
		t.Fatal(err)
	}
	transactionBytes, err := tx.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if err := utilityModule.CheckTransaction(transactionBytes); err != nil {
		t.Fatal(err)
	}
	if _, err := utilityModule.ApplyBlock(0, privateKeyB.Address(), [][]byte{transactionBytes}, nil); err != nil {
		t.Fatal(err)
	}
	state, err := ExportState(utilityModule)
	if err != nil {
		t.Fatal(err)
	}
	if len(state.Apps) != numOfApps+1 {
		t.Fatal("wrong app count")
	}
	if ok, _ := utilityModule.GetAppExists(addrA); !ok {
		t.Fatal("newly created app doesn't exist")
	}
	for _, v := range state.Apps {
		if bytes.Equal(v.Address, addrA) {
			if v.StakedTokens != types.BigIntToString(amountBig) {
				t.Fatalf("wrong staked tokens: got: %v, expected: %v", v.StakedTokens, types.BigIntToString(amountBig))
			}
			if v.Status != defaultStakeStatus {
				t.Fatalf("wrong status; got: %v, expected %v", v.Status, defaultStakeStatus)
			}
			if !ChainsEquality(v.Chains, defaultChains) {
				t.Fatalf("wrong chains; got: %v, expected: %v", v.Chains, defaultChains)
			}
			expectedMax, err := utilityModule.CalculateAppRelays(v.StakedTokens)
			if err != nil {
				t.Fatal(err)
			}
			if v.MaxRelays != expectedMax {
				t.Fatalf("wrong max_relays; got %v, expected: %v", v.MaxRelays, expectedMax)
			}
		}
	}
}

func TestEditStakeApp(t *testing.T) {
	//cdc := Codec()
	utilityModule := NewMockUtilityModule(0, memdb.New(comparer.DefaultComparer, 1000000))
	// init genesis
	genesisState, valKeys, appKeys, _, _, err := NewMockGenesisState(5, 1, 1, 5)
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
	amountToAddBig := big.NewInt(1)
	// build transaction
	privateKeyA := appKeys[0]
	privateKeyB := valKeys[0]
	pubKeyA := privateKeyA.PublicKey()
	addrA := pubKeyA.Address()
	tx, err := NewTransaction(&types.MessageEditStakeApp{
		Address:     addrA,
		Chains:      defaultChainsEdited,
		AmountToAdd: types.BigIntToString(amountToAddBig),
	}, types.BigIntToString(feeBig))
	if err := tx.Sign(privateKeyA); err != nil {
		t.Fatal(err)
	}
	transactionBytes, err := tx.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if err := utilityModule.CheckTransaction(transactionBytes); err != nil {
		t.Fatal(err)
	}
	if _, err := utilityModule.ApplyBlock(0, privateKeyB.Address(), [][]byte{transactionBytes}, nil); err != nil {
		t.Fatal(err)
	}
	state, err := ExportState(utilityModule)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range state.Apps {
		if bytes.Equal(v.Address, addrA) {
			amount := &big.Int{}
			amount.Add(defaultStakeBig, amountToAddBig)
			if v.StakedTokens != types.BigIntToString(amount) {
				t.Fatalf("wrong staked tokens: got: %v, expected: %v", v.StakedTokens, types.BigIntToString(amount))
			}
			if v.Status != defaultStakeStatus {
				t.Fatalf("wrong status; got: %v, expected %v", v.Status, defaultStakeStatus)
			}
			if !ChainsEquality(v.Chains, defaultChainsEdited) {
				t.Fatalf("wrong chains; got: %v, expected: %v", v.Chains, defaultChains)
			}
			expectedMax, err := utilityModule.CalculateAppRelays(v.StakedTokens)
			if err != nil {
				t.Fatal(err)
			}
			if v.MaxRelays != expectedMax {
				t.Fatalf("wrong max_relays; got %v, expected: %v", v.MaxRelays, expectedMax)
			}
		}
	}
	return
}

func TestPauseApp(t *testing.T) {
	//cdc := Codec()
	utilityModule := NewMockUtilityModule(0, memdb.New(comparer.DefaultComparer, 1000000))
	// init genesis
	genesisState, valKeys, appKeys, _, _, err := NewMockGenesisState(5, 1, 1, 5)
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
	// build transaction
	privateKeyA := appKeys[0]
	privateKeyB := valKeys[0]
	pubKeyA := privateKeyA.PublicKey()
	addrA := pubKeyA.Address()
	tx, err := NewTransaction(&types.MessagePauseApp{
		Address: addrA,
	}, types.BigIntToString(feeBig))
	if err := tx.Sign(privateKeyA); err != nil {
		t.Fatal(err)
	}
	transactionBytes, err := tx.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if err := utilityModule.CheckTransaction(transactionBytes); err != nil {
		t.Fatal(err)
	}
	if _, err := utilityModule.ApplyBlock(0, privateKeyB.Address(), [][]byte{transactionBytes}, nil); err != nil {
		t.Fatal(err)
	}
	state, err := ExportState(utilityModule)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range state.Apps {
		if bytes.Equal(v.Address, addrA) {
			if v.Paused != true {
				t.Fatalf("wrong paused status; got: %v, expected %v", v.Paused, true)
			}
		}
	}
	return
}

func TestUnpauseApp(t *testing.T) {
	//cdc := Codec()
	utilityModule := NewMockUtilityModule(0, memdb.New(comparer.DefaultComparer, 1000000))
	// init genesis
	genesisState, valKeys, appKeys, _, _, err := NewMockGenesisState(5, 1, 1, 5)
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
	// build transaction
	privateKeyA := appKeys[0]
	privateKeyB := valKeys[0]
	pubKeyA := privateKeyA.PublicKey()
	addrA := pubKeyA.Address()
	pausedBlocks, err := utilityModule.GetAppMinimumPauseBlocks()
	if err != nil {
		t.Fatal(err)
	}
	pauseHeight := int64(1)
	utilityModule.LatestHeight = int64(pausedBlocks) + pauseHeight
	if err := utilityModule.SetAppPauseHeight(addrA, pauseHeight); err != nil {
		t.Fatal(err)
	}
	tx, err := NewTransaction(&types.MessageUnpauseApp{
		Address: addrA,
	}, types.BigIntToString(feeBig))
	if err := tx.Sign(privateKeyA); err != nil {
		t.Fatal(err)
	}
	transactionBytes, err := tx.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if err := utilityModule.CheckTransaction(transactionBytes); err != nil {
		t.Fatal(err)
	}
	if _, err := utilityModule.ApplyBlock(utilityModule.LatestHeight, privateKeyB.Address(), [][]byte{transactionBytes}, nil); err != nil {
		t.Fatal(err)
	}
	state, err := ExportState(utilityModule)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range state.Apps {
		if bytes.Equal(v.Address, addrA) {
			if v.Paused != true {
				t.Fatalf("wrong paused status; got: %v, expected %v", v.Paused, false)
			}
		}
	}
	return
}

func TestUnstakeApp(t *testing.T) {
	//cdc := Codec()
	utilityModule := NewMockUtilityModule(0, memdb.New(comparer.DefaultComparer, 1000000))
	// init genesis
	genesisState, valKeys, appKeys, _, _, err := NewMockGenesisState(5, 1, 1, 5)
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
	// build transaction
	privateKeyA := appKeys[0]
	privateKeyB := valKeys[0]
	pubKeyA := privateKeyA.PublicKey()
	addrA := pubKeyA.Address()
	tx, err := NewTransaction(&types.MessageUnstakeApp{
		Address: addrA,
	}, types.BigIntToString(feeBig))
	if err := tx.Sign(privateKeyA); err != nil {
		t.Fatal(err)
	}
	transactionBytes, err := tx.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if err := utilityModule.CheckTransaction(transactionBytes); err != nil {
		t.Fatal(err)
	}
	if _, err := utilityModule.ApplyBlock(0, privateKeyB.Address(), [][]byte{transactionBytes}, nil); err != nil {
		t.Fatal(err)
	}
	state, err := ExportState(utilityModule)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range state.Apps {
		if bytes.Equal(v.Address, addrA) {
			if v.Status != types.UnstakingStatus {
				t.Fatalf("wrong status; got: %v, expected %v", v.Status, defaultStakeStatus)
			}
			blocks, err := utilityModule.GetAppUnstakingBlocks()
			if err != nil {
				t.Fatal(err)
			}
			utilityModule.LatestHeight = blocks
			readyToUnstake, err := utilityModule.GetAppsReadyToUnstake()
			if err != nil {
				t.Fatal(err)
			}
			if len(readyToUnstake) != 1 {
				t.Fatalf("should have 1 ready to unstake")
			}
		}
	}
	return
}

func TestStakeServiceNode(t *testing.T) {
	numOfServiceNodes := 5
	//cdc := Codec()
	utilityModule := NewMockUtilityModule(0, memdb.New(comparer.DefaultComparer, 1000000))
	// init genesis
	genesisState, valKeys, _, _, _, err := NewMockGenesisState(5, 1, 1, numOfServiceNodes)
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
	amountBig := big.NewInt(15000000000)
	// build transaction
	privateKeyA := valKeys[1]
	privateKeyB := valKeys[0]
	pubKeyA := privateKeyA.PublicKey()
	addrA := pubKeyA.Address()
	tx, err := NewTransaction(&types.MessageStakeServiceNode{
		PublicKey:     pubKeyA.Bytes(),
		Chains:        defaultChains,
		Amount:        types.BigIntToString(amountBig),
		ServiceURL:    defaultServiceURL,
		OutputAddress: addrA,
	}, types.BigIntToString(feeBig))
	if err := tx.Sign(privateKeyA); err != nil {
		t.Fatal(err)
	}
	transactionBytes, err := tx.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if err := utilityModule.CheckTransaction(transactionBytes); err != nil {
		t.Fatal(err)
	}
	if _, err := utilityModule.ApplyBlock(0, privateKeyB.Address(), [][]byte{transactionBytes}, nil); err != nil {
		t.Fatal(err)
	}
	state, err := ExportState(utilityModule)
	if err != nil {
		t.Fatal(err)
	}
	if len(state.ServiceNodes) != numOfServiceNodes+1 {
		t.Fatal("wrong sn count")
	}
	if ok, _ := utilityModule.GetServiceNodeExists(addrA); !ok {
		t.Fatal("newly created sn doesn't exist")
	}
	for _, v := range state.ServiceNodes {
		if bytes.Equal(v.Address, addrA) {
			if v.StakedTokens != types.BigIntToString(amountBig) {
				t.Fatalf("wrong staked tokens: got: %v, expected: %v", v.StakedTokens, types.BigIntToString(amountBig))
			}
			if v.Status != defaultStakeStatus {
				t.Fatalf("wrong status; got: %v, expected %v", v.Status, defaultStakeStatus)
			}
			if !ChainsEquality(v.Chains, defaultChains) {
				t.Fatalf("wrong chains; got: %v, expected: %v", v.Chains, defaultChains)
			}
			if v.ServiceURL != defaultServiceURL {
				t.Fatalf("wrong service url; got: %v, expected: %v", v.ServiceURL, defaultServiceURL)
			}
		}
	}
}

func TestEditStakeServiceNode(t *testing.T) {
	//cdc := Codec()
	utilityModule := NewMockUtilityModule(0, memdb.New(comparer.DefaultComparer, 1000000))
	// init genesis
	genesisState, valKeys, _, serviceNodeKeys, _, err := NewMockGenesisState(5, 1, 1, 5)
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
	amountToAddBig := big.NewInt(1)
	// build transaction
	privateKeyA := serviceNodeKeys[0]
	privateKeyB := valKeys[0]
	pubKeyA := privateKeyA.PublicKey()
	addrA := pubKeyA.Address()
	tx, err := NewTransaction(&types.MessageEditStakeServiceNode{
		Address:     addrA,
		Chains:      defaultChainsEdited,
		AmountToAdd: types.BigIntToString(amountToAddBig),
		ServiceURL:  defaultServiceURLEdited,
	}, types.BigIntToString(feeBig))
	if err := tx.Sign(privateKeyA); err != nil {
		t.Fatal(err)
	}
	transactionBytes, err := tx.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if err := utilityModule.CheckTransaction(transactionBytes); err != nil {
		t.Fatal(err)
	}
	if _, err := utilityModule.ApplyBlock(0, privateKeyB.Address(), [][]byte{transactionBytes}, nil); err != nil {
		t.Fatal(err)
	}
	state, err := ExportState(utilityModule)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range state.ServiceNodes {
		if bytes.Equal(v.Address, addrA) {
			amount := &big.Int{}
			amount.Add(defaultStakeBig, amountToAddBig)
			if v.StakedTokens != types.BigIntToString(amount) {
				t.Fatalf("wrong staked tokens: got: %v, expected: %v", v.StakedTokens, types.BigIntToString(amount))
			}
			if v.ServiceURL != defaultServiceURLEdited {
				t.Fatalf("wrong serviceurl: got: %v, expected: %v", v.ServiceURL, defaultServiceURLEdited)
			}
			if v.Status != defaultStakeStatus {
				t.Fatalf("wrong status; got: %v, expected %v", v.Status, defaultStakeStatus)
			}
			if !ChainsEquality(v.Chains, defaultChainsEdited) {
				t.Fatalf("wrong chains; got: %v, expected: %v", v.Chains, defaultChains)
			}
		}
	}
	return
}

func TestPauseServiceNode(t *testing.T) {
	//cdc := Codec()
	utilityModule := NewMockUtilityModule(0, memdb.New(comparer.DefaultComparer, 1000000))
	// init genesis
	genesisState, valKeys, _, serviceNodeKeys, _, err := NewMockGenesisState(5, 1, 1, 5)
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
	// build transaction
	privateKeyA := serviceNodeKeys[0]
	privateKeyB := valKeys[0]
	pubKeyA := privateKeyA.PublicKey()
	addrA := pubKeyA.Address()
	tx, err := NewTransaction(&types.MessagePauseServiceNode{
		Address: addrA,
	}, types.BigIntToString(feeBig))
	if err := tx.Sign(privateKeyA); err != nil {
		t.Fatal(err)
	}
	transactionBytes, err := tx.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if err := utilityModule.CheckTransaction(transactionBytes); err != nil {
		t.Fatal(err)
	}
	if _, err := utilityModule.ApplyBlock(0, privateKeyB.Address(), [][]byte{transactionBytes}, nil); err != nil {
		t.Fatal(err)
	}
	state, err := ExportState(utilityModule)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range state.ServiceNodes {
		if bytes.Equal(v.Address, addrA) {
			if v.Paused != true {
				t.Fatalf("wrong paused status; got: %v, expected %v", v.Paused, true)
			}
		}
	}
	return
}

func TestUnpauseServiceNode(t *testing.T) {
	//cdc := Codec()
	utilityModule := NewMockUtilityModule(0, memdb.New(comparer.DefaultComparer, 1000000))
	// init genesis
	genesisState, valKeys, _, snKeys, _, err := NewMockGenesisState(5, 1, 1, 5)
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
	// build transaction
	privateKeyA := snKeys[0]
	privateKeyB := valKeys[0]
	pubKeyA := privateKeyA.PublicKey()
	addrA := pubKeyA.Address()
	pausedBlocks, err := utilityModule.GetServiceNodeMinimumPauseBlocks()
	if err != nil {
		t.Fatal(err)
	}
	pauseHeight := int64(1)
	utilityModule.LatestHeight = int64(pausedBlocks) + pauseHeight
	if err := utilityModule.SetServiceNodePauseHeight(addrA, pauseHeight); err != nil {
		t.Fatal(err)
	}
	tx, err := NewTransaction(&types.MessageUnpauseServiceNode{
		Address: addrA,
	}, types.BigIntToString(feeBig))
	if err := tx.Sign(privateKeyA); err != nil {
		t.Fatal(err)
	}
	transactionBytes, err := tx.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if err := utilityModule.CheckTransaction(transactionBytes); err != nil {
		t.Fatal(err)
	}
	if _, err := utilityModule.ApplyBlock(utilityModule.LatestHeight, privateKeyB.Address(), [][]byte{transactionBytes}, nil); err != nil {
		t.Fatal(err)
	}
	state, err := ExportState(utilityModule)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range state.ServiceNodes {
		if bytes.Equal(v.Address, addrA) {
			if v.Paused != true {
				t.Fatalf("wrong paused status; got: %v, expected %v", v.Paused, false)
			}
		}
	}
	return
}

func TestUnstakeServiceNode(t *testing.T) {
	//cdc := Codec()
	utilityModule := NewMockUtilityModule(0, memdb.New(comparer.DefaultComparer, 1000000))
	// init genesis
	genesisState, valKeys, _, serviceNodeKeys, _, err := NewMockGenesisState(5, 1, 1, 5)
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
	// build transaction
	privateKeyA := serviceNodeKeys[0]
	privateKeyB := valKeys[0]
	pubKeyA := privateKeyA.PublicKey()
	addrA := pubKeyA.Address()
	tx, err := NewTransaction(&types.MessageUnstakeServiceNode{
		Address: addrA,
	}, types.BigIntToString(feeBig))
	if err := tx.Sign(privateKeyA); err != nil {
		t.Fatal(err)
	}
	transactionBytes, err := tx.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if err := utilityModule.CheckTransaction(transactionBytes); err != nil {
		t.Fatal(err)
	}
	if _, err := utilityModule.ApplyBlock(0, privateKeyB.Address(), [][]byte{transactionBytes}, nil); err != nil {
		t.Fatal(err)
	}
	state, err := ExportState(utilityModule)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range state.ServiceNodes {
		if bytes.Equal(v.Address, addrA) {
			if v.Status != types.UnstakingStatus {
				t.Fatalf("wrong status; got: %v, expected %v", v.Status, defaultStakeStatus)
			}
			blocks, err := utilityModule.GetServiceNodeUnstakingBlocks()
			if err != nil {
				t.Fatal(err)
			}
			utilityModule.LatestHeight = blocks
			readyToUnstake, err := utilityModule.GetServiceNodesReadyToUnstake()
			if err != nil {
				t.Fatal(err)
			}
			if len(readyToUnstake) != 1 {
				t.Fatalf("should have 1 ready to unstake")
			}
		}
	}
	return
}

func TestChangeParameter(t *testing.T) {
	cdc := Codec()
	utilityModule := NewMockUtilityModule(0, memdb.New(comparer.DefaultComparer, 1000000))
	// init genesis
	genesisState, valKeys, _, _, _, err := NewMockGenesisState(5, 1, 1, 5)
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
	// build transaction
	privateKeyA := DefaultParamsOwner
	privateKeyB := valKeys[0]
	newBlocksPerSessionValue := int32(10)
	newBlocksPerSession := wrapperspb.Int32(newBlocksPerSessionValue)
	a, err := cdc.ToAny(newBlocksPerSession)
	if err != nil {
		t.Fatal(err)
	}
	tx, err := NewTransaction(&types.MessageChangeParameter{
		Owner:          privateKeyA.Address(),
		ParameterKey:   types.BlocksPerSessionParamName,
		ParameterValue: a,
	}, types.BigIntToString(feeBig))
	if err := tx.Sign(privateKeyA); err != nil {
		t.Fatal(err)
	}
	transactionBytes, err := tx.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if err := utilityModule.CheckTransaction(transactionBytes); err != nil {
		t.Fatal(err)
	}
	if _, err := utilityModule.ApplyBlock(0, privateKeyB.Address(), [][]byte{transactionBytes}, nil); err != nil {
		t.Fatal(err)
	}
	state, err := ExportState(utilityModule)
	if err != nil {
		t.Fatal(err)
	}
	if state.Params.BlocksPerSession != newBlocksPerSessionValue {
		t.Fatalf("wrong parameter value; got: %v expected: %v\n", state.Params.BlocksPerSession, newBlocksPerSessionValue)
	}
	return
}
