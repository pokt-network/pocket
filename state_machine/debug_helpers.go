// +built test debug

package state_machine

import "github.com/pokt-network/pocket/shared/modules"

// WithDebugEventsChannel is used for testing purposes only. It allows us to capture the events
// from the FSM and publish them to debug channel for testing.
func WithDebugEventsChannel(eventsChannel modules.EventsChannel) modules.ModuleOption {
	return func(m modules.InitializableModule) {
		if m, ok := m.(*stateMachineModule); ok {
			m.debugChannels = append(m.debugChannels, eventsChannel)
		}
	}
}
