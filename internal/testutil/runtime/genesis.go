package runtime_testutil

import (
	"github.com/pokt-network/pocket/internal/testutil/p2p"
	"github.com/pokt-network/pocket/runtime/genesis"
	"github.com/pokt-network/pocket/runtime/test_artifacts"
	"github.com/pokt-network/pocket/shared/core/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/regen-network/gocuke"
)

func BaseGenesisStateMock(t gocuke.TestingT, valKeys []cryptoPocket.PrivateKey, serviceURLs []string) *genesis.GenesisState {
	t.Helper()

	genesisState := new(genesis.GenesisState)
	validators := make([]*types.Actor, len(valKeys))
	for i, valKey := range valKeys {
		addr := valKey.Address().String()
		mockActor := &types.Actor{
			ActorType:       types.ActorType_ACTOR_TYPE_VAL,
			Address:         addr,
			PublicKey:       valKey.PublicKey().String(),
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

func GenesisWithSequentialServiceURLs(t gocuke.TestingT, valKeys []cryptoPocket.PrivateKey) *genesis.GenesisState {
	t.Helper()

	serviceURLs := make([]string, len(valKeys))
	for i := range valKeys {
		serviceURLs[i] = p2p_testutil.NewServiceURL(i)
	}
	return BaseGenesisStateMock(t, valKeys, serviceURLs)
}
