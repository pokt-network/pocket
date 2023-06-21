package sql

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/pokt-network/pocket/persistence/indexer"
	ptypes "github.com/pokt-network/pocket/persistence/types"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
)

// actorTypeToSchmeaName maps an ActorType to a PostgreSQL schema name
var actorTypeToSchemaName = map[coreTypes.ActorType]ptypes.ProtocolActorSchema{
	coreTypes.ActorType_ACTOR_TYPE_APP:      ptypes.ApplicationActor,
	coreTypes.ActorType_ACTOR_TYPE_VAL:      ptypes.ValidatorActor,
	coreTypes.ActorType_ACTOR_TYPE_FISH:     ptypes.FishermanActor,
	coreTypes.ActorType_ACTOR_TYPE_SERVICER: ptypes.ServicerActor,
}

// GetActors is responsible for fetching the actors that have been updated at a given height.
func GetActors(
	pgtx pgx.Tx,
	actorType coreTypes.ActorType,
	height uint64,
) ([]*coreTypes.Actor, error) {
	actorSchema, ok := actorTypeToSchemaName[actorType]
	if !ok {
		return nil, fmt.Errorf("no schema found for actor type: %s", actorType)
	}

	// TECHDEBT(#813): Avoid this cast to int64
	query := actorSchema.GetUpdatedAtHeightQuery(int64(height))
	rows, err := pgtx.Query(context.TODO(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	addrs := make([][]byte, 0)
	for rows.Next() {
		var addr string
		if err := rows.Scan(&addr); err != nil {
			return nil, err
		}
		addrBz, err := hex.DecodeString(addr)
		if err != nil {
			return nil, err
		}
		addrs = append(addrs, addrBz)
	}

	actors := make([]*coreTypes.Actor, len(addrs))
	for i, addr := range addrs {
		// TECHDEBT(#813): Avoid this cast to int64
		actor, err := getActor(pgtx, actorSchema, addr, int64(height))
		if err != nil {
			return nil, err
		}
		actors[i] = actor
	}

	return actors, nil
}

// GetAccountsUpdated gets the AccountSchema accounts that have been updated at height
func GetAccountsUpdated(
	pgtx pgx.Tx,
	acctType ptypes.ProtocolAccountSchema,
	height uint64,
) ([]*coreTypes.Account, error) {
	accounts := []*coreTypes.Account{}

	// TECHDEBT(#813): Avoid this cast to int64
	query := acctType.GetAccountsUpdatedAtHeightQuery(int64(height))
	rows, err := pgtx.Query(context.TODO(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		acc := new(coreTypes.Account)
		if err := rows.Scan(&acc.Address, &acc.Amount); err != nil {
			return nil, err
		}
		accounts = append(accounts, acc)
	}

	return accounts, nil
}

// GetTransactions takes a transaction indexer and returns the transactions for the current height
func GetTransactions(txi indexer.TxIndexer, height uint64) ([]*coreTypes.IndexedTransaction, error) {
	// TECHDEBT(#813): Avoid this cast to int64
	indexedTxs, err := txi.GetByHeight(int64(height), false)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions by height: %w", err)
	}
	return indexedTxs, nil
}

// GetPools returns the pools updated at the given height
func GetPools(pgtx pgx.Tx, height uint64) ([]*coreTypes.Account, error) {
	pools, err := GetAccountsUpdated(pgtx, ptypes.Pool, height)
	if err != nil {
		return nil, fmt.Errorf("failed to get pools: %w", err)
	}
	return pools, nil
}

// GetAccounts returns the list of accounts updated at the provided height
func GetAccounts(pgtx pgx.Tx, height uint64) ([]*coreTypes.Account, error) {
	accounts, err := GetAccountsUpdated(pgtx, ptypes.Account, height)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %w", err)
	}
	return accounts, nil
}

// GetFlags returns the set of flags updated at the given height
func GetFlags(pgtx pgx.Tx, height uint64) ([]*coreTypes.Flag, error) {
	fields := "name,value,enabled"
	query := fmt.Sprintf("SELECT %s FROM %s WHERE height=%d ORDER BY name ASC", fields, ptypes.FlagsTableName, height)
	rows, err := pgtx.Query(context.TODO(), query)
	if err != nil {
		return nil, fmt.Errorf("failed to get flags: %w", err)
	}
	defer rows.Close()

	flagSlice := []*coreTypes.Flag{}
	for rows.Next() {
		flag := new(coreTypes.Flag)
		if err := rows.Scan(&flag.Name, &flag.Value, &flag.Enabled); err != nil {
			return nil, err
		}
		flag.Height = int64(height)
		flagSlice = append(flagSlice, flag)
	}

	return flagSlice, nil
}

// GetParams returns the set of params updated at the currented height
func GetParams(pgtx pgx.Tx, height uint64) ([]*coreTypes.Param, error) {
	fields := "name,value"
	query := fmt.Sprintf("SELECT %s FROM %s WHERE height=%d ORDER BY name ASC", fields, ptypes.ParamsTableName, height)
	rows, err := pgtx.Query(context.TODO(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var paramSlice []*coreTypes.Param
	for rows.Next() {
		param := new(coreTypes.Param)
		if err := rows.Scan(&param.Name, &param.Value); err != nil {
			return nil, err
		}
		param.Height = int64(height)
		paramSlice = append(paramSlice, param)
	}

	return paramSlice, nil
}

// GetIBCStoreUpdates returns the set of key-value pairs updated at the current height for the IBC store
func GetIBCStoreUpdates(pgtx pgx.Tx, height uint64) (keys, values [][]byte, err error) {
	fields := "key,value"
	query := fmt.Sprintf("SELECT %s FROM %s WHERE height=%d ORDER BY key ASC", fields, ptypes.IbcStoreTableName, height)
	rows, err := pgtx.Query(context.TODO(), query)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var hexKey, hexValue string
	for rows.Next() {
		if err := rows.Scan(&hexKey, &hexValue); err != nil {
			return nil, nil, err
		}
		key, err := hex.DecodeString(hexKey)
		if err != nil {
			return nil, nil, err
		}
		value, err := hex.DecodeString(hexValue)
		if err != nil {
			return nil, nil, err
		}
		keys = append(keys, key)
		values = append(values, value)
	}

	return keys, values, nil
}

func getActor(tx pgx.Tx, actorSchema ptypes.ProtocolActorSchema, address []byte, height int64) (actor *coreTypes.Actor, err error) {
	ctx := context.TODO()
	actor, height, err = getActorFromRow(actorSchema.GetActorType(), tx.QueryRow(ctx, actorSchema.GetQuery(hex.EncodeToString(address), height)))
	if err != nil {
		return
	}
	return getChainsForActor(ctx, tx, actorSchema, actor, height)
}

func getActorFromRow(actorType coreTypes.ActorType, row pgx.Row) (actor *coreTypes.Actor, height int64, err error) {
	actor = &coreTypes.Actor{
		ActorType: actorType,
	}
	err = row.Scan(
		&actor.Address,
		&actor.PublicKey,
		&actor.StakedAmount,
		&actor.ServiceUrl,
		&actor.Output,
		&actor.PausedHeight,
		&actor.UnstakingHeight,
		&height)
	return
}

func getChainsForActor(
	ctx context.Context,
	tx pgx.Tx,
	actorSchema ptypes.ProtocolActorSchema,
	actor *coreTypes.Actor,
	height int64,
) (a *coreTypes.Actor, err error) {
	if actorSchema.GetChainsTableName() == "" {
		return actor, nil
	}
	rows, err := tx.Query(ctx, actorSchema.GetChainsQuery(actor.Address, height))
	if err != nil {
		return actor, err
	}
	defer rows.Close()

	var chainAddr string
	var chainID string
	var chainEndHeight int64
	for rows.Next() {
		err = rows.Scan(&chainAddr, &chainID, &chainEndHeight)
		if err != nil {
			return
		}
		if chainAddr != actor.Address {
			return actor, fmt.Errorf("unexpected address %s, expected %s when reading chains", chainAddr, actor.Address)
		}
		actor.Chains = append(actor.Chains, chainID)
	}
	return actor, nil
}
