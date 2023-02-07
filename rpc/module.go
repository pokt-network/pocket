package rpc

import (

	// importing because used by code-generated files that are git ignored and to allow go mod tidy and go mod vendor to function properly
	_ "github.com/getkin/kin-openapi/openapi3"
	_ "github.com/labstack/echo/v4"

	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/shared/modules"
)

var _ modules.RPCModule = &rpcModule{}

type rpcModule struct {
	modules.BaseIntegratableModule
	modules.BaseInterruptableModule

	logger modules.Logger
	config *configs.RPCConfig
}

func Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	return new(rpcModule).Create(bus, options...)
}

func (*rpcModule) Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	runtimeMgr := bus.GetRuntimeMgr()
	cfg := runtimeMgr.GetConfig()
	rpcCfg := cfg.RPC
	m := modules.RPCModule(&rpcModule{
		config: rpcCfg,
	})
	if !rpcCfg.Enabled {
		m = &noopRpcModule{}
	}
	bus.RegisterModule(m)

	return m, nil
}

func (u *rpcModule) Start() error {
	u.logger = logger.Global.CreateLoggerForModule(u.GetModuleName())
	go NewRPCServer(u.GetBus()).StartRPC(u.config.Port, u.config.Timeout, &u.logger)
	return nil
}

func (u *rpcModule) GetModuleName() string {
	return modules.RPCModuleName
}
