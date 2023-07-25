package unit_of_work

import (
	"fmt"
	"math/big"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pokt-network/pocket/shared/modules"
	mockModules "github.com/pokt-network/pocket/shared/modules/mocks"
	"github.com/pokt-network/pocket/shared/utils"
)

var DefaultStakeBig = big.NewInt(1000000000000000)

func Test_leaderUtilityUnitOfWork_CreateProposalBlock(t *testing.T) {
	t.Helper()

	type fields struct {
		leaderUOW func(t *testing.T) *leaderUtilityUnitOfWork
	}
	type args struct {
		proposer   []byte
		maxTxBytes uint64
	}
	tests := []struct {
		name          string
		args          args
		fields        fields
		wantStateHash string
		wantTxs       [][]byte
		wantErr       bool
	}{
		{
			name: "should revert a failed block proposal",
			args: args{},
			fields: fields{
				leaderUOW: func(t *testing.T) *leaderUtilityUnitOfWork {
					ctrl := gomock.NewController(t)

					mockUtilityMod := newDefaultMockUtilityModule(t, ctrl)
					mockrwcontext := newDefaultMockRWContext(t, ctrl)

					mockrwcontext.EXPECT().ComputeStateHash().Return("", fmt.Errorf("rollback error"))
					mockbus := mockModules.NewMockBus(ctrl)
					mockbus.EXPECT().GetUtilityModule().Return(mockUtilityMod).AnyTimes()

					luow := NewLeaderUOW(0, mockrwcontext, mockrwcontext)
					luow.SetBus(mockbus)

					return luow
				},
			},
			wantErr: true,
			wantTxs: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			luow := tt.fields.leaderUOW(t)
			_, gotTxs, err := luow.CreateProposalBlock(tt.args.proposer, tt.args.maxTxBytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("leaderUtilityUnitOfWork.CreateProposalBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotTxs, tt.wantTxs) {
				t.Errorf("leaderUtilityUnitOfWork.CreateProposalBlock() gotTxs = %v, want %v", gotTxs, tt.wantTxs)
			}
		})
	}
}

func newDefaultMockRWContext(t *testing.T, ctrl *gomock.Controller) *mockModules.MockPersistenceRWContext {
	mockrwcontext := mockModules.NewMockPersistenceRWContext(ctrl)

	mockrwcontext.EXPECT().SetPoolAmount(gomock.Any(), gomock.Any()).AnyTimes()
	mockrwcontext.EXPECT().RollbackToSavePoint().Times(1)
	mockrwcontext.EXPECT().GetIntParam(gomock.Any(), gomock.Any()).Return(0, nil).AnyTimes()
	mockrwcontext.EXPECT().GetPoolAmount(gomock.Any(), gomock.Any()).Return(utils.BigIntToString(DefaultStakeBig), nil).Times(1)
	mockrwcontext.EXPECT().AddAccountAmount(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockrwcontext.EXPECT().AddPoolAmount(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockrwcontext.EXPECT().GetAppsReadyToUnstake(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	mockrwcontext.EXPECT().GetServicersReadyToUnstake(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	mockrwcontext.EXPECT().GetValidatorsReadyToUnstake(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	mockrwcontext.EXPECT().GetFishermenReadyToUnstake(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	mockrwcontext.EXPECT().SetServicerStatusAndUnstakingHeightIfPausedBefore(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockrwcontext.EXPECT().SetAppStatusAndUnstakingHeightIfPausedBefore(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockrwcontext.EXPECT().SetValidatorsStatusAndUnstakingHeightIfPausedBefore(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockrwcontext.EXPECT().SetFishermanStatusAndUnstakingHeightIfPausedBefore(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	return mockrwcontext
}

func newDefaultMockUtilityModule(t *testing.T, ctrl *gomock.Controller) *mockModules.MockUtilityModule {
	mockUtilityMod := mockModules.NewMockUtilityModule(ctrl)
	testmempool := NewTestingMempool(t)
	mockUtilityMod.EXPECT().GetModuleName().Return(modules.UtilityModuleName).AnyTimes()
	mockUtilityMod.EXPECT().SetBus(gomock.Any()).Return().AnyTimes()
	mockUtilityMod.EXPECT().GetMempool().Return(testmempool).AnyTimes()
	return mockUtilityMod
}
