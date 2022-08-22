package modules

// TODO(olshansky): Show an example of `TypicalUsage`
// TODO(pocket/issues/163): Add `Create` function
// TODO(olshansky): Do not embed this inside of modules but force it via an implicit cast at compile time
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
