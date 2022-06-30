package test

import (
	"encoding/hex"
	"fmt"
	"reflect"
	"testing"

	"github.com/iancoleman/strcase"
	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/persistence/schema"
)

func TestInsertProtocolActorAndExists(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}

	app_fns := map[string]any{
		"GetExists":                   db.GetAppExists,
		"Insert":                      db.InsertApp,
		"Update":                      db.UpdateApp,
		"Delete":                      db.DeleteApp,
		"GetReadyToUnstake":           db.GetAppsReadyToUnstake,
		"GetStatus":                   db.GetAppStatus,
		"SetUnstakingHeightAndStatus": db.SetAppUnstakingHeightAndStatus,
		"GetPauseHeightIfExists":      db.GetAppPauseHeightIfExists,
		"SetStatusAndUnstakingHeightPausedBefore": db.SetAppStatusAndUnstakingHeightPausedBefore,
		"SetPauseHeight":   db.SetAppPauseHeight,
		"GetOutputAddress": db.GetAppOutputAddress}
	fmt.Println(app_fns)

	// newTestActor := newTestApp
	// dbGetActorExists := db.GetAppExists
	// dbInsertActorH0 := db.InsertApp
	// db.Height += 1
	// dbInsertActorH1 := db.InsertApp

	// cases := []struct {
	// 	name         string
	// 	newTestActor func() (*typesGenesis.App, error)
	// }{
	// 	{"Application", newTestApp, dbGetActorExists, dbInsertActorH0, dbInsertActorH1},
	// }

	// for _, tc := range cases {
	// 	t.Run(tc.name, func(t *testing.T) {

	// actor, err := newTestActor()
	// require.NoError(t, err)

	// err = app_fns["Insert"](
	// 	actor.Address,
	// 	actor.PublicKey,
	// 	actor.Output,
	// 	false,
	// 	DefaultStakeStatus,
	// 	DefaultMaxRelays,
	// 	DefaultStake,
	// 	DefaultChains,
	// 	DefaultPauseHeight,
	// 	DefaultUnstakingHeight)
	// require.NoError(t, err)

	// actor2, err := newTestActor()
	// require.NoError(t, err)

	// err = dbInsertActorH1(
	// 	actor2.Address,
	// 	actor2.PublicKey,
	// 	actor2.Output,
	// 	false,
	// 	DefaultStakeStatus,
	// 	DefaultMaxRelays,
	// 	DefaultStake,
	// 	DefaultChains,
	// 	DefaultPauseHeight,
	// 	DefaultUnstakingHeight)
	// require.NoError(t, err)

	// log.Println(actor.Address, actor2.Address)

	// exists, err := dbGetActorExists(actor.Address, 0)
	// require.NoError(t, err)
	// require.True(t, exists, "actor that should exist at previous height does not")
	// exists, err = dbGetActorExists(actor.Address, 1)
	// require.NoError(t, err)
	// require.True(t, exists, "actor that should exist at current height does not")

	// exists, err = dbGetActorExists(actor2.Address, 0)
	// require.NoError(t, err)
	// require.False(t, exists, "actor that should not exist at previous height appears to")
	// exists, err = dbGetActorExists(actor2.Address, 1)
	// require.NoError(t, err)
	// require.True(t, exists, "actor that should exist at current height does not")
	// 	})
	// }
}

// func newTestActor[T]() (*T, error) {
// 	operatorKey, err := crypto.GeneratePublicKey()
// 	if err != nil {
// 		return nil, err
// 	}

// 	outputAddr, err := crypto.GenerateAddress()
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &typesGenesis.App{
// 		Address:         operatorKey.Address(),
// 		PublicKey:       operatorKey.Bytes(),
// 		Paused:          false,
// 		Status:          typesGenesis.DefaultStakeStatus,
// 		Chains:          typesGenesis.DefaultChains,
// 		MaxRelays:       DefaultMaxRelays,
// 		StakedTokens:    typesGenesis.DefaultStake,
// 		PausedHeight:    uint64(DefaultPauseHeight),
// 		UnstakingHeight: DefaultUnstakingHeight,
// 		Output:          outputAddr,
// 	}, nil
// }

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
