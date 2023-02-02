package types

import (
	"github.com/mindstand/gogm/v2"
)

type NeoPool struct {
	gogm.BaseNode // Provides required node fields for neo4j DB

	Name   string `gogm:"name=Name"`
	Amount string `gogm:"name=Amount"`
}

type NeoAccount struct {
	gogm.BaseNode // Provides required node fields for neo4j DB

	Address string `gogm:"name=Address"`
	Amount  string `gogm:"name=Amount"`
}

type NeoActor struct {
	gogm.BaseNode // Provides required node fields for neo4j DB

	ActorType string `gogm:"name=ActorType"`
	Address   string `gogm:"name=Address"`
	// PublicKey string `gogm:"name=PublicKey"`
	// Chains          []string  `gogm:"name=Chains"`
	// GenericParam    string `gogm:"name=GenericParam"`
	StakedAmount string `gogm:"name=StakedAmount"`
	// PausedHeight    int64  `gogm:"name=PausedHeight"`
	// UnstakingHeight int64  `gogm:"name=UnstakingHeight"`
	// Output          string `gogm:"name=Output"`
}
