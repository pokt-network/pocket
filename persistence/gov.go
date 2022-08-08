package persistence

import (
	"encoding/hex"
	"fmt"
	"log"
	"strconv"

	"github.com/jackc/pgx/v4"
	"github.com/pokt-network/pocket/persistence/schema"
	"github.com/pokt-network/pocket/shared/types"
	"github.com/pokt-network/pocket/shared/types/genesis"
)

// TODO(https://github.com/pokt-network/pocket/issues/76): Optimize gov parameters implementation & schema.

func (p PostgresContext) GetBlocksPerSession(height int64) (int, error) {
	return p.GetIntParam(types.BlocksPerSessionParamName, height)
}

func (p PostgresContext) GetServiceNodesPerSessionAt(height int64) (int, error) {
	return p.GetIntParam(types.ServiceNodesPerSessionParamName, height)
}

func (p PostgresContext) InitParams() error {
	ctx, conn, err := p.GetCtxAndConnection()
	if err != nil {
		return err
	}
	_, err = conn.Exec(ctx, schema.InsertParams(genesis.DefaultParams(), p.Height))
	return err
}

func (p PostgresContext) GetIntParam(paramName string, height int64) (int, error) {
	v, _, err := getParamOrFlag[int](p, schema.ParamsTableName, paramName, height)
	return v, err
}

func (p PostgresContext) GetStringParam(paramName string, height int64) (string, error) {
	v, _, err := getParamOrFlag[string](p, schema.ParamsTableName, paramName, height)
	return v, err
}

func (p PostgresContext) GetBytesParam(paramName string, height int64) (param []byte, err error) {
	v, _, err := getParamOrFlag[[]byte](p, schema.ParamsTableName, paramName, height)
	return v, err
}

func (p PostgresContext) SetParam(paramName string, value any) error {
	return p.setParamOrFlag(paramName, value, nil)
}

func (p PostgresContext) InitFlags() error {
	// TODO: not implemented
	return nil
}

func (p PostgresContext) GetIntFlag(flagName string, height int64) (value int, enabled bool, err error) {
	return getParamOrFlag[int](p, schema.FlagsTableName, flagName, height)
}

func (p PostgresContext) GetStringFlag(flagName string, height int64) (value string, enabled bool, err error) {
	return getParamOrFlag[string](p, schema.FlagsTableName, flagName, height)
}

func (p PostgresContext) GetBytesFlag(flagName string, height int64) (value []byte, enabled bool, err error) {
	return getParamOrFlag[[]byte](p, schema.FlagsTableName, flagName, height)
}

func (p PostgresContext) SetFlag(flagName string, value any, enabled bool) error {
	return p.setParamOrFlag(flagName, value, &enabled)
}

// setParamOrFlag simply wraps the call to the generic function with the supplied underlying type
func (p PostgresContext) setParamOrFlag(name string, value any, enabled *bool) error {
	switch t := value.(type) {
	case int:
		return setParamOrFlag(p, name, t, enabled)
	case int32:
		return setParamOrFlag(p, name, t, enabled)
	case int64:
		return setParamOrFlag(p, name, t, enabled)
	case []byte:
		return setParamOrFlag(p, name, t, enabled)
	case string:
		return setParamOrFlag(p, name, t, enabled)
	default:
		log.Fatalf("unhandled paramType %T for value %v", t, value)
	}
	return fmt.Errorf("how did you get here?")
}

// setParamOrFlag sets a param or a flag.
// If `enabled` is nil, we are dealing with a param, otherwise it's a flag
func setParamOrFlag[T schema.SupportedParamTypes](p PostgresContext, paramName string, paramValue T, enabled *bool) error {
	ctx, conn, err := p.GetCtxAndConnection()
	if err != nil {
		return err
	}
	height, err := p.GetHeight()
	if err != nil {
		return err
	}
	tx, err := conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	tableName := schema.ParamsTableName
	if enabled != nil {
		tableName = schema.FlagsTableName
	}

	if _, err = tx.Exec(ctx, schema.InsertParamOrFlag(tableName, paramName, height, paramValue, enabled)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func getParamOrFlag[T int | string | []byte](p PostgresContext, tableName, paramName string, height int64) (i T, enabled bool, err error) {
	ctx, conn, err := p.GetCtxAndConnection()
	if err != nil {
		return i, enabled, err
	}

	var stringVal string
	row := conn.QueryRow(ctx, schema.GetParamOrFlagQuery(tableName, paramName, height))
	if tableName == schema.ParamsTableName {
		err = row.Scan(&stringVal)
	} else {
		err = row.Scan(&stringVal, &enabled)
	}
	if err != nil {
		return
	}
	switch tp := any(i).(type) {
	case int, int32, int64:
		iConv, errr := strconv.Atoi(stringVal)
		if errr != nil {
			err = errr
			return
		}
		return any(iConv).(T), enabled, err
	case string:
		return any(stringVal).(T), enabled, err
	case []byte:
		v, e := hex.DecodeString(stringVal)
		return any(v).(T), enabled, e

	default:
		log.Fatalf("unhandled type for paramValue %T", tp)
	}

	return
}
