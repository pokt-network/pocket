package persistence

import (
	"context"
	"fmt"
	"math/big"
	"math/rand"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/pokt-network/pocket/persistence/kvstore"
	"github.com/pokt-network/pocket/persistence/schema"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/types"
	"github.com/pokt-network/pocket/shared/types/genesis"
)

const (
	CreateSchemaIfNotExists = "CREATE SCHEMA IF NOT EXISTS"
	SetSearchPathTo         = "SET search_path TO"
	CreateTableIfNotExists  = "CREATE TABLE IF NOT EXISTS"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var _ modules.PersistenceContext = &PostgresContext{}

// TODO: These are only externalized for testing purposes, so they should be made private and
// it is trivial to create a helper to initial a context with some values.
type PostgresContext struct {
	Height       int64
	PostgresDB   *pgx.Conn
	BlockStore   kvstore.KVStore
	ContextStore kvstore.KVStore
}

func (pg *PostgresContext) GetCtxAndConnection() (context.Context, *pgx.Conn, error) {
	conn, err := pg.GetConnection()
	if err != nil {
		return nil, nil, err
	}
	ctx, err := pg.GetContext()
	if err != nil {
		return nil, nil, err
	}
	return ctx, conn, nil
}

func (pg *PostgresContext) GetConnection() (*pgx.Conn, error) {
	return pg.PostgresDB, nil
}

func (pg *PostgresContext) GetContext() (context.Context, error) {
	return context.TODO(), nil
}

var protocolActorSchemas = []schema.ProtocolActorSchema{
	schema.ApplicationActor,
	schema.FishermanActor,
	schema.ServiceNodeActor,
	schema.ValidatorActor,
}

// TODO(pokt-network/pocket/issues/77): Enable proper up and down migrations
// TODO: Split `connect` and `initialize` into two separate compnents
func ConnectAndInitializeDatabase(postgresUrl string, schema string) (*pgx.Conn, error) {
	ctx := context.TODO()

	// Connect to the DB
	db, err := pgx.Connect(context.Background(), postgresUrl)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}

	// Creating and setting a new schema so we can running multiple nodes on one postgres instance. See
	// more details at https://github.com/go-pg/pg/issues/351.
	if _, err = db.Exec(ctx, fmt.Sprintf("%s %s", CreateSchemaIfNotExists, schema)); err != nil {
		return nil, err
	}

	if _, err = db.Exec(ctx, fmt.Sprintf("%s %s", SetSearchPathTo, schema)); err != nil {
		return nil, err
	}

	if err := initializeAllTables(ctx, db); err != nil {
		return nil, fmt.Errorf("unable to initialize tables: %v", err)
	}

	return db, nil

}

// TODO(pokt-network/pocket/issues/77): Delete all the `initializeAllTables` calls once proper migrations are implemented.
func initializeAllTables(ctx context.Context, db *pgx.Conn) error {
	if err := initializeAccountTables(ctx, db); err != nil {
		return err
	}

	if err := initializeGovTables(ctx, db); err != nil {
		return err
	}

	for _, actor := range protocolActorSchemas {
		if err := initializeProtocolActorTables(ctx, db, actor); err != nil {
			return err
		}
	}

	return nil
}

func initializeProtocolActorTables(ctx context.Context, db *pgx.Conn, actor schema.ProtocolActorSchema) error {
	if _, err := db.Exec(ctx, fmt.Sprintf(`%s %s %s`, CreateTableIfNotExists, actor.GetTableName(), actor.GetTableSchema())); err != nil {
		return err
	}
	if actor.GetChainsTableName() != "" {
		if _, err := db.Exec(ctx, fmt.Sprintf(`%s %s %s`, CreateTableIfNotExists, actor.GetChainsTableName(), actor.GetChainsTableSchema())); err != nil {
			return err
		}
	}
	return nil
}

func initializeAccountTables(ctx context.Context, db *pgx.Conn) error {
	if _, err := db.Exec(ctx, fmt.Sprintf(`%s %s %s`, CreateTableIfNotExists, schema.AccountTableName, schema.AccountTableSchema)); err != nil {
		return err
	}
	if _, err := db.Exec(ctx, fmt.Sprintf(`%s %s %s`, CreateTableIfNotExists, schema.PoolTableName, schema.PoolTableSchema)); err != nil {
		return err
	}
	return nil
}

