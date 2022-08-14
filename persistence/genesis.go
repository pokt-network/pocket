package persistence

import (
	"encoding/hex"
	"log"

	"math/big"

	"github.com/pokt-network/pocket/persistence/schema"
	"github.com/pokt-network/pocket/shared/types"
	"github.com/pokt-network/pocket/shared/types/genesis"
)

// TODO(Andrew): generalize with the `actors interface`` once merged with #111
func (pm *persistenceModule) populateGenesisState(state *genesis.GenesisState) {
	log.Println("Populating genesis state...")

	// REFACTOR: This business logic should probably live in `types/genesis.go`
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

	rwContext, err := pm.NewRWContext(0)
	if err != nil {
		log.Fatalf("an error occurred creating the rwContext for the genesis state: %s", err.Error())
	}

	for _, act := range state.Apps {
		if err = rwContext.InsertApp(act.Address, act.PublicKey, act.Output, act.Paused, int(act.Status), act.MaxRelays, act.StakedTokens, act.Chains, act.PausedHeight, act.UnstakingHeight); err != nil {
			log.Fatalf("an error occurred inserting an app in the genesis state: %s", err.Error())
		}
		if err = addValueToPool(genesis.AppStakePoolName, act.StakedTokens); err != nil {
			log.Fatalf("an error occurred inserting staked tokens into %s pool", genesis.AppStakePoolName)
		}
	}

	for _, act := range state.ServiceNodes {
		if err = rwContext.InsertServiceNode(act.Address, act.PublicKey, act.Output, act.Paused, int(act.Status), act.ServiceUrl, act.StakedTokens, act.Chains, act.PausedHeight, act.UnstakingHeight); err != nil {
			log.Fatalf("an error occurred inserting a service node in the genesis state: %s", err.Error())
		}
		if err = addValueToPool(genesis.ServiceNodeStakePoolName, act.StakedTokens); err != nil {
			log.Fatalf("an error occurred inserting staked tokens into %s pool", genesis.ServiceNodeStakePoolName)
		}
	}

	for _, act := range state.Fishermen {
		if err = rwContext.InsertFisherman(act.Address, act.PublicKey, act.Output, act.Paused, int(act.Status), act.ServiceUrl, act.StakedTokens, act.Chains, act.PausedHeight, act.UnstakingHeight); err != nil {
			log.Fatalf("an error occurred inserting a fisherman in the genesis state: %s", err.Error())
		}
		if err = addValueToPool(genesis.FishermanStakePoolName, act.StakedTokens); err != nil {
			log.Fatalf("an error occurred inserting staked tokens into %s pool", genesis.FishermanStakePoolName)
		}
	}

	for _, act := range state.Validators {
		if err = rwContext.InsertValidator(act.Address, act.PublicKey, act.Output, act.Paused, int(act.Status), act.ServiceUrl, act.StakedTokens, act.PausedHeight, act.UnstakingHeight); err != nil {
			log.Fatalf("an error occurred inserting a validator in the genesis state: %s", err.Error())
		}
		if err = addValueToPool(genesis.ValidatorStakePoolName, act.StakedTokens); err != nil {
			log.Fatalf("an error occurred inserting staked tokens into %s pool", genesis.ValidatorStakePoolName)
		}
	}

	// DISCUSS_IN_THIS_COMMIT: Do protocol actors need corresponding accounts?
	// See reference: https://github.com/pokt-network/pocket/pull/140#discussion_r939742930
	for _, acc := range state.Accounts {
		if err = rwContext.SetAccountAmount(acc.Address, acc.Amount); err != nil {
			log.Fatalf("an error occurred inserting an acc in the genesis state: %s", err.Error())
		}
	}

	for _, pool := range state.Pools {
		// REFACTOR(pocket/issues/154): This validation logic needs to live in `types/genesis.go` and
		// we need to add unit tests for it too.
		poolAmount, err := types.StringToBigInt(pool.Account.Amount)
		if err != nil {
			log.Fatalf("an error occurred converting the pool amount to a big.Int: %s", err.Error())
		}
		if poolValues[pool.Name] != poolAmount {
			log.Printf("[WARNING] The pool amount computed for %s does not match the amount in the pool", pool.Name)
		}

		if err := rwContext.InsertPool(pool.Name, pool.Account.Address, pool.Account.Amount); err != nil {
			log.Fatalf("an error occurred inserting an pool in the genesis state: %s", err.Error())
		}
	}

	// TODO(team): use params from genesis file not hardcoded
	if err = rwContext.InitParams(); err != nil {
		log.Fatalf("an error occurred initializing params: %s", err.Error())
	}

	if err = rwContext.Commit(); err != nil {
		log.Fatalf("an error occurred during commit() on genesis state %s ", err.Error())
	}
}

// TODO: GetAll<Actor> can easily be refactored and condensed into a single function using a generic type
// or a common interface. Left for now per https://github.com/pokt-network/pocket/pull/140/files#r939745088.
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

func (p PostgresContext) GetAllPools(height int64) (accs []*genesis.Pool, err error) {
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
		pool.Name = name
		if err != nil {
			return nil, err
		}
		pool.Account.Amount = balance
		accs = append(accs, pool)
	}
	return
}

func (p PostgresContext) GetAllApps(height int64) (apps []*genesis.App, err error) {
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
		var app *genesis.App
		actor, err = p.GetChainsForActor(ctx, txn, schema.ApplicationActor, actor, height)
		if err != nil {
			return
		}
		app, err = p.ActorToApp(actor)
		if err != nil {
			return
		}
		apps = append(apps, app)
	}
	return
}

