package persistence

import (
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/persistence/types"
	"github.com/pokt-network/pocket/runtime/genesis"
	"github.com/pokt-network/pocket/shared/modules"
)

func (p *PostgresContext) GetVersionAtHeight(height int64) (string, error) {
	ctx, tx := p.getCtxAndTx()
	var version string
	row := tx.QueryRow(ctx, `
	SELECT version FROM upgrades WHERE height <= $1 ORDER BY height DESC LIMIT 1
`, height)
	if err := row.Scan(&version); err != nil {
		return "", err
	}
	return version, nil
}

// TODO(#882): Implement this function
func (p *PostgresContext) GetRevisionNumber(height int64) uint64 {
	return 1
}

// TODO: Implement this function
func (p *PostgresContext) GetSupportedChains(height int64) ([]string, error) {
	// This is a placeholder function for the RPC endpoint "v1/query/supportedchains"
	return []string{}, nil
}

func (p *PostgresContext) InitGenesisParams(params *genesis.Params) error {
	ctx, tx := p.getCtxAndTx()
	if p.Height != 0 {
		return fmt.Errorf("cannot initialize params at height %d", p.Height)
	}

	sql := types.InsertParams(params, p.Height)

	if e := p.logger.Trace(); e.Enabled() {
		e.Msg("initializing genesis params: " + sql)
	}

	_, err := tx.Exec(ctx, sql)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
INSERT INTO upgrades (signer, version, height, created) VALUES ($1, $2, $3, $4)
ON CONFLICT (version)
DO UPDATE SET signer=EXCLUDED.signer, version=EXCLUDED.version, height=EXCLUDED.height, created=EXCLUDED.created
`, params.AclOwner, "1.0.0", p.Height, p.Height)
	return err
}

func GetParameter[T int | string | []byte](p modules.PersistenceReadContext, paramName string, height int64) (i T, err error) {
	switch tp := any(i).(type) {
	case int:
		v, er := p.GetIntParam(paramName, height)
		return any(v).(T), er
	case string:
		v, er := p.GetStringParam(paramName, height)
		return any(v).(T), er
	case []byte:
		v, er := p.GetBytesParam(paramName, height)
		return any(v).(T), er
	default:
		logger.Global.Fatal().Msgf("unhandled type for param (%s): got %s.", paramName, tp)
	}
	return
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
	// TODO(0xbigboss): not implemented
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

// GetAllParams returns a map of the current latest updated values for all parameters
// and their values in the form map[parameterName] = parameterValue
func (p *PostgresContext) GetAllParams() ([][]string, error) {
	ctx, tx := p.getCtxAndTx()
	// Get all the parameters in their most recently updated form
	rows, err := tx.Query(ctx, p.getLatestParamsOrFlagsQuery(types.ParamsTableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	paramSlice := make([][]string, 0)
	for rows.Next() {
		var paramName, paramValue string
		if err := rows.Scan(&paramName, &paramValue); err != nil {
			return nil, err
		}
		paramSlice = append(paramSlice, []string{paramName, paramValue})
	}
	return paramSlice, nil
}

func (p *PostgresContext) getLatestParamsOrFlagsQuery(tableName string) string {
	fields := "name,value"
	if tableName == types.FlagsTableName {
		fields += ",enabled"
	}
	// Return a query to select all params or queries but only the most recent update for each
	return fmt.Sprintf("SELECT DISTINCT ON (name) %s FROM %s ORDER BY name ASC,%s.height DESC", fields, tableName, tableName)
}

func (p *PostgresContext) SetUpgrade(signer, version string, height int64) error {
	ctx, tx := p.getCtxAndTx()
	_, err := tx.Exec(ctx, `
INSERT INTO upgrades(signer, version, height, created) VALUES ($1, $2, $3, $4)
`, signer, version, height, p.Height)
	if err != nil {
		p.logger.Debug().Err(err).Msg("failed to insert upgrade")
		return err
	}
	return nil
}
