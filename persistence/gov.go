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

func (p PostgresContext) InitFlags() error {
	//TODO: not implemented
	return nil
}

func (p PostgresContext) GetIntFlag(paramName string, height int64) (value int, enabled bool, err error) {
	return GetFlag[int](p, paramName, height)
}

func (p PostgresContext) GetStringFlag(paramName string, height int64) (value string, enabled bool, err error) {
	return GetFlag[string](p, paramName, height)
}

func (p PostgresContext) GetBytesFlag(paramName string, height int64) (value []byte, enabled bool, err error) {
	return GetFlag[[]byte](p, paramName, height)
}

func (p PostgresContext) SetFlag(paramName string, value interface{}, enabled bool) error {
	switch t := value.(type) {
	case int:
		return SetFlag(p, paramName, t, enabled)
	case int32:
		return SetFlag(p, paramName, t, enabled)
	case int64:
		return SetFlag(p, paramName, t, enabled)
	case []byte:
		return SetFlag(p, paramName, t, enabled)
	case string:
		return SetFlag(p, paramName, t, enabled)
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

func SetFlag[T schema.SupportedParamTypes](p PostgresContext, paramName string, paramValue T, enabled bool) error {
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
	if _, err = tx.Exec(ctx, schema.SetFlagQuery(paramName, paramValue, enabled, height)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func GetFlag[T int | string | []byte](p PostgresContext, paramName string, height int64) (i T, enabled bool, err error) {
	ctx, conn, err := p.GetCtxAndConnection()
	if err != nil {
		return i, enabled, err
	}

	var stringVal string
	err = conn.QueryRow(ctx, schema.GetFlagQuery(paramName, height)).Scan(&stringVal, &enabled)
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
