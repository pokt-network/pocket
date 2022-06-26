package test

import (
	"encoding/hex"
	"reflect"

	"github.com/pokt-network/pocket/persistence"
	query "github.com/pokt-network/pocket/persistence/schema"
)

func GetGenericActor[T any](getActor func(persistence.PostgresContext, []byte) (T, error)) func(persistence.PostgresContext, string) (*query.GenericActor, error) {
	return func(db persistence.PostgresContext, address string) (*query.GenericActor, error) {
		addr, err := hex.DecodeString(address)
		if err != nil {
			return nil, err
		}
		actor, err := getActor(db, addr)
		if err != nil {
			return nil, err
		}
		genericActor := getActorValues(reflect.Indirect(reflect.ValueOf(actor)))
		return &genericActor, nil
	}
}

func NewTestGenericActor[T any](newActor func() (T, error)) func() (query.GenericActor, error) {
	return func() (query.GenericActor, error) {
		actor, err := newActor()
		if err != nil {
			return query.GenericActor{}, err
		}
		return getActorValues(reflect.ValueOf(actor)), nil
	}
}

func getActorValues(actorValue reflect.Value) query.GenericActor {
	return query.GenericActor{
		Address:         hex.EncodeToString(actorValue.FieldByName("Address").Bytes()),
		PublicKey:       hex.EncodeToString(actorValue.FieldByName("PublicKey").Bytes()),
		StakedTokens:    actorValue.FieldByName("StakedTokens").String(),
		GenericParam:    actorValue.FieldByName("ServiceUrl").String(),
		OutputAddress:   hex.EncodeToString(actorValue.FieldByName("Output").Bytes()),
		PausedHeight:    int64(actorValue.FieldByName("PausedHeight").Uint()),
		UnstakingHeight: int64(actorValue.FieldByName("UnstakingHeight").Int()),
		Chains:          actorValue.FieldByName("Chains").Interface().([]string),
	}
}
