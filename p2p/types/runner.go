package types

// TODO(derrandz): Document what this is used for
type Runner interface {
	Sink() chan<- Packet
	Done() <-chan uint
}
