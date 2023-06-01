package telemetry_testutil

import (
	"log"
	"sync"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/regen-network/gocuke"

	"github.com/pokt-network/pocket/shared/modules/mocks"
	"github.com/pokt-network/pocket/telemetry"
)

func BaseEventMetricsAgentMock(t gocuke.TestingT) *mock_modules.MockEventMetricsAgent {
	t.Helper()

	ctrl := gomock.NewController(t)
	eventMetricsAgentMock := mock_modules.NewMockEventMetricsAgent(ctrl)
	eventMetricsAgentMock.EXPECT().EmitEvent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	return eventMetricsAgentMock
}

// TODO_THIS_COMMIT: refactor...
// Events metric mock - Needed to help with proper counts for number of expected network writes
func PrepareEventMetricsAgentMock(t *testing.T, valId string, wg *sync.WaitGroup, expectedNumNetworkWrites int) *mock_modules.MockEventMetricsAgent {
	ctrl := gomock.NewController(t)
	eventMetricsAgentMock := mock_modules.NewMockEventMetricsAgent(ctrl)

	eventMetricsAgentMock.EXPECT().EmitEvent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	eventMetricsAgentMock.EXPECT().EmitEvent(gomock.Any(), gomock.Any(), gomock.Eq(telemetry.P2P_RAINTREE_MESSAGE_EVENT_METRIC_SEND_LABEL), gomock.Any()).Do(func(n, e any, l ...any) {
		log.Printf("[valId: %s] Write\n", valId)
		wg.Done()
	}).Times(expectedNumNetworkWrites)
	eventMetricsAgentMock.EXPECT().EmitEvent(gomock.Any(), gomock.Any(), gomock.Not(telemetry.P2P_RAINTREE_MESSAGE_EVENT_METRIC_SEND_LABEL), gomock.Any()).AnyTimes()

	return eventMetricsAgentMock
}
