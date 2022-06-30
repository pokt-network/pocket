package test

import (
	"encoding/hex"
	"reflect"

	"github.com/iancoleman/strcase"
	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/persistence/schema"
)

func GetGenericActor[T any](protocolActorSchema schema.ProtocolActorSchema, getActor func(persistence.PostgresContext, []byte) (T, error)) func(persistence.PostgresContext, string) (*schema.GenericActor, error) {
	return func(db persistence.PostgresContext, address string) (*schema.GenericActor, error) {
		addr, err := hex.DecodeString(address)
		if err != nil {
			return nil, err
		}
		actor, err := getActor(db, addr)
		if err != nil {
			return nil, err
		}
		genericActor := getActorValues(protocolActorSchema, reflect.Indirect(reflect.ValueOf(actor)))
		return &genericActor, nil
	}
}

func NewTestGenericActor[T any](protocolActorSchema schema.ProtocolActorSchema, newActor func() (T, error)) func() (schema.GenericActor, error) {
	return func() (schema.GenericActor, error) {
		actor, err := newActor()
		if err != nil {
			return schema.GenericActor{}, err
		}
		return getActorValues(protocolActorSchema, reflect.Indirect(reflect.ValueOf(actor))), nil
	}
}

func getActorValues(protocolActorSchema schema.ProtocolActorSchema, actorValue reflect.Value) schema.GenericActor {
	chains := make([]string, 0)
	if actorValue.FieldByName("Chains").Kind() != 0 {
		chains = actorValue.FieldByName("Chains").Interface().([]string)
	}

	actorSpecificParam := strcase.ToCamel(protocolActorSchema.GetActorSpecificColName())

	return schema.GenericActor{
		Address:            hex.EncodeToString(actorValue.FieldByName("Address").Bytes()),
		PublicKey:          hex.EncodeToString(actorValue.FieldByName("PublicKey").Bytes()),
		StakedTokens:       actorValue.FieldByName("StakedTokens").String(),
		ActorSpecificParam: actorValue.FieldByName(actorSpecificParam).String(),
		OutputAddress:      hex.EncodeToString(actorValue.FieldByName("Output").Bytes()),
		PausedHeight:       int64(actorValue.FieldByName("PausedHeight").Uint()),
		UnstakingHeight:    int64(actorValue.FieldByName("UnstakingHeight").Int()),
		Chains:             chains,
	}
}
