package types

type Runner interface {
	Sink() chan<- Packet
	Done() <-chan uint
}
