package modules

//go:generate mockgen -destination=./mocks/modules_registry_mock.go github.com/pokt-network/pocket/shared/modules ModulesRegistry

type ModulesRegistry interface {
	// RegisterModule registers a Module with the ModuleRegistry
	RegisterModule(module InjectableModule)
	// GetModule returns a Module by name or nil if not found in the ModuleRegistry
	GetModule(moduleName string) (InjectableModule, error)
}
