package state_machine

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
	base_modules.IntegratableModule
	base_modules.InterruptableModule

	*fsm.FSM
	logger *modules.Logger
	// debugChannels is only used for testing purposes, events pushed to it are emitted in testing
	debugChannels []modules.EventsChannel
}

func Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	return new(stateMachineModule).Create(bus, options...)
}

func (*stateMachineModule) Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	m := &stateMachineModule{
		logger:        logger.Global.CreateLoggerForModule(modules.StateMachineModuleName),
		debugChannels: make([]modules.EventsChannel, 0),
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
			for _, channel := range m.debugChannels {
				channel <- newStateMachineTransitionEvent
			}
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
	return func(m modules.InitializableModule) {
		if m, ok := m.(*stateMachineModule); ok {
			m.FSM = stateMachine
		}
	}
}

// WithDebugEventsChannel is used for testing purposes. It allows us to capture the events
// from the FSM and publish them to debug channel for testing.
func WithDebugEventsChannel(eventsChannel modules.EventsChannel) modules.ModuleOption {
	return func(m modules.InitializableModule) {
		if m, ok := m.(*stateMachineModule); ok {
			m.debugChannels = append(m.debugChannels, eventsChannel)
		}
	}
}
