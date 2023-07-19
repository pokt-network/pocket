package runtime

import (
	"sync"

	"github.com/pokt-network/pocket/shared/modules"
)

var _ modules.ModulesRegistry = &modulesRegistry{}

type modulesRegistry struct {
	m        sync.Mutex
	registry map[string]modules.InjectableModule
}

func NewModulesRegistry() modules.ModulesRegistry {
	return &modulesRegistry{
		registry: make(map[string]modules.InjectableModule),
	}
}

func (m *modulesRegistry) RegisterModule(module modules.InjectableModule) {
	m.m.Lock()
	defer m.m.Unlock()
	m.registry[module.GetModuleName()] = module
}

func (m *modulesRegistry) GetModule(moduleName string) (modules.InjectableModule, error) {
	m.m.Lock()
	defer m.m.Unlock()
	if mod, ok := m.registry[moduleName]; ok {
		return mod, nil
	}
	return nil, ErrModuleNotRegistered(moduleName)
}
