package schema

import (
	"encoding/hex"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/pokt-network/pocket/shared/types/genesis"
)

// TODO(https://github.com/pokt-network/pocket/issues/76): Optimize gov parameters implementation & schema.

func init() {
	govParamMetadataMap = parseGovProto()
}

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
	govParamMetadataMap  map[string]govParamMetadata
	govParamMetadataKeys []string

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
//
// WARNING: reflections in prod
func InsertParams(params *genesis.Params, height int64) string {
	val := reflect.ValueOf(params)
	var subQuery string
	for _, k := range govParamMetadataKeys {
		pnt := govParamMetadataMap[k]
		pVal := val.Elem().FieldByName(pnt.PropertyName)

		subQuery += `(`
		pType := govParamMetadataMap[k].PropertyType
		switch pType {
		case ValTypeString:
			var stringVal string
			switch vt := pVal.Interface().(type) {
			case []byte:
				stringVal = hex.EncodeToString(vt)
			case string:
				stringVal = vt
			default:
				log.Fatalf("unhandled type for param: expected []byte or string, got %T", vt)
			}
			subQuery += fmt.Sprintf("'%s', %d, '%s', '%s'", k, height, pnt.PropertyType, stringVal)

		case ValTypeSmallInt, ValTypeBigInt:
			subQuery += fmt.Sprintf("'%s', %d, '%s', %d", k, height, pnt.PropertyType, pVal.Interface())
		default:
			log.Fatalf("unhandled PropertyType %s", pType)
		}
		subQuery += `),`
	}
	return fmt.Sprintf(`INSERT INTO %s VALUES %s ON CONFLICT ON CONSTRAINT %s DO UPDATE SET value=EXCLUDED.value, type=EXCLUDED.type`, ParamsTableName, subQuery[:len(subQuery)-1], "params_pkey")
}

func GetParamQuery(paramName string, height int64) string {
	return fmt.Sprintf(`SELECT value FROM %s WHERE name='%s' AND height<=%d ORDER BY height DESC LIMIT 1`, ParamsTableName, paramName, height)
}

func GetFlagQuery(flagName string, height int64) string {
	return fmt.Sprintf(`SELECT value, enabled FROM %s WHERE name='%s' AND height<=%d ORDER BY height DESC LIMIT 1`, FlagsTableName, flagName, height)
}

type SupportedParamTypes interface {
	int | int32 | int64 | []byte | string
}

func SetParamQuery[T SupportedParamTypes](paramName string, paramValue T, height int64) string {
	fields := "name,height,type,value"

	var value, valType string
	switch tp := any(paramValue).(type) {
	case int, int32:
		valType = ValTypeSmallInt
		value = fmt.Sprintf("%d", tp)
	case int64:
		valType = ValTypeBigInt
		value = fmt.Sprintf("%d", tp)
	case []byte:
		valType = ValTypeString
		value = fmt.Sprintf("'%s'", hex.EncodeToString(tp))
	case string:
		valType = ValTypeString
		value = fmt.Sprintf("'%s'", tp)
	default:
		log.Fatalf("unhandled type for paramValue %T", tp)
	}

	constraint := fmt.Sprintf("%s_pkey", ParamsTableName)
	return fmt.Sprintf(`INSERT INTO %s(%s) VALUES ('%s', %d, '%s', %s) ON CONFLICT ON CONSTRAINT %s DO UPDATE SET value=EXCLUDED.value, type=EXCLUDED.type`, ParamsTableName, fields, paramName, height, valType, value, constraint)
}

func SetFlagQuery[T SupportedParamTypes](paramName string, flagValue T, enabled bool, height int64) string {
	fields := "name,height,type,value,enabled"

	enabledStr := "true"
	if !enabled {
		enabledStr = "false"
	}

	var value, valType string
	switch tp := any(flagValue).(type) {
	case int, int32:
		valType = ValTypeSmallInt
		value = fmt.Sprintf("%d", tp)
	case int64:
		valType = ValTypeBigInt
		value = fmt.Sprintf("%d", tp)
	case []byte:
		valType = ValTypeString
		value = fmt.Sprintf("'%s'", hex.EncodeToString(tp))
	case string:
		valType = ValTypeString
		value = fmt.Sprintf("'%s'", tp)
	default:
		log.Fatalf("unhandled type for paramValue %T", tp)
	}

	constraint := fmt.Sprintf("%s_pkey", FlagsTableName)
	fmt.Println(fmt.Sprintf(`INSERT INTO %s(%s) VALUES ('%s', %d, '%s', %s, %s) ON CONFLICT ON CONSTRAINT %s DO UPDATE SET value=EXCLUDED.value, type=EXCLUDED.type, enabled=EXCLUDED.enabled`, FlagsTableName, fields, paramName, height, valType, value, enabledStr, constraint))
	return fmt.Sprintf(`INSERT INTO %s(%s) VALUES ('%s', %d, '%s', %s, %s) ON CONFLICT ON CONSTRAINT %s DO UPDATE SET value=EXCLUDED.value, type=EXCLUDED.type, enabled=EXCLUDED.enabled`, FlagsTableName, fields, paramName, height, valType, value, enabledStr, constraint)
}

func ClearAllGovParamsQuery() string {
	return fmt.Sprintf(`DELETE FROM %s`, ParamsTableName)
}

func ClearAllGovFlagsQuery() string {
	return fmt.Sprintf(`DELETE FROM %s`, FlagsTableName)
}

type govParamMetadata struct {
	PropertyType string
	PropertyName string
}

// parseGovProto parses genesis.Params{} (generated from gov.proto) in order to extract metadata about its fields
//
// WARNING: reflections in prod
func parseGovProto() (govParamMetadataMap map[string]govParamMetadata) {
	govParamMetadataMap = make(map[string]govParamMetadata)
	fields := reflect.VisibleFields(reflect.TypeOf(genesis.Params{}))
	for _, field := range fields {
		if !field.IsExported() {
			continue
		}
		poktTag := field.Tag.Get("pokt")
		protoTag := field.Tag.Get("protobuf")
		poktValType := extractStructTag(poktTag, "val_type=")
		protoName := extractStructTag(protoTag, "name=")
		govParamMetadataMap[protoName] = govParamMetadata{
			PropertyType: poktValType,
			PropertyName: field.Name,
		}
		govParamMetadataKeys = append(govParamMetadataKeys, protoName)
	}

	return govParamMetadataMap
}

func extractStructTag(structTag string, key string) string {
	for len(structTag) > 0 {
		i := strings.IndexByte(structTag, ',')
		if i < 0 { // not found
			i = len(structTag)
		}
		s := structTag[:i]
		if strings.HasPrefix(s, key) {
			return s[len(key):]
		}
		structTag = strings.TrimPrefix(structTag[i:], ",")
	}
	return ""
}
