package modules

// TODO: Show an example of `TypicalUsage`
type Module interface {
	Start() error
	Stop() error

	SetPocketBusMod(BusModule)
	GetBus() BusModule
}
