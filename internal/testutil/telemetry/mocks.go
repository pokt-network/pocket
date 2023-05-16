package telemetry_testutil

import (
	"github.com/golang/mock/gomock"
	"github.com/regen-network/gocuke"

	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/mocks"
)

func MinimalTelemetryMock(
	t gocuke.TestingT,
	busMock *mock_modules.MockBus,
) modules.TelemetryModule {
	t.Helper()

	ctrl := gomock.NewController(t)
	telemetryMock := mock_modules.NewMockTelemetryModule(ctrl)

	telemetryMock.EXPECT().Start().Return(nil).AnyTimes()
	telemetryMock.EXPECT().SetBus(gomock.Any()).Return().AnyTimes()
	// TODO_THIS_COMMIT: which one ^ v ?
	//telemetryMock.EXPECT().SetBus(busMock).Return().AnyTimes()
	telemetryMock.EXPECT().GetModuleName().Return(modules.TelemetryModuleName).AnyTimes()
	busMock.EXPECT().GetTelemetryModule().Return(telemetryMock).AnyTimes()
	//busMock.RegisterModule(telemetryMock)

	return telemetryMock
}

func BaseTelemetryMock(
	t gocuke.TestingT,
	busMock *mock_modules.MockBus,
) modules.TelemetryModule {
	t.Helper()
	return WithTimeSeriesAgent(t, WithEventMetricsAgent(t, MinimalTelemetryMock(t, busMock)))
}

func WithTimeSeriesAgent(t gocuke.TestingT, telemetryMod modules.TelemetryModule) *mock_modules.MockTelemetryModule {
	t.Helper()

	telemetryMock := telemetryMod.(*mock_modules.MockTelemetryModule)
	timeSeriesAgentMock := BaseTimeSeriesAgentMock(t)

	telemetryMock.EXPECT().GetTimeSeriesAgent().Return(timeSeriesAgentMock).AnyTimes()
	return telemetryMock
}

func WithEventMetricsAgent(t gocuke.TestingT, telemetryMod modules.TelemetryModule) modules.TelemetryModule {
	t.Helper()

	telemetryMock := telemetryMod.(*mock_modules.MockTelemetryModule)
	eventMetricsAgentMock := BaseEventMetricsAgentMock(t)

	telemetryMock.EXPECT().GetEventMetricsAgent().Return(eventMetricsAgentMock).AnyTimes()
	return telemetryMock
}