func (p PostgresContext) GetAllValidators(height int64) (vals []*genesis.Validator, err error) {
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
		var val *genesis.Validator
		actor, err = p.GetChainsForActor(ctx, txn, schema.ApplicationActor, actor, height)
		if err != nil {
			return
		}
		val, err = p.ActorToValidator(actor)
		if err != nil {
			return
		}
		vals = append(vals, val)
	}
	return
}

func (p PostgresContext) GetAllServiceNodes(height int64) (sn []*genesis.ServiceNode, err error) {
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
		var ser *genesis.ServiceNode
		actor, err = p.GetChainsForActor(ctx, txn, schema.ServiceNodeActor, actor, height)
		if err != nil {
			return
		}
		ser, err = p.ActorToServiceNode(actor)
		if err != nil {
			return
		}
		sn = append(sn, ser)
	}
	return
}

func (p PostgresContext) GetAllFishermen(height int64) (f []*genesis.Fisherman, err error) {
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
		var fish *genesis.Fisherman
		actor, err = p.GetChainsForActor(ctx, txn, schema.FishermanActor, actor, height)
		if err != nil {
			return
		}
		fish, err = p.ActorToFish(actor)
		if err != nil {
			return
		}
		f = append(f, fish)
	}
	return
}

// TODO (Team) once we move away from BaseActor we can simplify and genericize a lot of this
func (p PostgresContext) ActorToApp(actor schema.BaseActor) (*genesis.App, error) {
	addr, err := hex.DecodeString(actor.Address)
	if err != nil {
		return nil, err
	}
	pubKey, err := hex.DecodeString(actor.PublicKey)
	if err != nil {
		return nil, err
	}
	output, err := hex.DecodeString(actor.OutputAddress)
	if err != nil {
		return nil, err
	}
	status := int32(2)
	if actor.UnstakingHeight != types.HeightNotUsed && actor.UnstakingHeight != 0 {
		status = 1
	}
	return &genesis.App{
		Address:         addr,
		PublicKey:       pubKey,
		Paused:          actor.PausedHeight != types.HeightNotUsed && actor.PausedHeight != 0,
		Status:          status,
		Chains:          actor.Chains,
		MaxRelays:       actor.ActorSpecificParam,
		StakedTokens:    actor.StakedTokens,
		PausedHeight:    actor.PausedHeight,
		UnstakingHeight: actor.UnstakingHeight,
		Output:          output,
	}, nil
}

func (p PostgresContext) ActorToFish(actor schema.BaseActor) (*genesis.Fisherman, error) {
	addr, err := hex.DecodeString(actor.Address)
	if err != nil {
		return nil, err
	}
	pubKey, err := hex.DecodeString(actor.PublicKey)
	if err != nil {
		return nil, err
	}
	output, err := hex.DecodeString(actor.OutputAddress)
	if err != nil {
		return nil, err
	}
	status := int32(2)
	if actor.UnstakingHeight != types.HeightNotUsed && actor.UnstakingHeight != 0 {
		status = 1
	}
	return &genesis.Fisherman{
		Address:         addr,
		PublicKey:       pubKey,
		Paused:          actor.PausedHeight != types.HeightNotUsed && actor.PausedHeight != 0,
		Status:          status,
		Chains:          actor.Chains,
		ServiceUrl:      actor.ActorSpecificParam,
		StakedTokens:    actor.StakedTokens,
		PausedHeight:    actor.PausedHeight,
		UnstakingHeight: actor.UnstakingHeight,
		Output:          output,
	}, nil
}

func (p PostgresContext) ActorToServiceNode(actor schema.BaseActor) (*genesis.ServiceNode, error) {
	addr, err := hex.DecodeString(actor.Address)
	if err != nil {
		return nil, err
	}
	pubKey, err := hex.DecodeString(actor.PublicKey)
	if err != nil {
		return nil, err
	}
	output, err := hex.DecodeString(actor.OutputAddress)
	if err != nil {
		return nil, err
	}
	status := int32(2)
	if actor.UnstakingHeight != types.HeightNotUsed && actor.UnstakingHeight != 0 {
		status = 1
	}
	return &genesis.ServiceNode{
		Address:         addr,
		PublicKey:       pubKey,
		Paused:          actor.PausedHeight != types.HeightNotUsed && actor.PausedHeight != 0,
		Status:          status,
		Chains:          actor.Chains,
		ServiceUrl:      actor.ActorSpecificParam,
		StakedTokens:    actor.StakedTokens,
		PausedHeight:    actor.PausedHeight,
		UnstakingHeight: actor.UnstakingHeight,
		Output:          output,
	}, nil
}

func (p PostgresContext) ActorToValidator(actor schema.BaseActor) (*genesis.Validator, error) {
	addr, err := hex.DecodeString(actor.Address)
	if err != nil {
		return nil, err
	}
	pubKey, err := hex.DecodeString(actor.PublicKey)
	if err != nil {
		return nil, err
	}
	output, err := hex.DecodeString(actor.OutputAddress)
	if err != nil {
		return nil, err
	}
	status := int32(2)
	if actor.UnstakingHeight != types.HeightNotUsed && actor.UnstakingHeight != 0 {
		status = 1
	}
	return &genesis.Validator{
		Address:         addr,
		PublicKey:       pubKey,
		Paused:          actor.PausedHeight != types.HeightNotUsed && actor.PausedHeight != 0,
		Status:          status,
		ServiceUrl:      actor.ActorSpecificParam,
		StakedTokens:    actor.StakedTokens,
		PausedHeight:    actor.PausedHeight,
		UnstakingHeight: actor.UnstakingHeight,
		Output:          output,
	}, nil
}
