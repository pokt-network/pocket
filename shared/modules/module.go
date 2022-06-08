package modules

// TODO(olshansky): Show an example of `TypicalUsage`
type Module interface {
	Start() error
	Stop() error

	SetBus(Bus)
	GetBus() Bus
}
