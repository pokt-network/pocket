package telemetry_testutil

import (
	"github.com/golang/mock/gomock"
	"github.com/regen-network/gocuke"
	"sync"

	"github.com/pokt-network/pocket/shared/modules/mocks"
	"github.com/pokt-network/pocket/telemetry"
)

func WithP2PIntegrationEvents(
	t gocuke.TestingT,
	eventMetricsAgentMock *mock_modules.MockEventMetricsAgent,
) *mock_modules.MockEventMetricsAgent {
	t.Helper()

	eventMetricsAgentMock.EXPECT().EmitEvent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	// TODO_THIS_COMMIT: remove v -- may represent failure condition w/ reused nonces..
	eventMetricsAgentMock.EXPECT().EmitEvent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	// END TODO

	eventMetricsAgentMock.EXPECT().EmitEvent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	return eventMetricsAgentMock
}

// TODO_THIS_COMMIT: refactor...
// Events metric mock - Needed to help with proper counts for number of expected network writes
func PrepareEventMetricsAgentMock(t gocuke.TestingT, valId string, wg *sync.WaitGroup, expectedNumNetworkWrites int) *mock_modules.MockEventMetricsAgent {
	ctrl := gomock.NewController(t)
	eventMetricsAgentMock := mock_modules.NewMockEventMetricsAgent(ctrl)

	// TODO_THIS_COMMIT: remove
	logEvent := func(n, e string, l ...any) {
		//t.Logf("n: %s, e: %s, l: %v\n", n, e, l)
	}

	eventMetricsAgentMock.EXPECT().EmitEvent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Do(logEvent).AnyTimes()
	eventMetricsAgentMock.EXPECT().EmitEvent(gomock.Any(), gomock.Any(), gomock.Eq(telemetry.P2P_RAINTREE_MESSAGE_EVENT_METRIC_SEND_LABEL), gomock.Any()).Do(func(n, e any, l ...any) {
		t.Logf("[valId: %s] Write", valId)
		wg.Done()
	}).Do(logEvent).Times(expectedNumNetworkWrites)
	eventMetricsAgentMock.EXPECT().EmitEvent(gomock.Any(), gomock.Any(), gomock.Not(telemetry.P2P_RAINTREE_MESSAGE_EVENT_METRIC_SEND_LABEL), gomock.Any()).Do(logEvent).AnyTimes()

	return eventMetricsAgentMock
}

func WhyEventMetricsAgentMock(
	t gocuke.TestingT,
	eventMetricsAgentMock *mock_modules.MockEventMetricsAgent,
	valId string,
	wg *sync.WaitGroup,
	//handler func(namespace, eventName string, labels ...any),
	expectedNumNetworkWrites int,
) *mock_modules.MockEventMetricsAgent {
	// TODO_THIS_COMMIT: remove
	logEvent := func(n, e string, l ...any) {
		//t.Logf("n: %s, e: %s, l: %v\n", n, e, l)
	}

	eventMetricsAgentMock.EXPECT().EmitEvent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Do(logEvent).AnyTimes()
	eventMetricsAgentMock.EXPECT().EmitEvent(gomock.Any(), gomock.Any(), gomock.Eq(telemetry.P2P_RAINTREE_MESSAGE_EVENT_METRIC_SEND_LABEL), gomock.Any()).Do(func(n, e any, l ...any) {
		t.Logf("[valId: %s] Write", valId)
		wg.Done()
	}).Do(logEvent).Times(expectedNumNetworkWrites)
	eventMetricsAgentMock.EXPECT().EmitEvent(gomock.Any(), gomock.Any(), gomock.Not(telemetry.P2P_RAINTREE_MESSAGE_EVENT_METRIC_SEND_LABEL), gomock.Any()).Do(logEvent).AnyTimes()

	return eventMetricsAgentMock
}

func EventMetricsAgentMockWithHandler(
	t gocuke.TestingT,
	label string,
	// TODO_THIS_COMMIT: consider refactoring as a type
	handler func(namespace, eventName string, labels ...any),
	times int,
) *mock_modules.MockEventMetricsAgent {
	t.Helper()

	ctrl := gomock.NewController(t)
	eventMetricsAgentMock := mock_modules.NewMockEventMetricsAgent(ctrl)

	return WithEventMetricsHandler(t, eventMetricsAgentMock, label, handler, times)
}

func WithEventMetricsHandler(
	t gocuke.TestingT,
	eventMetricsAgentMock *mock_modules.MockEventMetricsAgent,
	label string,
	handler func(namespace, eventName string, labels ...any),
	times int,
) *mock_modules.MockEventMetricsAgent {
	t.Helper()

	//eventMetricsAgentMock.EXPECT().EmitEvent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Do(logEvent).AnyTimes()
	//eventMetricsAgentMock.EXPECT().EmitEvent(gomock.Any(), gomock.Any(), gomock.Eq(telemetry.P2P_RAINTREE_MESSAGE_EVENT_METRIC_SEND_LABEL), gomock.Any()).Do(func(n, e any, l ...any) {
	//	t.Logf("[valId: %s] Write", valId)
	//	wg.Done()
	//}).Do(logEvent).Times(expectedNumNetworkWrites)
	//eventMetricsAgentMock.EXPECT().EmitEvent(gomock.Any(), gomock.Any(), gomock.Not(telemetry.P2P_RAINTREE_MESSAGE_EVENT_METRIC_SEND_LABEL), gomock.Any()).Do(logEvent).AnyTimes()

	// TODO_THIS_COMMIT: scrutinize these & their order
	eventMetricsAgentMock.EXPECT().EmitEvent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	eventMetricsAgentMock.EXPECT().EmitEvent(gomock.Any(), gomock.Any(), gomock.Eq(label), gomock.Any()).Do(handler).Times(times)
	// TODO_THIS_COMMIT: is this really needed?
	eventMetricsAgentMock.EXPECT().EmitEvent(gomock.Any(), gomock.Any(), gomock.Not(label), gomock.Any()).AnyTimes()

	return eventMetricsAgentMock
}
