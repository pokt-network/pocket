package persistence

import (
	"encoding/hex"
	"fmt"
	"log"
	"strconv"

	"github.com/pokt-network/pocket/persistence/types"
	"github.com/pokt-network/pocket/runtime/genesis"
)

// TODO : Deprecate these two constants when we change the persistenceRWContext interface to pass the `paramName`
const (
	BlocksPerSessionParamName       = "blocks_per_session"
	ServiceNodesPerSessionParamName = "service_nodes_per_session"
)

// TODO (Team) BUG setting parameters twice on the same height causes issues. We need to move the schema away from 'end_height' and
// more towards the height_constraint architecture

func (p PostgresContext) GetBlocksPerSession(height int64) (int, error) {
	return p.GetIntParam(BlocksPerSessionParamName, height)
}

func (p PostgresContext) GetServiceNodesPerSessionAt(height int64) (int, error) {
	return p.GetIntParam(ServiceNodesPerSessionParamName, height)
}

func (p PostgresContext) InitGenesisParams(params *genesis.Params) error {
	ctx, tx, err := p.getCtxAndTx()
	if err != nil {
		return err
	}
	if p.Height != 0 {
		return fmt.Errorf("cannot initialize params at height %d", p.Height)
	}
	_, err = tx.Exec(ctx, types.InsertParams(params, p.Height))
	return err
}

func (p PostgresContext) GetIntParam(paramName string, height int64) (int, error) {
	v, _, err := getParamOrFlag[int](p, types.ParamsTableName, paramName, height)
	return v, err
}

func (p PostgresContext) GetStringParam(paramName string, height int64) (string, error) {
	v, _, err := getParamOrFlag[string](p, types.ParamsTableName, paramName, height)
	return v, err
}

func (p PostgresContext) GetBytesParam(paramName string, height int64) (param []byte, err error) {
	v, _, err := getParamOrFlag[[]byte](p, types.ParamsTableName, paramName, height)
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
	return getParamOrFlag[int](p, types.FlagsTableName, flagName, height)
}

func (p PostgresContext) GetStringFlag(flagName string, height int64) (value string, enabled bool, err error) {
	return getParamOrFlag[string](p, types.FlagsTableName, flagName, height)
}

func (p PostgresContext) GetBytesFlag(flagName string, height int64) (value []byte, enabled bool, err error) {
	return getParamOrFlag[[]byte](p, types.FlagsTableName, flagName, height)
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
		break
	}
	return fmt.Errorf("unhandled paramType %T for value %v", value, value)
}

// setParamOrFlag sets a param or a flag.
// If `enabled` is nil, we are dealing with a param, otherwise it's a flag
func setParamOrFlag[T types.SupportedParamTypes](p PostgresContext, paramName string, paramValue T, enabled *bool) error {
	ctx, tx, err := p.getCtxAndTx()
	if err != nil {
		return err
	}
	height, err := p.GetHeight()
	if err != nil {
		return err
	}
	tableName := types.ParamsTableName
	if enabled != nil {
		tableName = types.FlagsTableName
	}
	if _, err = tx.Exec(ctx, types.InsertParamOrFlag(tableName, paramName, height, paramValue, enabled)); err != nil {
		return err
	}
	return nil
}

func getParamOrFlag[T int | string | []byte](p PostgresContext, tableName, paramName string, height int64) (i T, enabled bool, err error) {
	ctx, tx, err := p.getCtxAndTx()
	if err != nil {
		return i, enabled, err
	}

	var stringVal string
	row := tx.QueryRow(ctx, types.GetParamOrFlagQuery(tableName, paramName, height))
	if tableName == types.ParamsTableName {
		err = row.Scan(&stringVal)
	} else {
		err = row.Scan(&stringVal, &enabled)
	}
	if err != nil {
		return
	}
	switch tp := any(i).(type) {
	case int, int32, int64:
		iConv, err := strconv.Atoi(stringVal)
		return any(iConv).(T), enabled, err
	case string:
		return any(stringVal).(T), enabled, err
	case []byte:
		v, err := hex.DecodeString(stringVal)
		return any(v).(T), enabled, err

	default:
		log.Fatalf("unhandled type for paramValue %T", tp)
	}
	return
}
