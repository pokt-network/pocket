// +built test debug

package runtime

import "github.com/pokt-network/pocket/shared/modules"

// WithDebugEventsChannel is used initialize a secondary (debug) bus that receives all the same events
// as the main bus, but does pull events when `GetBusEvent` is called
func WithDebugEventsChannel(eventsChannel modules.EventsChannel) modules.BusOption {
	return func(m modules.Bus) {
		if m, ok := m.(*bus); ok {
			m.debugChannel = eventsChannel
		}
	}
}
