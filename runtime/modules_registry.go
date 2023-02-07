package runtime

import (
	"sync"

	"github.com/pokt-network/pocket/shared/modules"
)

var _ modules.ModulesRegistry = &modulesRegistry{}

type modulesRegistry struct {
	m        sync.Mutex
	registry map[string]modules.Module
}

func NewModulesRegistry() *modulesRegistry {
	return &modulesRegistry{
		registry: make(map[string]modules.Module),
	}
}

func (m *modulesRegistry) RegisterModule(module modules.Module) {
	m.m.Lock()
	defer m.m.Unlock()
	m.registry[module.GetModuleName()] = module
}

func (m *modulesRegistry) GetModule(moduleName string) (modules.Module, error) {
	m.m.Lock()
	defer m.m.Unlock()
	if mod, ok := m.registry[moduleName]; ok {
		return mod, nil
	}
	return nil, ErrModuleNotRegistered(moduleName)
}
