package schema

import (
	"encoding/hex"
	"fmt"
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
		enabled BOOLEAN NOT NULL,
		type val_type NOT NULL,
		value TEXT NOT NULL,
		PRIMARY KEY(name, height)
		)`

	FlagsTableName   = "flags"
	FlagsTableSchema = `(
		name VARCHAR(64) NOT NULL,
		height BIGINT NOT NULL,
		enabled BOOLEAN NOT NULL,
		value BOOLEAN NOT NULL,
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

func InsertParams(params *genesis.Params) string {
	val := reflect.ValueOf(params)
	var subQuery string
	for _, k := range govParamMetadataKeys {
		pnt := govParamMetadataMap[k]
		pVal := val.Elem().FieldByName(pnt.PropertyName)

		subQuery += `(`
		switch govParamMetadataMap[k].PropertyType {
		case ValTypeString:
			var stringVal string
			switch vt := pVal.Interface().(type) {
			case []byte:
				stringVal = hex.EncodeToString(vt)
			case string:
				stringVal = vt
			}
			subQuery += fmt.Sprintf("'%s', %d, true, '%s', '%s'", k, DefaultBigInt, pnt.PropertyType, stringVal)

		case ValTypeSmallInt, ValTypeBigInt:
			subQuery += fmt.Sprintf("'%s', %d, true, '%s', %d", k, DefaultBigInt, pnt.PropertyType, pVal.Interface())
		}
		subQuery += `),`
	}
	return fmt.Sprintf(`INSERT INTO %s VALUES %s`, ParamsTableName, subQuery[:len(subQuery)-1])
}

func GetParamQuery(paramName string) string {
	//TODO (@deblasis): Fix this
	return fmt.Sprintf(`SELECT value FROM %s WHERE name='%s' AND height<=%d and enabled=true order by height desc limit 1`, ParamsTableName, paramName, DefaultBigInt)
}

func NullifyParamQuery(paramName string, height int64) string {
	//TODO (@deblasis): Fix this
	return fmt.Sprintf(`UPDATE %s SET height=%d WHERE name='%s' AND height=%d`, ParamsTableName, height, paramName, DefaultBigInt)
}

type ParamTypes interface {
	int | int32 | int64 | []byte | string
}

func SetParam[T ParamTypes](paramName string, paramValue T, height int64) string {
	//TODO (@deblasis): Fix this
	fields := "name,height,value,enabled,type"

	subQuery := fmt.Sprintf(`SELECT %s`, fields)
	subQuery += fmt.Sprintf(` FROM %s WHERE name='%s' AND height=%d`, ParamsTableName, paramName, height)

	return fmt.Sprintf(`INSERT INTO %s(%s) %s`, ParamsTableName, fields, subQuery)
}

func ClearAllGovQuery() string {
	return fmt.Sprintf(`DELETE FROM %s`, ParamsTableName)
}

type govParamMetadata struct {
	PropertyType string
	PropertyName string
}

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
