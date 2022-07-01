package telemetry

import (
	"fmt"

	"github.com/pokt-network/pocket/shared/config"
	"github.com/pokt-network/pocket/shared/logging"
	"github.com/pokt-network/pocket/shared/modules"
)

var (
	_ modules.TelemetryModule   = &PrometheusStdLoggerModule{}
	_ modules.EventMetricsAgent = &PrometheusStdLoggerModule{}
	_ modules.TimeSeriesAgent   = &PrometheusStdLoggerModule{}
)

type PrometheusStdLoggerModule struct {
	PrometheusTelemetryModule
	namespacedLoggers map[logging.Namespace]logging.Logger
}

func CreatePrometheusStdLoggerModule(cfg *config.Config) (*PrometheusStdLoggerModule, error) {
	promTelemetryMod, err := CreatePrometheusTelemetryModule(cfg)
	if err != nil {
		return nil, err
	}

	return &PrometheusStdLoggerModule{
		PrometheusTelemetryModule: *promTelemetryMod,
		namespacedLoggers:         make(map[logging.Namespace]logging.Logger, 0),
	}, nil
}

func (psl *PrometheusStdLoggerModule) LoggerGet(namespace logging.Namespace) logging.Logger {
	if logger, ok := psl.namespacedLoggers[namespace]; ok {
		return logger
	} else {
		logging.GetGlobalLogger().Warn("PrometheusStdLoggerModule.LoggerGet: no logger is registered for the provided namespace (%s)", namespace)
		logging.GetGlobalLogger().Warn("PrometheusStdLoggerModule.LoggerGet: will use the global logger for the non registered namespace (%s)", namespace)
		return nil
	}
}

func (psl *PrometheusStdLoggerModule) LoggerRegister(namespace logging.Namespace, level logging.LogLevel) error {
	if _, ok := psl.namespacedLoggers[namespace]; ok {
		return fmt.Errorf("a logger is already registered for this namespace (%s)", namespace)
	}

	logger := logging.CreateStdLogger(level)

	logger.SetNamespace(namespace)
	logger.SetLevel(level)

	psl.namespacedLoggers[namespace] = logger
}
