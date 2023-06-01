package consensus_testutil

import (
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/mocks"
)

// Consensus mock - only needed for validatorMap access
func PrepareConsensusMock(t *testing.T, busMock *mock_modules.MockBus) *mock_modules.MockConsensusModule {
	ctrl := gomock.NewController(t)
	consensusMock := mock_modules.NewMockConsensusModule(ctrl)
	consensusMock.EXPECT().CurrentHeight().Return(uint64(1)).AnyTimes()

	consensusMock.EXPECT().GetBus().Return(busMock).AnyTimes()
	consensusMock.EXPECT().SetBus(busMock).AnyTimes()
	consensusMock.EXPECT().GetModuleName().Return(modules.ConsensusModuleName).AnyTimes()
	busMock.RegisterModule(consensusMock)

	return consensusMock
}
