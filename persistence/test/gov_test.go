package test

import (
	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/shared/types"
	"testing"
)

func TestInitParams(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	if err := db.InitParams(); err != nil {
		t.Fatal(err)
	}
}

func TestGetSetParam(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	if err := db.InitParams(); err != nil {
		t.Fatal(err)
	}
	if err := db.SetParam(types.AppMaxChainsParamName, ParamToUpdate); err != nil {
		t.Fatal(err)
	}
	maxChains, err := db.GetMaxAppChains()
	if err != nil {
		t.Fatal(err)
	}
	if maxChains != ParamToUpdate {
		t.Fatal("unexpected param value")
	}
}
