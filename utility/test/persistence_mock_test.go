package test

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pokt-network/pocket/runtime"
	"github.com/pokt-network/pocket/runtime/test_artifacts"
	"github.com/pokt-network/pocket/shared/modules"
	modulesMock "github.com/pokt-network/pocket/shared/modules/mocks"
	"github.com/pokt-network/pocket/utility"
	typesUtil "github.com/pokt-network/pocket/utility/types"
	"github.com/stretchr/testify/require"
)

func NewTestUtilityContext(t *testing.T, height int64, options ...func(*modulesMock.MockPersistenceRWContext)) utility.UtilityContext {
	ctrl := gomock.NewController(t)
	persistenceContextMock := modulesMock.NewMockPersistenceRWContext(ctrl)

	// Base mocks required for all tests
	persistenceContextMock.EXPECT().GetHeight().Return(height, nil).AnyTimes()
	persistenceContextMock.EXPECT().Release().Return(nil).AnyTimes()

	// Adding behavioural mocks based on the options provided
	for _, o := range options {
		o(persistenceContextMock)
	}

	return utility.UtilityContext{
		LatestHeight: height,
		Mempool:      NewTestingMempool(t),
		Context: &utility.Context{
			PersistenceRWContext: persistenceContextMock,
			SavePointsM:          make(map[string]struct{}),
			SavePoints:           make([][]byte, 0),
		},
	}
}

func persistenceRuntimeMgr(t *testing.T) *runtime.Manager {
	genesisState, validatorKeys := test_artifacts.NewGenesisState(1, 1, 1, 1)
	config := test_artifacts.NewDefaultConfigs(validatorKeys)
	return runtime.NewManager(config[0], genesisState)
}

// Stateful mocks of the basic account related functions in the persistence module interface
func withBaseAccountMock(
	t *testing.T,
	runTimeMgr *runtime.Manager,
) func(*modulesMock.MockPersistenceRWContext) {
	return func(mock *modulesMock.MockPersistenceRWContext) {
		accounts := runTimeMgr.GetGenesis().GetPersistenceGenesisState().GetAccs()

		accountsMap := make(map[string]*big.Int)
		for _, acc := range accounts {
			amount, err := typesUtil.StringToBigInt(acc.GetAmount())
			require.NoError(t, err)
			accountsMap[acc.GetAddress()] = amount
		}

		getAccountMocks := func() (accMocks []modules.Account) {
			for addr, amount := range accountsMap {
				ctrl := gomock.NewController(t)
				acc := modulesMock.NewMockAccount(ctrl)
				acc.EXPECT().GetAddress().Return(addr).AnyTimes()
				acc.EXPECT().GetAmount().Return(typesUtil.BigIntToString(amount)).AnyTimes()
				accMocks = append(accMocks, acc)
			}
			return
		}

		mock.EXPECT().GetAllAccounts(gomock.Any()).DoAndReturn(func(height int64) ([]modules.Account, error) {
			return getAccountMocks(), nil
		}).AnyTimes()
		mock.EXPECT().AddAccountAmount(gomock.Any(), gomock.Any()).DoAndReturn(func(addrBz []byte, amountStr string) error {
			amount, err := typesUtil.StringToBigInt(amountStr)
			require.NoError(t, err)

			addr := hex.EncodeToString(addrBz)
			accountsMap[addr].Add(accountsMap[addr], amount)

			return nil
		}).AnyTimes()
		mock.EXPECT().SubtractAccountAmount(gomock.Any(), gomock.Any()).DoAndReturn(func(addrBz []byte, amountStr string) error {
			amount, err := typesUtil.StringToBigInt(amountStr)
			require.NoError(t, err)

			addr := hex.EncodeToString(addrBz)
			accountsMap[addr].Sub(accountsMap[addr], amount)

			return nil
		}).AnyTimes()
		mock.EXPECT().SetAccountAmount(gomock.Any(), gomock.Any()).DoAndReturn(func(addrBz []byte, amountStr string) error {
			amount, err := typesUtil.StringToBigInt(amountStr)
			require.NoError(t, err)

			accountsMap[hex.EncodeToString(addrBz)] = amount

			return nil
		}).AnyTimes()
		mock.EXPECT().GetAccountAmount(gomock.Any(), gomock.Any()).DoAndReturn(func(addrBz []byte, height int64) (string, error) {
			return typesUtil.BigIntToString(accountsMap[hex.EncodeToString(addrBz)]), nil
		}).AnyTimes()
	}
}

// Stateful mocks of the basic pool related functions in the persistence module interface
func withBasePoolMock(
	t *testing.T,
	runTimeMgr *runtime.Manager,
) func(*modulesMock.MockPersistenceRWContext) {
	return func(mock *modulesMock.MockPersistenceRWContext) {
		pools := runTimeMgr.GetGenesis().GetPersistenceGenesisState().GetAccPools()
		poolsMap := make(map[string]*big.Int)
		for _, pool := range pools {
			amount, err := typesUtil.StringToBigInt(pool.GetAmount())
			require.NoError(t, err)
			poolsMap[pool.GetAddress()] = amount
		}

		getPoolsMock := func() (accMocks []modules.Account) {
			for name, amount := range poolsMap {
				ctrl := gomock.NewController(t)
				acc := modulesMock.NewMockAccount(ctrl)
				acc.EXPECT().GetAddress().Return(name).AnyTimes()
				acc.EXPECT().GetAmount().Return(typesUtil.BigIntToString(amount)).AnyTimes()
				accMocks = append(accMocks, acc)
			}
			return
		}

		mock.EXPECT().GetAllPools(gomock.Any()).DoAndReturn(func(height int64) ([]modules.Account, error) {
			return getPoolsMock(), nil
		}).AnyTimes()
		mock.EXPECT().InsertPool(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(name string, address []byte, amountStr string) error {
			amount, err := typesUtil.StringToBigInt(amountStr)
			require.NoError(t, err)

			poolsMap[name] = amount

			return nil
		}).AnyTimes()
		mock.EXPECT().AddPoolAmount(gomock.Any(), gomock.Any()).DoAndReturn(func(name, amountStr string) error {
			amount, err := typesUtil.StringToBigInt(amountStr)
			require.NoError(t, err)

			poolsMap[name].Add(poolsMap[name], amount)

			return nil
		}).AnyTimes()
		mock.EXPECT().SubtractPoolAmount(gomock.Any(), gomock.Any()).DoAndReturn(func(name, amountStr string) error {
			amount, err := typesUtil.StringToBigInt(amountStr)
			require.NoError(t, err)

			poolsMap[name].Sub(poolsMap[name], amount)

			return nil
		}).AnyTimes()
		mock.EXPECT().SetPoolAmount(gomock.Any(), gomock.Any()).DoAndReturn(func(name, amountStr string) error {
			amount, err := typesUtil.StringToBigInt(amountStr)
			require.NoError(t, err)

			poolsMap[name] = amount

			return nil
		}).AnyTimes()
		mock.EXPECT().GetPoolAmount(gomock.Any(), gomock.Any()).DoAndReturn(func(name string, height int64) (string, error) {
			return typesUtil.BigIntToString(poolsMap[name]), nil
		}).AnyTimes()
	}
}
