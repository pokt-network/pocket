package p2p_testutil

import (
	"testing"

	"github.com/golang/mock/gomock"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/mocks"
)

// Creates a p2p module mock with mock implementations of some basic functionality
func BaseP2PMock(t *testing.T, eventsChannel modules.EventsChannel) *mock_modules.MockP2PModule {
	ctrl := gomock.NewController(t)
	p2pMock := mock_modules.NewMockP2PModule(ctrl)

	p2pMock.EXPECT().Start().Return(nil).AnyTimes()
	p2pMock.EXPECT().SetBus(gomock.Any()).Return().AnyTimes()
	p2pMock.EXPECT().
		Broadcast(gomock.Any()).
		Do(func(msg *anypb.Any) {
			e := &messaging.PocketEnvelope{Content: msg}
			eventsChannel <- e
		}).
		AnyTimes()
	// CONSIDERATION: Adding a check to not to send message to itself
	p2pMock.EXPECT().
		Send(gomock.Any(), gomock.Any()).
		Do(func(addr crypto.Address, msg *anypb.Any) {
			e := &messaging.PocketEnvelope{Content: msg}
			eventsChannel <- e
		}).
		AnyTimes()
	p2pMock.EXPECT().GetModuleName().Return(modules.P2PModuleName).AnyTimes()
	p2pMock.EXPECT().HandleEvent(gomock.Any()).Return(nil).AnyTimes()

	return p2pMock
}
