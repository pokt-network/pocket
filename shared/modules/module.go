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

// TODO(@okdas): this should also initialize metrics-related stuff, just the logger for now
// InitMetrics(pathToConfigJSON string)
// InitTracing(pathToConfigJSON string)
type ObservableModule interface {
	InitLogger(pathToConfigJSON string)
	GetLogger() Logger
}
