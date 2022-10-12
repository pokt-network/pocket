package modules

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
