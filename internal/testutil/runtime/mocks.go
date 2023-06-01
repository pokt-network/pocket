package runtime_testutil

import (
	"net"
	"strconv"

	"github.com/golang/mock/gomock"
	"github.com/regen-network/gocuke"
	"github.com/stretchr/testify/require"

	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/runtime/configs/types"
	"github.com/pokt-network/pocket/runtime/defaults"
	"github.com/pokt-network/pocket/runtime/genesis"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
	mock_modules "github.com/pokt-network/pocket/shared/modules/mocks"
)

func BaseRuntimeManagerMock(
	t gocuke.TestingT,
	privKey cryptoPocket.PrivateKey,
	serviceURL string,
	genesisState *genesis.GenesisState,
) modules.RuntimeMgr {
	ctrl := gomock.NewController(t)
	runtimeMgrMock := mock_modules.NewMockRuntimeMgr(ctrl)

	hostname, portStr, err := net.SplitHostPort(serviceURL)
	require.NoError(t, err)

	port, err := strconv.Atoi(portStr)
	require.NoError(t, err)

	cfg := &configs.Config{
		RootDirectory: "",
		// TODO: need this?
		//PrivateKey:    privKey.String(),
		P2P: &configs.P2PConfig{
			Hostname:       hostname,
			PrivateKey:     privKey.String(),
			Port:           uint32(port),
			ConnectionType: types.ConnectionType_EmptyConnection,
			MaxNonces:      defaults.DefaultP2PMaxNonces,
		},
	}

	runtimeMgrMock.EXPECT().GetConfig().Return(cfg).AnyTimes()
	runtimeMgrMock.EXPECT().GetGenesis().Return(genesisState).AnyTimes()
	return runtimeMgrMock
}
