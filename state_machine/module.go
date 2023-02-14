package state_machine

import (
	"context"

	"github.com/looplab/fsm"
	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
)

var _ modules.StateMachineModule = &stateMachineModule{}

type stateMachineModule struct {
	base_modules.IntegratableModule
	base_modules.InterruptableModule

	*fsm.FSM
	logger modules.Logger
}

func Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	return new(stateMachineModule).Create(bus, options...)
}

func (*stateMachineModule) Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	m := &stateMachineModule{
		logger: logger.Global.CreateLoggerForModule(modules.StateMachineModuleName),
	}

	m.FSM = NewNodeFSM(&fsm.Callbacks{
		"enter_state": func(_ context.Context, e *fsm.Event) {
			m.logger.Info().
				Str("event", e.Event).
				Str("sourceState", e.Src).
				Msgf("entering state %s", e.Dst)

			newStateMachineTransitionEvent, err := messaging.PackMessage(&messaging.StateMachineTransitionEvent{
				Event: e.Event,
				Src:   e.Src,
				Dst:   e.Dst,
			})
			if err != nil {
				m.logger.Fatal().Err(err).Msg("failed to pack state machine transition event")
			}

			bus.PublishEventToBus(newStateMachineTransitionEvent)
		},
	})

	for _, option := range options {
		option(m)
	}

	bus.RegisterModule(m)

	return m, nil
}

func (m *stateMachineModule) GetModuleName() string {
	return modules.StateMachineModuleName
}

// options

func WithCustomStateMachine(stateMachine *fsm.FSM) modules.ModuleOption {
	return func(m modules.InitializableModule) {
		if m, ok := m.(*stateMachineModule); ok {
			m.FSM = stateMachine
		}
	}
}
