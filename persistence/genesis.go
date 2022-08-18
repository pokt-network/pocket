package persistence

import (
	"encoding/hex"
	"fmt"
	"log"
	"math/big"

	"github.com/pokt-network/pocket/persistence/schema"
	"github.com/pokt-network/pocket/shared/types"
	"github.com/pokt-network/pocket/shared/types/genesis"
)

// TODO(Andrew): generalize with the `actors interface`` once merged with #111
// WARNING: This function crashes the process if there is an error populating the genesis state.
func (m *persistenceModule) populateGenesisState(state *genesis.GenesisState) {
	log.Println("Populating genesis state...")

	// REFACTOR: This business logic should probably live in `types/genesis.go`
	//           and we need to add proper unit tests for it.`
	poolValues := make(map[string]*big.Int, 0)
	addValueToPool := func(poolName string, valueToAdd string) error {
		value, err := types.StringToBigInt(valueToAdd)
		if err != nil {
			return err
		}
		if present := poolValues[poolName]; present == nil {
			poolValues[poolName] = big.NewInt(0)
		}
		poolValues[poolName].Add(poolValues[poolName], value)
		return nil
	}

	log.Println("Populating genesis state...")
	rwContext, err := m.NewRWContext(0)
	if err != nil {
		log.Fatalf("an error occurred creating the rwContext for the genesis state: %s", err.Error())
	}
	defer rwContext.Commit()

	if err != nil {
		log.Fatal(fmt.Sprintf("an error occurred creating the rwContext for the genesis state: %s", err.Error()))
	}
	for _, acc := range state.Utility.Accounts {
		addrBz, err := hex.DecodeString(acc.Address)
		if err != nil {
			log.Fatal(fmt.Sprintf("an error occurred converting address to bytes %s", acc.Address))
		}
		err = rwContext.SetAccountAmount(addrBz, acc.Amount)
		if err != nil {
			log.Fatal(fmt.Sprintf("an error occurred inserting an acc in the genesis state: %s", err.Error()))
		}
	}
	for _, pool := range state.Utility.Pools {
		poolNameBytes := []byte(pool.Address)
		err = rwContext.InsertPool(pool.Address, poolNameBytes, pool.Amount)
		if err != nil {
			log.Fatal(fmt.Sprintf("an error occurred inserting an pool in the genesis state: %s", err.Error()))
		}
	}
	for _, act := range state.Utility.Applications { // TODO (Andrew) genericize the genesis population logic for actors #163
		addrBz, err := hex.DecodeString(act.Address)
		if err != nil {
			log.Fatal(fmt.Sprintf("an error occurred converting address to bytes %s", act.Address))
		}
		pubKeyBz, err := hex.DecodeString(act.PublicKey)
		if err != nil {
			log.Fatal(fmt.Sprintf("an error occurred converting pubKey to bytes %s", act.PublicKey))
		}
		outputBz, err := hex.DecodeString(act.Output)
		if err != nil {
			log.Fatal(fmt.Sprintf("an error occurred converting output to bytes %s", act.Output))
		}
		err = rwContext.InsertApp(addrBz, pubKeyBz, outputBz, false, StakedStatus, act.GenericParam, act.StakedAmount, act.Chains, act.PausedHeight, act.UnstakingHeight)
		if err != nil {
			log.Fatal(fmt.Sprintf("an error occurred inserting an app in the genesis state: %s", err.Error()))
		}
		if err = addValueToPool(genesis.Pool_Names_AppStakePool.String(), act.StakedAmount); err != nil {
			log.Fatal(fmt.Sprintf("an error occurred inserting staked tokens into %s pool", genesis.Pool_Names_AppStakePool))
		}
	}
	for _, act := range state.Utility.ServiceNodes {
		addrBz, err := hex.DecodeString(act.Address)
		if err != nil {
			log.Fatal(fmt.Sprintf("an error occurred converting address to bytes %s", act.Address))
		}
		pubKeyBz, err := hex.DecodeString(act.PublicKey)
		if err != nil {
			log.Fatal(fmt.Sprintf("an error occurred converting pubKey to bytes %s", act.PublicKey))
		}
		outputBz, err := hex.DecodeString(act.Output)
		if err != nil {
			log.Fatal(fmt.Sprintf("an error occurred converting output to bytes %s", act.Output))
		}
		err = rwContext.InsertServiceNode(addrBz, pubKeyBz, outputBz, false, StakedStatus, act.GenericParam, act.StakedAmount, act.Chains, act.PausedHeight, act.UnstakingHeight)
		if err != nil {
			log.Fatal(fmt.Sprintf("an error occurred inserting a service node in the genesis state: %s", err.Error()))
		}
		if err = addValueToPool(genesis.Pool_Names_ServiceNodeStakePool.String(), act.StakedAmount); err != nil {
			log.Fatal(fmt.Sprintf("an error occurred inserting staked tokens into %s pool", genesis.Pool_Names_ServiceNodeStakePool.String()))
		}
	}
	for _, act := range state.Utility.Fishermen {
		addrBz, err := hex.DecodeString(act.Address)
		if err != nil {
			log.Fatal(fmt.Sprintf("an error occurred converting address to bytes %s", act.Address))
		}
		pubKeyBz, err := hex.DecodeString(act.PublicKey)
		if err != nil {
			log.Fatal(fmt.Sprintf("an error occurred converting pubKey to bytes %s", act.PublicKey))
		}
		outputBz, err := hex.DecodeString(act.Output)
		if err != nil {
			log.Fatal(fmt.Sprintf("an error occurred converting output to bytes %s", act.Output))
		}
		err = rwContext.InsertFisherman(addrBz, pubKeyBz, outputBz, false, StakedStatus, act.GenericParam, act.StakedAmount, act.Chains, act.PausedHeight, act.UnstakingHeight)
		if err != nil {
			log.Fatal(fmt.Sprintf("an error occurred inserting a fisherman in the genesis state: %s", err.Error()))
		}
		if err = addValueToPool(genesis.Pool_Names_FishermanStakePool.String(), act.StakedAmount); err != nil {
			log.Fatal(fmt.Sprintf("an error occurred inserting staked tokens into %s pool", genesis.Pool_Names_FishermanStakePool.String()))
		}
	}
	for _, act := range state.Utility.Validators {
		addrBz, err := hex.DecodeString(act.Address)
		if err != nil {
			log.Fatal(fmt.Sprintf("an error occurred converting address to bytes %s", act.Address))
		}
		pubKeyBz, err := hex.DecodeString(act.PublicKey)
		if err != nil {
			log.Fatal(fmt.Sprintf("an error occurred converting pubKey to bytes %s", act.PublicKey))
		}
		outputBz, err := hex.DecodeString(act.Output)
		if err != nil {
			log.Fatal(fmt.Sprintf("an error occurred converting output to bytes %s", act.Output))
		}
		err = rwContext.InsertValidator(addrBz, pubKeyBz, outputBz, false, StakedStatus, act.GenericParam, act.StakedAmount, act.PausedHeight, act.UnstakingHeight)
		if err != nil {
			log.Fatal(fmt.Sprintf("an error occurred inserting a validator in the genesis state: %s", err.Error()))
		}
		if err = addValueToPool(genesis.Pool_Names_ValidatorStakePool.String(), act.StakedAmount); err != nil {
			log.Fatal(fmt.Sprintf("an error occurred inserting staked tokens into %s pool", genesis.Pool_Names_ValidatorStakePool.String()))
		}
	}
	// TODO(team): use params from genesis file - not the hardcoded
	if err = rwContext.InitParams(); err != nil {
		log.Fatalf("an error occurred initializing params: %s", err.Error())
	}

	if err = rwContext.InitFlags(); err != nil { // TODO (Team) use flags from genesis file not hardcoded
		log.Fatalf("an error occurred initializing flags: %s", err.Error())
	}

	if err = rwContext.Commit(); err != nil {
		log.Fatalf("an error occurred during commit() on genesis state %s ", err.Error())
	}
}

