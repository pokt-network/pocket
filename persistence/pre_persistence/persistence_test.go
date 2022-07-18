package pre_persistence

import (
	"testing"

	"github.com/pokt-network/pocket/shared/config"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/types"
	"github.com/pokt-network/pocket/shared/types/genesis"
	"github.com/stretchr/testify/require"
	"github.com/syndtr/goleveldb/leveldb/comparer"
	"github.com/syndtr/goleveldb/leveldb/memdb"
)

func NewTestingPrePersistenceModule(t *testing.T) *PrePersistenceModule {
	db := memdb.New(comparer.DefaultComparer, 10000000)
	cfg := &config.Config{
		GenesisSource: &genesis.GenesisSource{
			Source: &genesis.GenesisSource_Config{
				Config: genesisConfig(),
			},
		},
	}
	err := cfg.HydrateGenesisState()
	require.NoError(t, err)

	return NewPrePersistenceModule(db, types.NewMempool(10000, 10000), cfg)
}

func NewTestingPrePersistenceContext(t *testing.T) modules.PersistenceContext {
	persistenceModule := NewTestingPrePersistenceModule(t)
	persistenceContext, err := persistenceModule.NewContext(0)
	if err != nil {
		t.Fatal(err)
	}
	return persistenceContext
}

func genesisConfig() *genesis.GenesisConfig {
	config := &genesis.GenesisConfig{
		NumValidators:   5,
		NumApplications: 1,
		NumFisherman:    1,
		NumServicers:    5,
		// ValidatorUrlFormat: "",
		SeedStart: 42,
	}
	return config
}
