package validator

import (
	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
)

const (
	ValidatorModuleName = "validator"
)

type ValidatorModule interface {
	modules.Module

	// TODO: add validator module functions here
	ValidatorUtility
}

// TODO_IN_THIS_COMMIT: exists to help with type assertions, drop once you've added your own functions
type ValidatorUtility interface {
	Stake()
}

type validator struct {
	base_modules.IntegratableModule
	logger *modules.Logger
}

// type assertions for validator module
var (
	_ ValidatorModule = &validator{}
	// TODO: add validator module functions here
)

func CreateValidator(bus modules.Bus, options ...modules.ModuleOption) (ValidatorModule, error) {
	m, err := new(validator).Create(bus, options...)
	if err != nil {
		return nil, err
	}
	return m.(ValidatorModule), nil
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

func (m *validator) Stake() {
	m.logger.Info().Msg("📝 Validator module staked 📝")
}
