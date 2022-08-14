package test

import (
	"encoding/hex"
	"reflect"

	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/persistence/schema"
)

func GetGenericActor[T any](protocolActorSchema schema.ProtocolActorSchema, getActor func(persistence.PostgresContext, []byte) (T, error)) func(persistence.PostgresContext, string) (*schema.BaseActor, error) {
	return func(db persistence.PostgresContext, address string) (*schema.BaseActor, error) {
		addr, err := hex.DecodeString(address)
		if err != nil {
			return nil, err
		}
		actor, err := getActor(db, addr)
		if err != nil {
			return nil, err
		}
		baseActor := getActorValues(protocolActorSchema, reflect.Indirect(reflect.ValueOf(actor)))
		return &baseActor, nil
	}
}

func NewTestGenericActor[T any](protocolActorSchema schema.ProtocolActorSchema, newActor func() (T, error)) func() (schema.BaseActor, error) {
	return func() (schema.BaseActor, error) {
		actor, err := newActor()
		if err != nil {
			return schema.BaseActor{}, err
		}
		return getActorValues(protocolActorSchema, reflect.Indirect(reflect.ValueOf(actor))), nil
	}
}

func getActorValues(_ schema.ProtocolActorSchema, actorValue reflect.Value) schema.BaseActor {
	chains := make([]string, 0)
	if actorValue.FieldByName("Chains").Kind() != 0 {
		chains = actorValue.FieldByName("Chains").Interface().([]string)
	}

	return schema.BaseActor{
		Address:            actorValue.FieldByName("Address").String(),
		PublicKey:          actorValue.FieldByName("PublicKey").String(),
		StakedTokens:       actorValue.FieldByName("StakedAmount").String(),
		ActorSpecificParam: actorValue.FieldByName("GenericParam").String(),
		OutputAddress:      actorValue.FieldByName("Output").String(),
		PausedHeight:       int64(actorValue.FieldByName("PausedHeight").Int()),
		UnstakingHeight:    int64(actorValue.FieldByName("UnstakingHeight").Int()),
		Chains:             chains,
	}
}
