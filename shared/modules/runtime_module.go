package modules

type Runtime interface {
	GetConfig() Config
	GetGenesis() GenesisState
	ShouldUseRandomPK() bool // INVESTIGATE: look into how we can remove this from the runtime interface
}
