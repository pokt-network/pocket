package types

type Runner interface {
	Sink() chan<- Work
	Done() <-chan uint
}
