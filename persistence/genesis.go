package persistence

import (
	"encoding/hex"
	"log"
	"math/big"

	"github.com/pokt-network/pocket/persistence/types"
	"github.com/pokt-network/pocket/shared/modules"
)

// TODO(andrew): generalize with the `actors interface`

// WARNING: This function crashes the process if there is an error populating the genesis state.
func (m *PersistenceModule) populateGenesisState(state *types.PersistenceGenesisState) {
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

	if err != nil {
		log.Fatalf("an error occurred creating the rwContext for the genesis state: %s", err.Error())
	}
	for _, acc := range state.GetAccs() {
		addrBz, err := hex.DecodeString(acc.GetAddress())
		if err != nil {
			log.Fatalf("an error occurred converting address to bytes %s", acc.GetAddress())
		}
		err = rwContext.SetAccountAmount(addrBz, acc.GetAmount())
		if err != nil {
			log.Fatalf("an error occurred inserting an acc in the genesis state: %s", err.Error())
		}
	}
	for _, pool := range state.GetAccPools() {
		poolNameBytes := []byte(pool.GetAddress())
		err = rwContext.InsertPool(pool.GetAddress(), poolNameBytes, pool.GetAmount())
		if err != nil {
			log.Fatalf("an error occurred inserting an pool in the genesis state: %s", err.Error())
		}
	}
	for _, act := range state.GetApps() { // TODO (Andrew) genericize the genesis population logic for actors #149
		addrBz, err := hex.DecodeString(act.GetAddress())
		if err != nil {
			log.Fatalf("an error occurred converting address to bytes %s", act.GetAddress())
		}
		pubKeyBz, err := hex.DecodeString(act.GetPublicKey())
		if err != nil {
			log.Fatalf("an error occurred converting pubKey to bytes %s", act.GetPublicKey())
		}
		outputBz, err := hex.DecodeString(act.GetOutput())
		if err != nil {
			log.Fatalf("an error occurred converting output to bytes %s", act.GetOutput())
		}
		err = rwContext.InsertApp(addrBz, pubKeyBz, outputBz, false, StakedStatus, act.GetGenericParam(), act.GetStakedAmount(), act.GetChains(), act.GetPausedHeight(), act.GetUnstakingHeight())
		if err != nil {
			log.Fatalf("an error occurred inserting an app in the genesis state: %s", err.Error())
		}
		if err = addValueToPool(types.Pool_Names_AppStakePool.String(), act.GetStakedAmount()); err != nil {
			log.Fatalf("an error occurred inserting staked tokens into %s pool: %s", types.Pool_Names_AppStakePool, err.Error())
		}
	}
	for _, act := range state.GetNodes() {
		addrBz, err := hex.DecodeString(act.GetAddress())
		if err != nil {
			log.Fatalf("an error occurred converting address to bytes %s", act.GetAddress())
		}
		pubKeyBz, err := hex.DecodeString(act.GetPublicKey())
		if err != nil {
			log.Fatalf("an error occurred converting pubKey to bytes %s", act.GetPublicKey())
		}
		outputBz, err := hex.DecodeString(act.GetOutput())
		if err != nil {
			log.Fatalf("an error occurred converting output to bytes %s", act.GetOutput())
		}
		err = rwContext.InsertServiceNode(addrBz, pubKeyBz, outputBz, false, StakedStatus, act.GetGenericParam(), act.GetStakedAmount(), act.GetChains(), act.GetPausedHeight(), act.GetUnstakingHeight())
		if err != nil {
			log.Fatalf("an error occurred inserting a service node in the genesis state: %s", err.Error())
		}
		if err = addValueToPool(types.Pool_Names_ServiceNodeStakePool.String(), act.GetStakedAmount()); err != nil {
			log.Fatalf("an error occurred inserting staked tokens into %s pool: %s", types.Pool_Names_ServiceNodeStakePool.String(), err.Error())
		}
	}
	for _, act := range state.GetFish() {
		addrBz, err := hex.DecodeString(act.GetAddress())
		if err != nil {
			log.Fatalf("an error occurred converting address to bytes %s", act.GetAddress())
		}
		pubKeyBz, err := hex.DecodeString(act.GetPublicKey())
		if err != nil {
			log.Fatalf("an error occurred converting pubKey to bytes %s", act.GetPublicKey())
		}
		outputBz, err := hex.DecodeString(act.GetOutput())
		if err != nil {
			log.Fatalf("an error occurred converting output to bytes %s", act.GetOutput())
		}
		err = rwContext.InsertFisherman(addrBz, pubKeyBz, outputBz, false, StakedStatus, act.GetGenericParam(), act.GetStakedAmount(), act.GetChains(), act.GetPausedHeight(), act.GetUnstakingHeight())
		if err != nil {
			log.Fatalf("an error occurred inserting a fisherman in the genesis state: %s", err.Error())
		}
		if err = addValueToPool(types.Pool_Names_FishermanStakePool.String(), act.GetStakedAmount()); err != nil {
			log.Fatalf("an error occurred inserting staked tokens into %s pool: %s", types.Pool_Names_FishermanStakePool.String(), err.Error())
		}
	}
	for _, act := range state.GetVals() {
		addrBz, err := hex.DecodeString(act.GetAddress())
		if err != nil {
			log.Fatalf("an error occurred converting address to bytes %s", act.GetAddress())
		}
		pubKeyBz, err := hex.DecodeString(act.GetPublicKey())
		if err != nil {
			log.Fatalf("an error occurred converting pubKey to bytes %s", act.GetPublicKey())
		}
		outputBz, err := hex.DecodeString(act.GetOutput())
		if err != nil {
			log.Fatalf("an error occurred converting output to bytes %s", act.GetOutput())
		}
		err = rwContext.InsertValidator(addrBz, pubKeyBz, outputBz, false, StakedStatus, act.GetGenericParam(), act.GetStakedAmount(), act.GetPausedHeight(), act.GetUnstakingHeight())
		if err != nil {
			log.Fatalf("an error occurred inserting a validator in the genesis state: %s", err.Error())
		}
		if err = addValueToPool(types.Pool_Names_ValidatorStakePool.String(), act.GetStakedAmount()); err != nil {
			log.Fatalf("an error occurred inserting staked tokens into %s pool: %s", types.Pool_Names_ValidatorStakePool.String(), err.Error())
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
func (p PostgresContext) GetAllAccounts(height int64) (accs []modules.Account, err error) {
	ctx, txn, err := p.GetCtxAndTxn()
	if err != nil {
		return nil, err
	}
	rows, err := txn.Query(ctx, types.SelectAccounts(height, types.AccountTableName))
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		acc := new(types.Account)
		if err = rows.Scan(&acc.Address, &acc.Amount, &height); err != nil {
			return nil, err
		}
		// acc.Address, err = address
		if err != nil {
			return nil, err
		}
		accs = append(accs, acc)
	}
	return
}

// CLEANUP: Consolidate with GetAllAccounts.
func (p PostgresContext) GetAllPools(height int64) (accs []modules.Account, err error) {
	ctx, txn, err := p.GetCtxAndTxn()
	if err != nil {
		return nil, err
	}
	rows, err := txn.Query(ctx, types.SelectPools(height, types.PoolTableName))
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		pool := new(types.Account)
		if err = rows.Scan(&pool.Address, &pool.Amount, &height); err != nil {
			return nil, err
		}
		accs = append(accs, pool)
	}
	return
}

func (p PostgresContext) GetAllApps(height int64) (apps []modules.Actor, err error) {
	ctx, txn, err := p.GetCtxAndTxn()
	if err != nil {
		return nil, err
	}
	rows, err := txn.Query(ctx, types.ApplicationActor.GetAllQuery(height))
	if err != nil {
		return nil, err
	}
	var actors []types.BaseActor
	for rows.Next() {
		var actor types.BaseActor
		actor, height, err = p.GetActorFromRow(rows)
		if err != nil {
			return
		}
		actors = append(actors, actor)
	}
	rows.Close()
	for _, actor := range actors {
		actor, err = p.GetChainsForActor(ctx, txn, types.ApplicationActor, actor, height)
		if err != nil {
			return
		}
		apps = append(apps, p.BaseActorToActor(actor, types.ActorType_App))
	}
	return
}

func (p PostgresContext) GetAllValidators(height int64) (vals []modules.Actor, err error) {
	ctx, txn, err := p.GetCtxAndTxn()
	if err != nil {
		return nil, err
	}
	rows, err := txn.Query(ctx, types.ValidatorActor.GetAllQuery(height))
	if err != nil {
		return nil, err
	}
	var actors []types.BaseActor
	for rows.Next() {
		var actor types.BaseActor
		actor, height, err = p.GetActorFromRow(rows)
		if err != nil {
			return
		}
		actors = append(actors, actor)
	}
	rows.Close()
	for _, actor := range actors {
		actor, err = p.GetChainsForActor(ctx, txn, types.ApplicationActor, actor, height)
		if err != nil {
			return
		}
		vals = append(vals, p.BaseActorToActor(actor, types.ActorType_Val))
	}
	return
}

func (p PostgresContext) GetAllServiceNodes(height int64) (sn []modules.Actor, err error) {
	ctx, txn, err := p.GetCtxAndTxn()
	if err != nil {
		return nil, err
	}
	rows, err := txn.Query(ctx, types.ServiceNodeActor.GetAllQuery(height))
	if err != nil {
		return nil, err
	}
	var actors []types.BaseActor
	for rows.Next() {
		var actor types.BaseActor
		actor, height, err = p.GetActorFromRow(rows)
		if err != nil {
			return
		}
		actors = append(actors, actor)
	}
	rows.Close()
	for _, actor := range actors {
		actor, err = p.GetChainsForActor(ctx, txn, types.ServiceNodeActor, actor, height)
		if err != nil {
			return
		}
		sn = append(sn, p.BaseActorToActor(actor, types.ActorType_Node))
	}
	return
}

func (p PostgresContext) GetAllFishermen(height int64) (f []modules.Actor, err error) {
	ctx, txn, err := p.GetCtxAndTxn()
	if err != nil {
		return nil, err
	}
	rows, err := txn.Query(ctx, types.FishermanActor.GetAllQuery(height))
	if err != nil {
		return nil, err
	}
	var actors []types.BaseActor
	for rows.Next() {
		var actor types.BaseActor
		actor, height, err = p.GetActorFromRow(rows)
		if err != nil {
			return
		}
		actors = append(actors, actor)
	}
	rows.Close()
	for _, actor := range actors {
		actor, err = p.GetChainsForActor(ctx, txn, types.FishermanActor, actor, height)
		if err != nil {
			return
		}
		f = append(f, p.BaseActorToActor(actor, types.ActorType_Fish))
	}
	return
}

// TODO (Team) deprecate with interface #163 <Bumped to #149> as #163 is getting large
func (p PostgresContext) BaseActorToActor(ba types.BaseActor, actorType types.ActorType) *types.Actor {
	actor := new(types.Actor)
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
