package events

import (
	core_types "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
)

var _ modules.EventLogger = &EventManager{}

type EventManager struct {
	base_modules.IntegrableModule

	logger *modules.Logger
}

func Create(bus modules.Bus, options ...modules.EventLoggerOption) (modules.EventLogger, error) {
	return new(EventManager).Create(bus, options...)
}

func WithLogger(logger *modules.Logger) modules.EventLoggerOption {
	return func(m modules.EventLogger) {
		if mod, ok := m.(*EventManager); ok {
			mod.logger = logger
		}
	}
}

func (*EventManager) Create(bus modules.Bus, options ...modules.EventLoggerOption) (modules.EventLogger, error) {
	e := &EventManager{}

	for _, option := range options {
		option(e)
	}

	e.logger.Info().Msg("🪵 Creating Event Logger 🪵")

	bus.RegisterModule(e)

	return e, nil
}

func (e *EventManager) GetModuleName() string { return modules.EventLoggerModuleName }

func (e *EventManager) EmitEvent(event *core_types.IBCEvent) error {
	wCtx := e.GetBus().GetPersistenceModule().NewWriteContext()
	defer wCtx.Release()
	return wCtx.SetIBCEvent(event)
}

func (e *EventManager) QueryEvents(topic string, height uint64) ([]*core_types.IBCEvent, error) {
	currHeight := e.GetBus().GetConsensusModule().CurrentHeight()
	rCtx, err := e.GetBus().GetPersistenceModule().NewReadContext(int64(currHeight))
	if err != nil {
		return nil, err
	}
	defer rCtx.Release()
	return rCtx.GetIBCEvents(height, topic)
}
