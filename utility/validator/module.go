package validator

import (
	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
)

const (
	ValidatorModuleName = "validator"
)

type validator struct {
	base_modules.IntegrableModule
	logger *modules.Logger
}

var (
	_ modules.ValidatorModule = &validator{}
)

func CreateValidator(bus modules.Bus, options ...modules.ModuleOption) (modules.ValidatorModule, error) {
	m, err := new(validator).Create(bus, options...)
	if err != nil {
		return nil, err
	}
	return m.(modules.ValidatorModule), nil
}

func (*validator) Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	m := &validator{}

	for _, option := range options {
		option(m)
	}

	bus.RegisterModule(m)

	m.logger = logger.Global.CreateLoggerForModule(m.GetModuleName())

	return m, nil
}

// TODO: implement this function
func (m *validator) Start() error {
	m.logger.Info().Msg("📝 Validator module started 📝")
	return nil
}

func (m *validator) Stop() error {
	m.logger.Info().Msg("📝 Validator module stopped 📝")
	return nil
}

func (m *validator) GetModuleName() string {
	return ValidatorModuleName
}
