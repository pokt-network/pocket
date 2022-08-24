package modules

// TODO(olshansky): Show an example of `TypicalUsage`
// TODO(pocket/issues/163): Add `Create` function
type Module interface {
	IntegratableModule
	InterruptableModule
}

type IntegratableModule interface {
	SetBus(Bus)
	GetBus() Bus
}

type InterruptableModule interface {
	Start() error
	Stop() error
}
