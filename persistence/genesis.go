package persistence

import (
	"encoding/hex"
	"log"
	"math/big"

	"github.com/pokt-network/pocket/persistence/types"
	"github.com/pokt-network/pocket/shared/converters"
	"github.com/pokt-network/pocket/shared/modules"
)

// CONSIDERATION: Should this return an error and let the caller decide if it should log a fatal error?
func (m *persistenceModule) populateGenesisState(state modules.PersistenceGenesisState) {
	log.Println("Populating genesis state...")

	// REFACTOR: This business logic should probably live in `types/genesis.go`
	//           and we need to add proper unit tests for it.`
	poolValues := make(map[string]*big.Int, 0)
	addValueToPool := func(poolName string, valueToAdd string) error {
		value, err := converters.StringToBigInt(valueToAdd)
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
		if err = addValueToPool(types.PoolNames_AppStakePool.String(), act.GetStakedAmount()); err != nil {
			log.Fatalf("an error occurred inserting staked tokens into %s pool: %s", types.PoolNames_AppStakePool, err.Error())
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
		if err = addValueToPool(types.PoolNames_ServiceNodeStakePool.String(), act.GetStakedAmount()); err != nil {
			log.Fatalf("an error occurred inserting staked tokens into %s pool: %s", types.PoolNames_ServiceNodeStakePool.String(), err.Error())
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
		if err = addValueToPool(types.PoolNames_FishermanStakePool.String(), act.GetStakedAmount()); err != nil {
			log.Fatalf("an error occurred inserting staked tokens into %s pool: %s", types.PoolNames_FishermanStakePool.String(), err.Error())
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
		if err = addValueToPool(types.PoolNames_ValidatorStakePool.String(), act.GetStakedAmount()); err != nil {
			log.Fatalf("an error occurred inserting staked tokens into %s pool: %s", types.PoolNames_ValidatorStakePool.String(), err.Error())
		}
	}
	// TODO(team): use params from genesis file - not the hardcoded
	if err = rwContext.InitParams(); err != nil {
		log.Fatalf("an error occurred initializing params: %s", err.Error())
	}

	if err = rwContext.InitFlags(); err != nil { // TODO (Team) use flags from genesis file not hardcoded
		log.Fatalf("an error occurred initializing flags: %s", err.Error())
	}

	// Updates all the merkle trees
	appHash, err := rwContext.ComputeAppHash()
	if err != nil {
		log.Fatalf("an error occurred updating the app hash during genesis: %s", err.Error())
	}

	if err := rwContext.SetProposalBlock(hex.EncodeToString(appHash), nil, nil, nil); err != nil {
		log.Fatalf("an error occurred setting the proposal block during genesis: %s", err.Error())
	}

	// This update the DB, blockstore, and commits the state
	if err = rwContext.Commit(nil); err != nil {
		log.Fatalf("error committing genesis state to DB %s ", err.Error())
	}
}

// TODO (#399): All of the functions below following a structure similar to `GetAll<Actor>`
//	can easily be refactored and condensed into a single function using a generic type or a common
//  interface.
func (p PostgresContext) GetAllAccounts(height int64) (accs []modules.Account, err error) {
	ctx, tx, err := p.getCtxAndTx()
	if err != nil {
		return nil, err
	}
	rows, err := tx.Query(ctx, types.SelectAccounts(height, types.AccountTableName))
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
	ctx, tx, err := p.getCtxAndTx()
	if err != nil {
		return nil, err
	}
	rows, err := tx.Query(ctx, types.SelectPools(height, types.PoolTableName))
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
