package state_machine

import (
	"context"
	"fmt"

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
	logger        *modules.Logger
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
			fmt.Println("Event bus in state machine: ", bus.GetEventBus())
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

// TODO_IN_THIS_COMMIT(gohkan): make sure to document that this is used for debugging purposes.
// We do not want to ever mock the FSM in unit tests because it drives the nodes state and must
// be use as is. However, we need to capture the events form a variety of different nodes.
func WithDebugEventsChannel(eventsChannel modules.EventsChannel) modules.ModuleOption {
	return func(m modules.InitializableModule) {
		if m, ok := m.(*stateMachineModule); ok {
			m.debugChannels = append(m.debugChannels, eventsChannel)
		}
	}
}
