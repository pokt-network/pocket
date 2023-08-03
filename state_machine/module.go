package state_machine

// TECHDEBT(#821): Remove the dependency of state sync on FSM, as well as the FSM in general.

import (
	"context"

	"github.com/looplab/fsm"
	"github.com/pokt-network/pocket/logger"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
)

var _ modules.StateMachineModule = &stateMachineModule{}

type stateMachineModule struct {
	base_modules.IntegrableModule
	base_modules.InterruptableModule

	*fsm.FSM
	logger *modules.Logger
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
				Event:         e.Event,
				PreviousState: e.Src,
				NewState:      e.Dst,
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

func (m *stateMachineModule) SendEvent(event coreTypes.StateMachineEvent, args ...any) error {
	return m.Event(context.TODO(), string(event), args)
}

// options

func WithCustomStateMachine(stateMachine *fsm.FSM) modules.ModuleOption {
	return func(m modules.InjectableModule) {
		if m, ok := m.(*stateMachineModule); ok {
			m.FSM = stateMachine
		}
	}
}
