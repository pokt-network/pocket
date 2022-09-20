package modules

type Runtime interface {
	GetConfig() Config
	GetGenesis() GenesisState
}
