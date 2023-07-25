package modules

//go:generate mockgen -destination=./mocks/ibc_event_module_mock.go github.com/pokt-network/pocket/shared/modules EventLogger

import (
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
)

const EventLoggerModuleName = "event_logger"

type EventLoggerOption func(EventLogger)

type eventLoggerFactory = FactoryWithOptions[EventLogger, EventLoggerOption]

type EventLogger interface {
	Submodule
	eventLoggerFactory

	EmitEvent(event *coreTypes.IBCEvent) error
	QueryEvents(topic string, height uint64) ([]*coreTypes.IBCEvent, error)
}
