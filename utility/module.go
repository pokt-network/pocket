package utility

import (
	"errors"

	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/shared/mempool"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
	"github.com/pokt-network/pocket/utility/fisherman"
	fisherman_module "github.com/pokt-network/pocket/utility/fisherman"
	portal_module "github.com/pokt-network/pocket/utility/portal"
	servicer_module "github.com/pokt-network/pocket/utility/servicer"
	"github.com/pokt-network/pocket/utility/types"
	validator_module "github.com/pokt-network/pocket/utility/validator"
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

	actorModules []modules.Module

	validator *validator_module.ValidatorModule
	servicer  *servicer_module.ServicerModule
	fisherman *fisherman_module.FishermanModule
	portal    *portal_module.PortalModule
}

func Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	return new(utilityModule).Create(bus, options...)
}

func (*utilityModule) Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	m := &utilityModule{
		actorModules: []modules.Module{},
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

	return m, enableActorModules(cfg, m, bus)
}

// enableActorModules enables the actor-specific modules and adds them to the utility module's actorModules to be started later.
func enableActorModules(cfg *configs.Config, m *utilityModule, bus modules.Bus) error {
	fishermanCfg := cfg.Fisherman
	servicerCfg := cfg.Servicer
	validatorCfg := cfg.Validator
	portalCfg := cfg.Portal

	if servicerCfg.Enabled {
		servicer, err := servicer_module.CreateServicer(bus)
		if err != nil {
			m.logger.Error().Err(err).Msg("failed to create servicer module")
			return err
		}
		m.servicer = &servicer
		m.actorModules = append(m.actorModules, servicer)
	}

	if fishermanCfg.Enabled {
		fisherman, err := fisherman.CreateFisherman(bus)
		if err != nil {
			m.logger.Error().Err(err).Msg("failed to create fisherman module")
			return err
		}
		m.fisherman = &fisherman
		m.actorModules = append(m.actorModules, fisherman)
	}

	if validatorCfg.Enabled {
		validator, err := validator_module.CreateValidator(bus)
		if err != nil {
			m.logger.Error().Err(err).Msg("failed to create validator module")
			return err
		}
		m.validator = &validator
		m.actorModules = append(m.actorModules, validator)
	}

	if portalCfg.Enabled {
		portal, err := portal_module.CreatePortal(bus)
		if err != nil {
			m.logger.Error().Err(err).Msg("failed to create portal module")
			return err
		}
		m.portal = &portal
		m.actorModules = append(m.actorModules, portal)
	}

	actors := m.GetActorModuleNames()
	if len(m.actorModules) > 1 {
		// only case where this is allowed is if the node is a validator and a servicer
		if !(validatorCfg.Enabled && servicerCfg.Enabled) {
			m.logger.Error().Strs("actors", actors).Msg(ErrInvalidActorsEnabled)
			m.actorModules = []modules.Module{} // reset the actorModules
			return errors.New(ErrInvalidActorsEnabled)
		}
	} else {
		m.logger.Info().Strs("actors", actors).Msg("Node actors enabled")
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

func (u *utilityModule) GetLogger() *modules.Logger {
	return u.logger
}

func (u *utilityModule) GetActorModules() []modules.Module {
	return u.actorModules
}

func (u *utilityModule) GetActorModuleNames() []string {
	names := []string{}
	for _, submodule := range u.actorModules {
		names = append(names, submodule.GetModuleName())
	}
	return names
}
