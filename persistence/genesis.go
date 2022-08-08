package persistence

import (
	"encoding/hex"
	"log"

	"github.com/pokt-network/pocket/persistence/schema"
	"github.com/pokt-network/pocket/shared/types/genesis"
	"github.com/pokt-network/pocket/utility/types"
)

func (pm *persistenceModule) PopulateGenesisState(state *genesis.GenesisState) { // TODO (Andrew) genericize with actors interface once merged with #111
	log.Println("Populating genesis state...")
	rwContext, err := pm.NewRWContext(0)
	if err != nil {
		log.Fatalf("an error occurred creating the rwContext for the genesis state: %s", err.Error())
	}
	for _, acc := range state.Accounts {
		if err = rwContext.SetAccountAmount(acc.Address, acc.Amount); err != nil {
			log.Fatalf("an error occurred inserting an acc in the genesis state: %s", err.Error())
		}
	}
	for _, pool := range state.Pools {
		if err = rwContext.InsertPool(pool.Name, pool.Account.Address, pool.Account.Amount); err != nil {
			log.Fatalf("an error occurred inserting an pool in the genesis state: %s", err.Error())
		}
	}
	for _, act := range state.Apps {
		if err = rwContext.InsertApp(act.Address, act.PublicKey, act.Output, act.Paused, int(act.Status), act.MaxRelays, act.StakedTokens, act.Chains, act.PausedHeight, act.UnstakingHeight); err != nil {
			log.Fatalf("an error occurred inserting an app in the genesis state: %s", err.Error())
		}
	}
	for _, act := range state.ServiceNodes {
		if err = rwContext.InsertServiceNode(act.Address, act.PublicKey, act.Output, act.Paused, int(act.Status), act.ServiceUrl, act.StakedTokens, act.Chains, act.PausedHeight, act.UnstakingHeight); err != nil {
			log.Fatalf("an error occurred inserting a service node in the genesis state: %s", err.Error())
		}
	}
	for _, act := range state.Fishermen {
		if err = rwContext.InsertFisherman(act.Address, act.PublicKey, act.Output, act.Paused, int(act.Status), act.ServiceUrl, act.StakedTokens, act.Chains, act.PausedHeight, act.UnstakingHeight); err != nil {
			log.Fatalf("an error occurred inserting a fisherman in the genesis state: %s", err.Error())
		}
	}
	for _, act := range state.Validators {
		if err = rwContext.InsertValidator(act.Address, act.PublicKey, act.Output, act.Paused, int(act.Status), act.ServiceUrl, act.StakedTokens, act.PausedHeight, act.UnstakingHeight); err != nil {
			log.Fatalf("an error occurred inserting a validator in the genesis state: %s", err.Error())
		}
	}
	if err = rwContext.InitParams(); err != nil { // TODO (Team) use params from genesis file not hardcoded
		log.Fatalf("an error occurred initializing params: %s", err.Error())
	}
	if err = rwContext.Commit(); err != nil {
		log.Fatalf("an error occurred during commit() on genesis state %s ", err.Error())
	}
}

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