// TODO(pocket/issues/149): All of the functions below following a structure similar to `GetAll<Actor>`
//  can easily be refactored and condensed into a single function using a generic type or a common
// interface.
func (p PostgresContext) GetAllAccounts(height int64) (accs []*genesis.Account, err error) {
	ctx, txn, err := p.DB.GetCtxAndTxn()
	if err != nil {
		return nil, err
	}
	rows, err := txn.Query(ctx, schema.SelectAccounts(height, schema.AccountTableName))
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		acc := new(genesis.Account)
		var address, balance string
		if err = rows.Scan(&address, &balance, &height); err != nil {
			return nil, err
		}
		acc.Address, err = hex.DecodeString(address)
		if err != nil {
			return nil, err
		}
		acc.Amount = balance
		accs = append(accs, acc)
	}
	return
}

func (p PostgresContext) GetAllPools(height int64) (accs []*genesis.Account, err error) {
	ctx, txn, err := p.DB.GetCtxAndTxn()
	if err != nil {
		return nil, err
	}
	rows, err := txn.Query(ctx, schema.SelectPools(height, schema.PoolTableName))
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		pool := new(genesis.Pool)
		pool.Account = new(genesis.Account)
		var name, balance string
		if err = rows.Scan(&name, &balance, &height); err != nil {
			return nil, err
		}
		pool.Address = name
		if err != nil {
			return nil, err
		}
		pool.Amount = balance
		accs = append(accs, pool)
	}
	return
}

