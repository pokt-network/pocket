package modules

// TODO(olshansky): Show an example of `TypicalUsage`
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
