package modules

// TODO: Show an example of `TypicalUsage`
type Module interface {
	Start() error
	Stop() error

	SetBus(Bus)
	GetBus() Bus
}
