package telemetry_testutil

import (
	"github.com/golang/mock/gomock"
	"github.com/pokt-network/pocket/internal/testutil"
	"github.com/regen-network/gocuke"

	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/mocks"
)

func MinimalTelemetryMock(
	t gocuke.TestingT,
	busMock *mock_modules.MockBus,
) *mock_modules.MockTelemetryModule {
	t.Helper()

	ctrl := gomock.NewController(t)
	telemetryMock := mock_modules.NewMockTelemetryModule(ctrl)

	busMock.EXPECT().GetTelemetryModule().Return(telemetryMock).AnyTimes()

	return telemetryMock
}

func BehavesLikeBaseTelemetryMock(
	t gocuke.TestingT,
	telemetryMock *mock_modules.MockTelemetryModule,
) *mock_modules.MockTelemetryModule {
	t.Helper()

	telemetryMock.EXPECT().Start().Return(nil).AnyTimes()
	telemetryMock.EXPECT().SetBus(gomock.Any()).Return().AnyTimes()
	telemetryMock.EXPECT().GetModuleName().Return(modules.TelemetryModuleName).AnyTimes()

	return telemetryMock
}

func BaseTelemetryMock(
	t gocuke.TestingT,
	busMock *mock_modules.MockBus,
) *mock_modules.MockTelemetryModule {
	t.Helper()

	return testutil.PipeTwoToOne[
		gocuke.TestingT,
		*mock_modules.MockTelemetryModule
	](
		t, MinimalTelemetryMock(t, busMock),
		BehavesLikeBaseTelemetryMock,
		WithEventMetricsAgent,
		WithTimeSeriesAgent,
	)
}

func WithTimeSeriesAgent(
	t gocuke.TestingT,
	telemetryMock *mock_modules.MockTelemetryModule,
) *mock_modules.MockTelemetryModule {
	t.Helper()

	timeSeriesAgentMock := BaseTimeSeriesAgentMock(t)

	telemetryMock.EXPECT().GetTimeSeriesAgent().Return(timeSeriesAgentMock).AnyTimes()
	return telemetryMock
}

func WithEventMetricsAgent(
	t gocuke.TestingT,
	telemetryMock *mock_modules.MockTelemetryModule,
) *mock_modules.MockTelemetryModule {
	t.Helper()

	ctrl := gomock.NewController(t)
	eventMetricsAgentMock := mock_modules.NewMockEventMetricsAgent(ctrl)

	telemetryMock.EXPECT().GetEventMetricsAgent().Return(eventMetricsAgentMock).AnyTimes()
	return telemetryMock
}
