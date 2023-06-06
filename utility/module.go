package utility

import (
	"errors"

	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/shared/mempool"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
	"github.com/pokt-network/pocket/utility/fisherman"
	"github.com/pokt-network/pocket/utility/servicer"
	"github.com/pokt-network/pocket/utility/types"
	"github.com/pokt-network/pocket/utility/validator"
)

const (
	ErrInvalidActorsEnabled = "invalid actors combination enabled"
)

var (
	_ modules.UtilityModule = &utilityModule{}
)

type utilityModule struct {
	base_modules.IntegratableModule

	logger *modules.Logger

	config *configs.UtilityConfig

	mempool mempool.TXMempool

	actorModules map[string]modules.Module
}

func Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	return new(utilityModule).Create(bus, options...)
}

func (*utilityModule) Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	m := &utilityModule{
		actorModules: map[string]modules.Module{},
	}

	for _, option := range options {
		option(m)
	}

	bus.RegisterModule(m)

	runtimeMgr := bus.GetRuntimeMgr()

	cfg := runtimeMgr.GetConfig()
	utilityCfg := cfg.Utility

	m.config = utilityCfg
	m.mempool = types.NewTxFIFOMempool(utilityCfg.MaxMempoolTransactionBytes, utilityCfg.MaxMempoolTransactions)
	m.logger = logger.Global.CreateLoggerForModule(m.GetModuleName())

	if err := m.enableActorModules(cfg); err != nil {
		return m, err
	}

	return m, nil
}

// enableActorModules enables the actor-specific modules and adds them to the utility module's actorModules to be started later.
func (u *utilityModule) enableActorModules(cfg *configs.Config) error {
	fishermanCfg := cfg.Fisherman
	servicerCfg := cfg.Servicer
	validatorCfg := cfg.Validator

	if servicerCfg.Enabled {
		s, err := servicer.CreateServicer(u.GetBus())
		if err != nil {
			u.logger.Error().Err(err).Msg("failed to create servicer module")
			return err
		}
		u.actorModules[s.GetModuleName()] = s
	}

	if fishermanCfg.Enabled {
		f, err := fisherman.CreateFisherman(u.GetBus())
		if err != nil {
			u.logger.Error().Err(err).Msg("failed to create fisherman module")
			return err
		}
		u.actorModules[f.GetModuleName()] = f
	}

	if validatorCfg.Enabled {
		v, err := validator.CreateValidator(u.GetBus())
		if err != nil {
			u.logger.Error().Err(err).Msg("failed to create validator module")
			return err
		}
		u.actorModules[v.GetModuleName()] = v
	}

	if err := u.validateActorModuleExclusivity(cfg); err != nil {
		return err
	}

	return nil
}

func (u *utilityModule) Start() error {
	// start the actorModules
	for _, actorModule := range u.actorModules {
		if err := actorModule.Start(); err != nil {
			u.logger.Error().Err(err).Msgf("failed to start %s", actorModule.GetModuleName())
			return err
		}
	}

	return nil
}

func (u *utilityModule) Stop() error {
	// stop the actorModules
	for _, actorModule := range u.actorModules {
		if err := actorModule.Stop(); err != nil {
			u.logger.Error().Err(err).Msgf("failed to stop %s", actorModule.GetModuleName())
			return err
		}
	}

	return nil
}

func (u *utilityModule) GetModuleName() string {
	return modules.UtilityModuleName
}

func (u *utilityModule) GetMempool() mempool.TXMempool {
	return u.mempool
}

func (u *utilityModule) GetActorModules() map[string]modules.Module {
	return u.actorModules
}

func (u *utilityModule) GetServicerModule() modules.ServicerModule {
	return u.actorModules[servicer.ServicerModuleName].(modules.ServicerModule)
}

func (u *utilityModule) GetFishermanModule() modules.FishermanModule {
	return u.actorModules[fisherman.FishermanModuleName].(modules.FishermanModule)
}

func (u *utilityModule) GetValidatorModule() modules.ValidatorModule {
	return u.actorModules[validator.ValidatorModuleName].(modules.ValidatorModule)
}

// validateActorModuleExclusivity validates that the actor modules are enabled in a valid combination.
// TODO: There are probably more rules that need to be added here.
func (u *utilityModule) validateActorModuleExclusivity(cfg *configs.Config) error {
	servicerCfg := cfg.Servicer
	validatorCfg := cfg.Validator
	actors := []string{}
	for _, submodule := range u.actorModules {
		actors = append(actors, submodule.GetModuleName())
	}

	if len(u.actorModules) > 1 {
		// only case where this is allowed is if the node is a validator and a servicer
		isVal := (validatorCfg != nil && validatorCfg.Enabled)
		isServ := (servicerCfg != nil && servicerCfg.Enabled)
		if !isVal || !isServ {
			u.logger.Error().Strs("actors", actors).Msg(ErrInvalidActorsEnabled)
			u.actorModules = map[string]modules.Module{}
			return errors.New(ErrInvalidActorsEnabled)
		}
	}

	u.logger.Info().Strs("actors", actors).Msg("Node actors enabled")

	return nil
}
