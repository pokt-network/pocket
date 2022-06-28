package test

import (
	"encoding/hex"
	"reflect"

	"github.com/iancoleman/strcase"
	"github.com/pokt-network/pocket/persistence"
	query "github.com/pokt-network/pocket/persistence/schema"
)

func GetGenericActor[T any](protocolActor query.ProtocolActor, getActor func(persistence.PostgresContext, []byte) (T, error)) func(persistence.PostgresContext, string) (*query.GenericActor, error) {
	return func(db persistence.PostgresContext, address string) (*query.GenericActor, error) {
		addr, err := hex.DecodeString(address)
		if err != nil {
			return nil, err
		}
		actor, err := getActor(db, addr)
		if err != nil {
			return nil, err
		}
		genericActor := getActorValues(protocolActor, reflect.Indirect(reflect.ValueOf(actor)))
		return &genericActor, nil
	}
}

func NewTestGenericActor[T any](protocolActor query.ProtocolActor, newActor func() (T, error)) func() (query.GenericActor, error) {
	return func() (query.GenericActor, error) {
		actor, err := newActor()
		if err != nil {
			return query.GenericActor{}, err
		}
		return getActorValues(protocolActor, reflect.Indirect(reflect.ValueOf(actor))), nil
	}
}

func getActorValues(protocolActor query.ProtocolActor, actorValue reflect.Value) query.GenericActor {
	chains := make([]string, 0)
	if actorValue.FieldByName("Chains").Kind() != 0 { // != reflect.Zero(reflect.TypeOf(chains)) {
		chains = actorValue.FieldByName("Chains").Interface().([]string)
	}

	genericParamName := strcase.ToCamel(protocolActor.GetActorSpecificColName())

	return query.GenericActor{
		Address:         hex.EncodeToString(actorValue.FieldByName("Address").Bytes()),
		PublicKey:       hex.EncodeToString(actorValue.FieldByName("PublicKey").Bytes()),
		StakedTokens:    actorValue.FieldByName("StakedTokens").String(),
		GenericParam:    actorValue.FieldByName(genericParamName).String(),
		OutputAddress:   hex.EncodeToString(actorValue.FieldByName("Output").Bytes()),
		PausedHeight:    int64(actorValue.FieldByName("PausedHeight").Uint()),
		UnstakingHeight: int64(actorValue.FieldByName("UnstakingHeight").Int()),
		Chains:          chains,
	}
}
