package watcher

import (
	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
)

const (
	WatcherModuleName = "watcher"
)

type watcher struct {
	base_modules.IntegrableModule
	logger *modules.Logger
}

var (
	_ modules.WatcherModule = &watcher{}
)

func CreateWatcher(bus modules.Bus, options ...modules.ModuleOption) (modules.WatcherModule, error) {
	m, err := new(watcher).Create(bus, options...)
	if err != nil {
		return nil, err
	}
	return m.(modules.WatcherModule), nil
}

func (*watcher) Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	m := &watcher{}

	for _, option := range options {
		option(m)
	}

	bus.RegisterModule(m)

	m.logger = logger.Global.CreateLoggerForModule(m.GetModuleName())

	return m, nil
}

// TODO: implement this function
func (m *watcher) Start() error {
	m.logger.Info().Msg("🎣 Watcher module started 🎣")
	return nil
}

func (m *watcher) Stop() error {
	m.logger.Info().Msg("🎣 Watcher module stopped 🎣")
	return nil
}

func (m *watcher) GetModuleName() string {
	return WatcherModuleName
}
