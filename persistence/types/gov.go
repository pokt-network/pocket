package types

import (
	"encoding/hex"
	"fmt"
	"github.com/pokt-network/pocket/shared/utils"
	"log"
	"reflect"
	"strings"

	"github.com/pokt-network/pocket/runtime/genesis"
)

const (
	ValTypeName     = "val_type"
	ValTypeString   = "STRING"
	ValTypeBigInt   = "BIGINT"
	ValTypeSmallInt = "SMALLINT"

	ParamsTableName   = "params"
	ParamsTableSchema = `(
		name VARCHAR(64) NOT NULL,
		height BIGINT NOT NULL,
		type val_type NOT NULL,
		value TEXT NOT NULL,
		PRIMARY KEY(name, height)
		)`

	FlagsTableName   = "flags"
	FlagsTableSchema = `(
		name VARCHAR(64) NOT NULL,
		height BIGINT NOT NULL,
		type val_type NOT NULL,
		value TEXT NOT NULL,
		enabled BOOLEAN NOT NULL,
		PRIMARY KEY(name, height)
		)`
)

var (
	ValTypeEnumTypes = fmt.Sprintf(`(
		'%s',
		'%s',
		'%s'
		)`,
		ValTypeString,
		ValTypeBigInt,
		ValTypeSmallInt,
	)
)

// InsertParams generates the SQL INSERT statement given a *genesis.Params
// It leverages metadata in the form of struct tags (see `parseGovProto` for more information).
// WARNING: reflections in prod
func InsertParams(params *genesis.Params, height int64) string {
	val := reflect.ValueOf(params)
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("INSERT INTO %s VALUES ", ParamsTableName))

	l := len(utils.GovParamMetadataKeys)
	for i, k := range utils.GovParamMetadataKeys {
		pnt := utils.GovParamMetadataMap[k]
		sb.WriteString(fmt.Sprintf("('%s', %d, '%s', ", k, height, pnt.PoktType))
		pVal := val.Elem().FieldByName(pnt.PropertyName)
		switch pnt.PoktType {
		case ValTypeString:
			switch vt := pVal.Interface().(type) {
			case []byte:
				fmt.Fprintf(&sb, "'%s')", hex.EncodeToString(vt))
			case string:
				fmt.Fprintf(&sb, "'%s')", vt)
			default:
				log.Fatalf("unhandled type for param: expected []byte or string, got %T", vt)
			}

		case ValTypeSmallInt, ValTypeBigInt:
			fmt.Fprintf(&sb, "%d)", pVal.Interface())
		default:
			log.Fatalf("unhandled PropertyType: %s.", pnt.PoktType)
		}

		if i < l-1 {
			sb.WriteString(",")
		}
	}

	constraint := fmt.Sprintf("%s_pkey", ParamsTableName)
	fmt.Fprintf(&sb, " ON CONFLICT ON CONSTRAINT %s DO UPDATE SET value=EXCLUDED.value, type=EXCLUDED.type", constraint)

	return sb.String()
}

func GetParamOrFlagQuery(tableName, flagName string, height int64) string {
	fields := "value"
	if tableName == FlagsTableName {
		fields += ",enabled"
	}
	return fmt.Sprintf(`SELECT %s FROM %s WHERE name='%s' AND height<=%d ORDER BY height DESC LIMIT 1`, fields, tableName, flagName, height)
}

// SupportedParamTypes represents the types currently supported for the `value` property in params and flags
type SupportedParamTypes interface {
	int | int32 | int64 | []byte | string
}

// InsertParamOrFlag returns the SQL SQL INSERT (with conflict handling so that it's effectively an "upsert") required to set a parameter/flag
func InsertParamOrFlag[T SupportedParamTypes](tableName, name string, height int64, value T, enabled *bool) string {
	fields := "name,height,type,value"
	upsertFields := "type=EXCLUDED.type,value=EXCLUDED.value"
	if tableName == FlagsTableName {
		fields += ",enabled"
		upsertFields += ",enabled=EXCLUDED.enabled"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("INSERT INTO %s(%s) VALUES ('%s', %d, ", tableName, fields, name, height))

	switch tp := any(value).(type) {
	case int, int32:
		sb.WriteString(fmt.Sprintf("'%s', %d", ValTypeSmallInt, tp))
	case int64:
		sb.WriteString(fmt.Sprintf("'%s', %d", ValTypeBigInt, tp))
	case []byte:
		sb.WriteString(fmt.Sprintf("'%s', '%s'", ValTypeBigInt, hex.EncodeToString(tp)))
	case string:
		sb.WriteString(fmt.Sprintf("'%s', '%s'", ValTypeString, tp))
	default:
		log.Fatalf("unhandled type for paramValue %T", tp)
	}

	if enabled != nil {
		sb.WriteString(fmt.Sprintf(",%t", *enabled))
	}

	sb.WriteString(")")

	constraint := fmt.Sprintf("%s_pkey", tableName)
	sb.WriteString(fmt.Sprintf("ON CONFLICT ON CONSTRAINT %s DO UPDATE SET %s", constraint, upsertFields))
	return sb.String()
}

func ClearAllGovParamsQuery() string {
	return fmt.Sprintf(`DELETE FROM %s`, ParamsTableName)
}

func ClearAllGovFlagsQuery() string {
	return fmt.Sprintf(`DELETE FROM %s`, FlagsTableName)
}