func (p PostgresContext) GetAllApps(height int64) (apps []*genesis.Actor, err error) {
	ctx, txn, err := p.DB.GetCtxAndTxn()
	if err != nil {
		return nil, err
	}
	rows, err := txn.Query(ctx, schema.ApplicationActor.GetAllQuery(height))
	if err != nil {
		return nil, err
	}
	var actors []schema.BaseActor
	for rows.Next() {
		var actor schema.BaseActor
		actor, height, err = p.GetActorFromRow(rows)
		if err != nil {
			return
		}
		actors = append(actors, actor)
	}
	rows.Close()
	for _, actor := range actors {
		actor, err = p.GetChainsForActor(ctx, txn, schema.ApplicationActor, actor, height)
		if err != nil {
			return
		}
		apps = append(apps, p.BaseActorToActor(actor, genesis.ActorType_App))
	}
	return
}

func (p PostgresContext) GetAllValidators(height int64) (vals []*genesis.Actor, err error) {
	ctx, txn, err := p.DB.GetCtxAndTxn()
	if err != nil {
		return nil, err
	}
	rows, err := txn.Query(ctx, schema.ValidatorActor.GetAllQuery(height))
	if err != nil {
		return nil, err
	}
	var actors []schema.BaseActor
	for rows.Next() {
		var actor schema.BaseActor
		actor, height, err = p.GetActorFromRow(rows)
		if err != nil {
			return
		}
		actors = append(actors, actor)
	}
	rows.Close()
	for _, actor := range actors {
		actor, err = p.GetChainsForActor(ctx, txn, schema.ApplicationActor, actor, height)
		if err != nil {
			return
		}
		vals = append(vals, p.BaseActorToActor(actor, genesis.ActorType_Val))
	}
	return
}

func (p PostgresContext) GetAllServiceNodes(height int64) (sn []*genesis.Actor, err error) {
	ctx, txn, err := p.DB.GetCtxAndTxn()
	if err != nil {
		return nil, err
	}
	rows, err := txn.Query(ctx, schema.ServiceNodeActor.GetAllQuery(height))
	if err != nil {
		return nil, err
	}
	var actors []schema.BaseActor
	for rows.Next() {
		var actor schema.BaseActor
		actor, height, err = p.GetActorFromRow(rows)
		if err != nil {
			return
		}
		actors = append(actors, actor)
	}
	rows.Close()
	for _, actor := range actors {
		actor, err = p.GetChainsForActor(ctx, txn, schema.ServiceNodeActor, actor, height)
		if err != nil {
			return
		}
		sn = append(sn, p.BaseActorToActor(actor, genesis.ActorType_Node))
	}
	return
}

func (p PostgresContext) GetAllFishermen(height int64) (f []*genesis.Actor, err error) {
	ctx, txn, err := p.DB.GetCtxAndTxn()
	if err != nil {
		return nil, err
	}
	rows, err := txn.Query(ctx, schema.FishermanActor.GetAllQuery(height))
	if err != nil {
		return nil, err
	}
	var actors []schema.BaseActor
	for rows.Next() {
		var actor schema.BaseActor
		actor, height, err = p.GetActorFromRow(rows)
		if err != nil {
			return
		}
		actors = append(actors, actor)
	}
	rows.Close()
	for _, actor := range actors {
		actor, err = p.GetChainsForActor(ctx, txn, schema.FishermanActor, actor, height)
		if err != nil {
			return
		}
		f = append(f, p.BaseActorToActor(actor, genesis.ActorType_Fish))
	}
	return
}

func (p PostgresContext) BaseActorToActor(ba schema.BaseActor, actorType genesis.ActorType) *genesis.Actor { // TODO (Team) deprecate with interface #163
	actor := new(genesis.Actor)
	actor.ActorType = actorType
	actor.Address = ba.Address
	actor.PublicKey = ba.PublicKey
	actor.StakedAmount = ba.StakedTokens
	actor.GenericParam = ba.ActorSpecificParam
	actor.PausedHeight = ba.PausedHeight
	actor.UnstakingHeight = ba.UnstakingHeight
	actor.Output = ba.OutputAddress
	actor.Chains = ba.Chains
	return actor
}
