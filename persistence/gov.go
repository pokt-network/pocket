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
	return GetParam[int](p, types.BlocksPerSessionParamName, height)
}

func (p PostgresContext) GetServiceNodesPerSessionAt(height int64) (int, error) {
	return GetParam[int](p, types.ServiceNodesPerSessionParamName, height)
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
	return GetParam[int](p, paramName, height)
}

func (p PostgresContext) GetStringParam(paramName string, height int64) (string, error) {
	return GetParam[string](p, paramName, height)
}

func (p PostgresContext) GetBytesParam(paramName string, height int64) (param []byte, err error) {
	return GetParam[[]byte](p, paramName, height)
}

func (p PostgresContext) SetParam(paramName string, value interface{}) error {
	switch t := value.(type) {
	case int:
		return SetParam(p, paramName, t)
	case int32:
		return SetParam(p, paramName, t)
	case int64:
		return SetParam(p, paramName, t)
	case []byte:
		return SetParam(p, paramName, t)
	case string:
		return SetParam(p, paramName, t)
	default:
		log.Fatalf("unhandled paramType %T for value %v", t, value)
	}
	return fmt.Errorf("how did you get here?")
}

func SetParam[T schema.SupportedParamTypes](p PostgresContext, paramName string, paramValue T) error {
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
	if _, err = tx.Exec(ctx, schema.SetParamQuery(paramName, paramValue, height)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func GetParam[T int | string | []byte](p PostgresContext, paramName string, height int64) (i T, err error) {
	ctx, conn, err := p.GetCtxAndConnection()
	if err != nil {
		return i, err
	}

	var stringVal string
	err = conn.QueryRow(ctx, schema.GetParamQuery(paramName, height)).Scan(&stringVal)
	switch tp := any(i).(type) {
	case int, int32, int64:
		iConv, errr := strconv.Atoi(stringVal)
		if errr != nil {
			err = errr
			return
		}
		return any(iConv).(T), err
	case string:
		return any(stringVal).(T), err
	case []byte:
		v, e := hex.DecodeString(stringVal)
		return any(v).(T), e

	default:
		log.Fatalf("unhandled type for paramValue %T", tp)
	}

	return
}

func (p PostgresContext) InitFlags() error {
	//TODO: not implemented
	return nil
}

func (p PostgresContext) GetFlag(flagName string, height int64) (bool, error) {
	//TODO: not implemented
	return false, nil
}

func (p PostgresContext) SetFlag(flagName string, value bool) error {
	//TODO: not implemented
	return nil
}
