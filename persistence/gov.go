package persistence

import (
	"encoding/hex"
	"fmt"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/utils"
	"strconv"

	"github.com/pokt-network/pocket/persistence/types"
	"github.com/pokt-network/pocket/runtime/genesis"
)

func (p *PostgresContext) InitGenesisParams(params *genesis.Params) error {
	ctx, tx := p.getCtxAndTx()
	if p.Height != 0 {
		return fmt.Errorf("cannot initialize params at height %d", p.Height)
	}
	_, err := tx.Exec(ctx, types.InsertParams(params, p.Height))
	return err
}

// Match paramName against the ParameterNameTypeMap struct and call the proper
// getter function getParamOrFlag[int | string | byte] and return its values
func (p *PostgresContext) GetParameter(paramName string, height int64) (v any, err error) {
	paramType := utils.GovParamMetadataMap[paramName].GoType
	switch paramType {
	case "int", "int32", "int64":
		v, _, err = getParamOrFlag[int](p, types.ParamsTableName, paramName, height)
	case "string":
		v, _, err = getParamOrFlag[string](p, types.ParamsTableName, paramName, height)
	case "[]uint8": // []byte
		v, _, err = getParamOrFlag[[]byte](p, types.ParamsTableName, paramName, height)
	default:
		return nil, fmt.Errorf("unhandled type for param (%s): got %s.", paramName, paramType)
	}
	return v, err
}

func (p *PostgresContext) GetIntParam(paramName string, height int64) (int, error) {
	v, _, err := getParamOrFlag[int](p, types.ParamsTableName, paramName, height)
	return v, err
}

func (p *PostgresContext) GetStringParam(paramName string, height int64) (string, error) {
	v, _, err := getParamOrFlag[string](p, types.ParamsTableName, paramName, height)
	return v, err
}

func (p *PostgresContext) GetBytesParam(paramName string, height int64) (param []byte, err error) {
	v, _, err := getParamOrFlag[[]byte](p, types.ParamsTableName, paramName, height)
	return v, err
}

func (p *PostgresContext) SetParam(paramName string, value any) error {
	return p.setParamOrFlag(paramName, value, nil)
}

func (p *PostgresContext) InitFlags() error {
	// TODO: not implemented
	return nil
}

func (p *PostgresContext) GetIntFlag(flagName string, height int64) (value int, enabled bool, err error) {
	return getParamOrFlag[int](p, types.FlagsTableName, flagName, height)
}

func (p *PostgresContext) GetStringFlag(flagName string, height int64) (value string, enabled bool, err error) {
	return getParamOrFlag[string](p, types.FlagsTableName, flagName, height)
}

func (p *PostgresContext) GetBytesFlag(flagName string, height int64) (value []byte, enabled bool, err error) {
	return getParamOrFlag[[]byte](p, types.FlagsTableName, flagName, height)
}

func (p *PostgresContext) SetFlag(flagName string, value any, enabled bool) error {
	return p.setParamOrFlag(flagName, value, &enabled)
}

// setParamOrFlag simply wraps the call to the generic function with the supplied underlying type
func (p *PostgresContext) setParamOrFlag(name string, value any, enabled *bool) error {
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
func setParamOrFlag[T types.SupportedParamTypes](p *PostgresContext, paramName string, paramValue T, enabled *bool) error {
	ctx, tx := p.getCtxAndTx()
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

func getParamOrFlag[T int | string | []byte](p *PostgresContext, tableName, paramName string, height int64) (i T, enabled bool, err error) {
	ctx, tx := p.getCtxAndTx()

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
		p.logger.Fatal().Msgf("unhandled type for paramValue %T", tp)
	}
	return
}

func (p *PostgresContext) getParamsUpdated(height int64) ([]*coreTypes.Param, error) {
	ctx, tx := p.getCtxAndTx()
	// Get all parameters / flags at current height
	rows, err := tx.Query(ctx, p.getParamsOrFlagsUpdateAtHeightQuery(types.ParamsTableName, height))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var paramSlice []*coreTypes.Param // Store returned rows
	// Loop over all rows returned and load them into the ParamOrFlag struct array
	for rows.Next() {
		param := new(coreTypes.Param)
		err := rows.Scan(&param.Name, &param.Value)
		if err != nil {
			return nil, err
		}
		param.Height = height
		paramSlice = append(paramSlice, param)
	}
	return paramSlice, nil
}

func (p *PostgresContext) getFlagsUpdated(height int64) ([]*coreTypes.Flag, error) {
	ctx, tx := p.getCtxAndTx()
	// Get all parameters / flags at current height
	rows, err := tx.Query(ctx, p.getParamsOrFlagsUpdateAtHeightQuery(types.FlagsTableName, height))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var flagSlice []*coreTypes.Flag // Store returned rows
	// Loop over all rows returned and load them into the ParamOrFlag struct array
	for rows.Next() {
		flag := new(coreTypes.Flag)
		err := rows.Scan(&flag.Name, &flag.Value, &flag.Enabled)
		if err != nil {
			return nil, err
		}
		flag.Height = height
		flagSlice = append(flagSlice, flag)
	}
	return flagSlice, nil
}

// GetAllParams returns a map of the current latest updated values for all parameters
// and their values in the form map[parameterName] = parameterValue
func (p *PostgresContext) GetAllParams() (map[string]string, error) {
	ctx, tx := p.getCtxAndTx()
	// Get all the parameters in their most recently updated form
	rows, err := tx.Query(ctx, p.getLatestParamsOrFlagsQuery(types.ParamsTableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	paramValueMap := make(map[string]string)
	for rows.Next() {
		var paramName, paramValue string
		if err := rows.Scan(&paramName, &paramValue); err != nil {
			return nil, err
		}
		paramValueMap[paramName] = paramValue
	}
	return paramValueMap, nil
}

func (p *PostgresContext) getParamsOrFlagsUpdateAtHeightQuery(tableName string, height int64) string {
	fields := "name,value"
	if tableName == types.FlagsTableName {
		fields += ",enabled"
	}
	// Build correct query to get all Params/Flags at certain height ordered by their name values
	return fmt.Sprintf("SELECT %s FROM %s WHERE height=%d ORDER BY name ASC", fields, tableName, height)
}

func (p *PostgresContext) getLatestParamsOrFlagsQuery(tableName string) string {
	fields := "name,value"
	if tableName == types.FlagsTableName {
		fields += ",enabled"
	}
	// Return a query to select all params or queries but only the most recent update for each
	return fmt.Sprintf("SELECT DISTINCT ON (name) %s FROM %s ORDER BY name ASC,%s.height DESC", fields, tableName, tableName)
}
