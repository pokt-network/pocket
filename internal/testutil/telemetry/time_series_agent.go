package telemetry_testutil

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/regen-network/gocuke"

	"github.com/pokt-network/pocket/shared/modules/mocks"
)

func BaseTimeSeriesAgentMock(t gocuke.TestingT) *mock_modules.MockTimeSeriesAgent {
	t.Helper()

	ctrl := gomock.NewController(t)
	timeSeriesAgentMock := mock_modules.NewMockTimeSeriesAgent(ctrl)
	timeSeriesAgentMock.EXPECT().CounterRegister(gomock.Any(), gomock.Any()).AnyTimes()
	timeSeriesAgentMock.EXPECT().CounterIncrement(gomock.Any()).AnyTimes()
	return timeSeriesAgentMock
}

// Noop mock - no specific business logic to tend to in the timeseries agent mock
func NoopTelemetryTimeSeriesAgentMock(t *testing.T) *mock_modules.MockTimeSeriesAgent {
	ctrl := gomock.NewController(t)
	timeseriesAgentMock := mock_modules.NewMockTimeSeriesAgent(ctrl)

	timeseriesAgentMock.EXPECT().CounterRegister(gomock.Any(), gomock.Any()).AnyTimes()
	timeseriesAgentMock.EXPECT().CounterIncrement(gomock.Any()).AnyTimes()

	return timeseriesAgentMock
}
