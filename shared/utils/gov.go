package utils

import (
	"github.com/pokt-network/pocket/runtime/genesis"
	"reflect"
	"strings"
)

// init initializes a map that contains the metadata extracted from `gov.proto`.
// Since protobuf files do not change at runtime, it seems efficient to do it here.
func init() {
	GovParamMetadataMap = parseGovProto()
}

var (
	GovParamMetadataMap  map[string]govParamMetadata
	GovParamMetadataKeys []string
)

type govParamMetadata struct {
	PropertyName string
	ParamName    string
	ParamOwner   string
	PoktType     string
	GoType       string
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
		poktOwner := extractStructTag(poktTag, "owner=")
		golangType := field.Type.Name() // Get string version of field's Golang type
		protoName := extractStructTag(protoTag, "name=")
		govParamMetadataMap[protoName] = govParamMetadata{
			PropertyName: field.Name,
			ParamName:    protoName,
			ParamOwner:   poktOwner,
			PoktType:     poktValType,
			GoType:       golangType,
		}
		GovParamMetadataKeys = append(GovParamMetadataKeys, protoName)
	}
	return govParamMetadataMap
}

func extractStructTag(structTag, key string) string {
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
