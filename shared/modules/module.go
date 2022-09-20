package modules

// TODO(olshansky): Show an example of `TypicalUsage`
// TODO(drewsky): Add `Create` function; pocket/issues/163
// TODO(drewsky): Do not embed this inside of modules but force it via an implicit cast at compile time
type Module interface {
	InitializableModule
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

type InitializableModule interface {
	GetModuleName() string
	InitConfig(pathToConfigJSON string) (IConfig, error)
	InitGenesis(pathToGenesisJSON string) (IGenesis, error)
}