func initializeGovTables(ctx context.Context, db *pgx.Conn) error {
	_, err := db.Exec(ctx, fmt.Sprintf(`%s %s %s`, CreateTableIfNotExists, schema.ParamsTableName, schema.ParamsTableSchema))
	if err != nil {
		return err
	}
	return nil
}

func (m *persistenceModule) hydrateGenesisDbState() error {
	state := m.GetBus().GetConfig().GenesisSource.GetState()
	if nil == state {
		return fmt.Errorf("unable to hydrate genesis DB state because genesis source is misconfigured")
	}

	ctx, err := m.NewContext(0)
	if err != nil {
		return err
	}
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

	for _, v := range state.Validators {
		// DISCUSS_IN_THIS_COMMIT: If the validator (and other actors) have a balance, do we expect that
		// to be in the `accounts` type corresponding to the same address?
		if err := ctx.SetAccountAmount(v.Address, "0"); err != nil {
			return err
		}
		if err := ctx.InsertValidator(v.Address, v.PublicKey, v.Output, v.Paused, int(v.Status), v.ServiceUrl, v.StakedTokens, v.PausedHeight, v.UnstakingHeight); err != nil {
			return err
		}
		if err := addValueToPool(genesis.ValidatorStakePoolName, v.StakedTokens); err != nil {
			return err
		}
	}

	for _, f := range state.Fishermen {
		if err := ctx.InsertFisherman(f.Address, f.PublicKey, f.Output, f.Paused, int(f.Status), f.ServiceUrl, f.StakedTokens, f.Chains, f.PausedHeight, f.UnstakingHeight); err != nil {
			return err
		}
		if err := ctx.SetAccountAmount(f.Address, "0"); err != nil {
			return err
		}
		if err := addValueToPool(genesis.FishermanStakePoolName, f.StakedTokens); err != nil {
			return err
		}
	}

	for _, sn := range state.ServiceNodes {
		if err := ctx.InsertServiceNode(sn.Address, sn.PublicKey, sn.Output, sn.Paused, int(sn.Status), sn.ServiceUrl, sn.StakedTokens, sn.Chains, sn.PausedHeight, sn.UnstakingHeight); err != nil {
			return err
		}
		if err := ctx.SetAccountAmount(sn.Address, "0"); err != nil {
			return err
		}
		if err := addValueToPool(genesis.ServiceNodeStakePoolName, sn.StakedTokens); err != nil {
			return err
		}
	}

	for _, app := range state.Apps {
		if err := ctx.InsertApp(app.Address, app.PublicKey, app.Output, app.Paused, int(app.Status), app.MaxRelays, app.StakedTokens, app.Chains, app.PausedHeight, app.UnstakingHeight); err != nil {
			return err
		}
		if err := ctx.SetAccountAmount(app.Address, "0"); err != nil {
			return err
		}
		if err := addValueToPool(genesis.AppStakePoolName, app.StakedTokens); err != nil {
			return err
		}
	}

	for _, acc := range state.Accounts {
		if err := ctx.AddAccountAmount(acc.Address, acc.Amount); err != nil {
			return err
		}
	}

	for _, pool := range state.Pools {
		if err := ctx.InsertPool(pool.Name, pool.Account.Address, pool.Account.Amount); err != nil {
			return err
		}
		// DISCUSS: What if there's a discrepency between `pool.Account.Amount` and `poolValues`?
		if err := ctx.SetAccountAmount(pool.Account.Address, pool.Account.Amount); err != nil {
			return err
		}
	}

	if err := ctx.InitParams(); err != nil {
		return err
	}
	// DISCUSS_IN_THIS_COMMIT: Calling this for each param is not practical. Need someone to refactor the gov paramstable.
	if err := ctx.SetValidatorMaximumMissedBlocks(int(state.Params.ValidatorMaximumMissedBlocks)); err != nil {
		return err
	}

	return nil
}

// Exposed for debugging purposes only
func (p PostgresContext) ClearAllDebug() error {
	ctx, conn, err := p.GetCtxAndConnection()
	if err != nil {
		return err
	}

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	for _, actor := range protocolActorSchemas {
		if _, err = tx.Exec(ctx, actor.ClearAllQuery()); err != nil {
			return err
		}
		if actor.GetChainsTableName() != "" {
			if _, err = tx.Exec(ctx, actor.ClearAllChainsQuery()); err != nil {
				return err
			}
		}
	}

	tx.Exec(ctx, schema.ClearAllGovQuery())

	return nil
}
