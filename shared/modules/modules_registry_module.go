package modules

//go:generate mockgen -source=$GOFILE -destination=./mocks/modules_registry_mock.go -aux_files=github.com/pokt-network/pocket/shared/modules=module.go

type ModulesRegistry interface {
	// RegisterModule registers a Module with the ModuleRegistry
	RegisterModule(module Module)
	// GetModuleByName returns a Module by name or nil if not found in the ModuleRegistry
	GetModule(moduleName string) (Module, error)
}
