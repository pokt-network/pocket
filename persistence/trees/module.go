package trees

import (
	"github.com/pokt-network/pocket/shared/modules"
)

func (*treeStore) Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	m := &treeStore{}

	for _, option := range options {
		option(m)
	}

	bus.RegisterModule(m)

	return m, nil
}

func Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	return new(treeStore).Create(bus, options...)
}

func (t *treeStore) GetModuleName() string {
	return modules.TreeStoreModuleName
}
