package schema

import (
	"encoding/hex"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/pokt-network/pocket/shared/types/genesis"
)

// init initializes a map that contains the metadata extracted from `gov.proto`.
//
// Since protobuf files do not change at runtime, it seems efficient to do it here.
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
// It leverages metadata in the form of struct tags (see `parseGovProto` for more information).
//
// WARNING: reflections in prod
func InsertParams(params *genesis.Params, height int64) string {
	val := reflect.ValueOf(params)
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("INSERT INTO %s VALUES ", ParamsTableName))

	l := len(govParamMetadataKeys)
	for i, k := range govParamMetadataKeys {
		pnt := govParamMetadataMap[k]
		sb.WriteString(fmt.Sprintf("('%s', %d, '%s', ", k, height, pnt.PropertyType))
		pVal := val.Elem().FieldByName(pnt.PropertyName)
		pType := govParamMetadataMap[k].PropertyType
		switch pType {
		case ValTypeString:
			switch vt := pVal.Interface().(type) {
			case []byte:
				sb.WriteString(fmt.Sprintf("'%s')", hex.EncodeToString(vt)))
			case string:
				sb.WriteString(fmt.Sprintf("'%s')", vt))
			default:
				log.Fatalf("unhandled type for param: expected []byte or string, got %T", vt)
			}

		case ValTypeSmallInt, ValTypeBigInt:
			sb.WriteString(fmt.Sprintf("%d)", pVal.Interface()))
		default:
			log.Fatalf("unhandled PropertyType %s", pType)
		}

		if i < l-1 {
			sb.WriteString(",")
		}
	}

	constraint := fmt.Sprintf("%s_pkey", ParamsTableName)
	sb.WriteString(fmt.Sprintf(" ON CONFLICT ON CONSTRAINT %s DO UPDATE SET value=EXCLUDED.value, type=EXCLUDED.type", constraint))

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

type govParamMetadata struct {
	PropertyType string
	PropertyName string
}

// parseGovProto parses genesis.Params{} (generated from gov.proto) in order to extract metadata about its fields.
//
// The metadata comes in the form of struct tags that we attached to gov.proto and also from the tags that protoc injects automatically.
// Since currently we need to specify a mapping between the fields and a custom enum in the database (and potentially other things as well in the future),
// instead of having to maintain multiple maps, which would lead to having to maintain multiple sources of truth, we centralized the declaration of the fields
// and related metadata into the protobuf file.
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
