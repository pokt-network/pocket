package persistence

import (
	"encoding/hex"
	"log"
	"math/big"

	"github.com/pokt-network/pocket/persistence/types"
	"github.com/pokt-network/pocket/runtime/genesis"
	"github.com/pokt-network/pocket/shared/converters"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
)

// CONSIDERATION: Should this return an error and let the caller decide if it should log a fatal error?
func (m *persistenceModule) populateGenesisState(state *genesis.GenesisState) {
	m.logger.Info().Msg("Populating genesis state...")

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
		m.logger.Fatal().Err(err).Msg("an error occurred creating the rwContext for the genesis state")
	}

	for _, acc := range state.GetAccounts() {
		addrBz, err := hex.DecodeString(acc.GetAddress())
		if err != nil {
			m.logger.Fatal().Err(err).Str("address", acc.GetAddress()).Msg("an error occurred converting address to bytes")
		}
		err = rwContext.SetAccountAmount(addrBz, acc.GetAmount())
		if err != nil {
			m.logger.Fatal().Err(err).Str("address", acc.GetAddress()).Msg("an error occurred inserting an acc in the genesis state")
		}
	}
	for _, pool := range state.GetPools() {
		err = rwContext.InsertPool(pool.GetAddress(), pool.GetAmount()) // pool.GetAddress() returns the pool's semantic name
		if err != nil {
			m.logger.Fatal().Err(err).Str("address", pool.GetAddress()).Msg("an error occurred inserting an pool in the genesis state")
		}
	}

	stakedActorsInsertConfigs := []struct {
		Name     string
		Getter   func() []*coreTypes.Actor
		InsertFn func(address, publicKey, output []byte, paused bool, status int32, serviceURL, stakedTokens string, chains []string, pausedHeight, unstakingHeight int64) error
		Pool     coreTypes.Pools
	}{
		{
			Name:     "app",
			Getter:   state.GetApplications,
			InsertFn: rwContext.InsertApp,
			Pool:     coreTypes.Pools_POOLS_APP_STAKE,
		},
		{
			Name:     "servicer",
			Getter:   state.GetServicers,
			InsertFn: rwContext.InsertServicer,
			Pool:     coreTypes.Pools_POOLS_SERVICE_NODE_STAKE,
		},
		{
			Name:     "fisherman",
			Getter:   state.GetFishermen,
			InsertFn: rwContext.InsertFisherman,
			Pool:     coreTypes.Pools_POOLS_FISHERMAN_STAKE,
		},
		{
			Name:   "validator",
			Getter: state.GetValidators,
			InsertFn: func(address, publicKey, output []byte, paused bool, status int32, serviceURL, stakedTokens string, chains []string, pausedHeight, unstakingHeight int64) error {
				return rwContext.InsertValidator(address, publicKey, output, paused, status, serviceURL, stakedTokens, pausedHeight, unstakingHeight)
			},
			Pool: coreTypes.Pools_POOLS_VALIDATOR_STAKE,
		},
	}

	for _, saic := range stakedActorsInsertConfigs {
		for _, act := range saic.Getter() {
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
			err = saic.InsertFn(addrBz, pubKeyBz, outputBz, false, StakedStatus, act.GetGenericParam(), act.GetStakedAmount(), act.GetChains(), act.GetPausedHeight(), act.GetUnstakingHeight())
			if err != nil {
				log.Fatalf("an error occurred inserting an %s in the genesis state: %s", saic.Name, err.Error())
			}
			if err = addValueToPool(saic.Pool.FriendlyName(), act.GetStakedAmount()); err != nil {
				log.Fatalf("an error occurred inserting staked tokens into %s pool: %s", saic.Pool, err.Error())
			}
		}
	}

	if err = rwContext.InitGenesisParams(state.Params); err != nil {
		log.Fatalf("an error occurred initializing params: %s", err.Error())
	}

	if err = rwContext.InitFlags(); err != nil { // TODO (Team) use flags from genesis file not hardcoded
		m.logger.Fatal().Err(err).Msg("an error occurred initializing flags")
	}

	// Updates all the merkle trees
	stateHash, err := rwContext.ComputeStateHash()
	if err != nil {
		m.logger.Fatal().Err(err).Msg("an error occurred updating the app hash during genesis")
	}
	m.logger.Info().Str("stateHash", stateHash).Msg("PopulateGenesisState - computed state hash")

	// This updates the DB, blockstore, and commits the genesis state.
	// Note that the `quorumCert for genesis` is nil.
	if err = rwContext.Commit(nil, nil); err != nil {
		m.logger.Fatal().Err(err).Msg("an error occurred committing the genesis state to the DB")
	}
}

// TODO (#399): All of the functions below following a structure similar to `GetAll<Actor>`
//
//		can easily be refactored and condensed into a single function using a generic type or a common
//	 interface.
func (p *PostgresContext) GetAllAccounts(height int64) (accs []*coreTypes.Account, err error) {
	ctx, tx := p.getCtxAndTx()
	rows, err := tx.Query(ctx, types.Account.GetAllQuery(height))
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		acc := new(coreTypes.Account)
		if err := rows.Scan(&acc.Address, &acc.Amount, &height); err != nil {
			return nil, err
		}
		accs = append(accs, acc)
	}
	return
}

// CLEANUP: Consolidate with GetAllAccounts.
func (p *PostgresContext) GetAllPools(height int64) (accs []*coreTypes.Account, err error) {
	ctx, tx := p.getCtxAndTx()
	rows, err := tx.Query(ctx, types.Pool.GetAllQuery(height))
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		pool := new(coreTypes.Account)
		if err := rows.Scan(&pool.Address, &pool.Amount, &height); err != nil {
			return nil, err
		}
		accs = append(accs, pool)
	}
	return
}
