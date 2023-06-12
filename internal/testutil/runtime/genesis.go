package runtime_testutil

import (
	"github.com/pokt-network/pocket/internal/testutil"
	"github.com/pokt-network/pocket/runtime/genesis"
	"github.com/pokt-network/pocket/runtime/test_artifacts"
	"github.com/pokt-network/pocket/shared/core/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/regen-network/gocuke"
)

func BaseGenesisStateMock(t gocuke.TestingT, valKeys []cryptoPocket.PublicKey, serviceURLs []string) *genesis.GenesisState {
	t.Helper()

	genesisState := new(genesis.GenesisState)
	validators := make([]*types.Actor, len(valKeys))
	for i, valKey := range valKeys {
		addr := valKey.Address().String()
		mockActor := &types.Actor{
			ActorType:       types.ActorType_ACTOR_TYPE_VAL,
			Address:         addr,
			PublicKey:       valKey.String(),
			ServiceUrl:      serviceURLs[i],
			StakedAmount:    test_artifacts.DefaultStakeAmountString,
			PausedHeight:    int64(0),
			UnstakingHeight: int64(0),
			Output:          addr,
		}
		validators[i] = mockActor
	}
	genesisState.Validators = validators

	return genesisState
}

func BaseGenesisStateMockFromServiceURLKeyMap(t gocuke.TestingT, serviceURLKeyMap map[string]cryptoPocket.PrivateKey) *genesis.GenesisState {
	t.Helper()

	var validators []*types.Actor
	genesisState := new(genesis.GenesisState)
	for serviceURL, privKey := range serviceURLKeyMap {
		addr := privKey.Address().String()
		mockValidator := &types.Actor{
			ActorType:       types.ActorType_ACTOR_TYPE_VAL,
			Address:         addr,
			PublicKey:       privKey.PublicKey().String(),
			ServiceUrl:      serviceURL,
			StakedAmount:    test_artifacts.DefaultStakeAmountString,
			PausedHeight:    int64(0),
			UnstakingHeight: int64(0),
			Output:          addr,
		}
		validators = append(validators, mockValidator)
	}
	genesisState.Validators = validators

	return genesisState
}

func GenesisWithSequentialServiceURLs(t gocuke.TestingT, valKeys []cryptoPocket.PublicKey) *genesis.GenesisState {
	t.Helper()

	serviceURLs := make([]string, len(valKeys))
	for i := range valKeys {
		serviceURLs[i] = testutil.NewServiceURL(i + 1)
	}
	return BaseGenesisStateMock(t, valKeys, serviceURLs)
}
