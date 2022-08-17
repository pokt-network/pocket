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
// WARNING: This function crashes the process if there is an error populating the genesis state.
func (m *persistenceModule) populateGenesisState(state *genesis.GenesisState) {
	log.Println("Populating genesis state...")

	// HACK: This is needed to avoid block a previous from interfering with the genesis state hydration
	// until proper state sync is implemented.
	deleteContext, err := m.NewRWContext(0)
	if err != nil {
		log.Fatalf("an error occurred creating the rwContext to prepare for the genesis state: %s", err.Error())
	}
	deleteContext.Close()
	// END HACK

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

	rwContext, err := m.NewRWContext(0)
	if err != nil {
		log.Fatalf("an error occurred creating the rwContext for the genesis state: %s", err.Error())
	}
	defer rwContext.Commit()

	for _, app := range state.Apps {
		if err = rwContext.InsertApp(app.Address, app.PublicKey, app.Output, app.Paused, int(app.Status), app.MaxRelays, app.StakedTokens, app.Chains, app.PausedHeight, app.UnstakingHeight); err != nil {
			log.Fatalf("an error occurred inserting an app in the genesis state: %s", err.Error())
		}
		if err = addValueToPool(genesis.AppStakePoolName, app.StakedTokens); err != nil {
			log.Fatalf("an error occurred inserting staked tokens into %s pool", genesis.AppStakePoolName)
		}
	}

	for _, serviceNode := range state.ServiceNodes {
		if err = rwContext.InsertServiceNode(serviceNode.Address, serviceNode.PublicKey, serviceNode.Output, serviceNode.Paused, int(serviceNode.Status), serviceNode.ServiceUrl, serviceNode.StakedTokens, serviceNode.Chains, serviceNode.PausedHeight, serviceNode.UnstakingHeight); err != nil {
			log.Fatalf("an error occurred inserting a service node in the genesis state: %s", err.Error())
		}
		if err = addValueToPool(genesis.ServiceNodeStakePoolName, serviceNode.StakedTokens); err != nil {
			log.Fatalf("an error occurred inserting staked tokens into %s pool", genesis.ServiceNodeStakePoolName)
		}
	}

	for _, fish := range state.Fishermen {
		if err = rwContext.InsertFisherman(fish.Address, fish.PublicKey, fish.Output, fish.Paused, int(fish.Status), fish.ServiceUrl, fish.StakedTokens, fish.Chains, fish.PausedHeight, fish.UnstakingHeight); err != nil {
			log.Fatalf("an error occurred inserting a fisherman in the genesis state: %s", err.Error())
		}
		if err = addValueToPool(genesis.FishermanStakePoolName, fish.StakedTokens); err != nil {
			log.Fatalf("an error occurred inserting staked tokens into %s pool", genesis.FishermanStakePoolName)
		}
	}

	for _, val := range state.Validators {
		if err = rwContext.InsertValidator(val.Address, val.PublicKey, val.Output, val.Paused, int(val.Status), val.ServiceUrl, val.StakedTokens, val.PausedHeight, val.UnstakingHeight); err != nil {
			log.Fatalf("an error occurred inserting a validator in the genesis state: %s", err.Error())
		}
		if err = addValueToPool(genesis.ValidatorStakePoolName, val.StakedTokens); err != nil {
			log.Fatalf("an error occurred inserting staked tokens into %s pool", genesis.ValidatorStakePoolName)
		}
	}

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
