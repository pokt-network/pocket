package rpc

import (
	"log"

	// importing because used by code-generated files that are git ignored and to allow go mod tidy and go mod vendor to function properly
	_ "github.com/getkin/kin-openapi/openapi3"
	_ "github.com/labstack/echo/v4"

	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/shared/modules"
)

var (
	_ modules.RPCModule = &rpcModule{}
)

type rpcModule struct {
	bus    modules.Bus
	config *configs.RPCConfig
}

const (
	rpcModuleName = "rpc"
)

func Create(runtime modules.RuntimeMgr) (modules.Module, error) {
	return new(rpcModule).Create(runtime)
}

func (m *rpcModule) Create(runtime modules.RuntimeMgr) (modules.Module, error) {
	rpcCfg := runtime.GetConfig().RPC

	if !rpcCfg.Enabled {
		return &noopRpcModule{}, nil
	}

	return &rpcModule{
		config: rpcCfg,
	}, nil
}

func (u *rpcModule) Start() error {
	go NewRPCServer(u.GetBus()).StartRPC(u.config.Port, u.config.Timeout)
	return nil
}

func (u *rpcModule) Stop() error {
	return nil
}

func (u *rpcModule) GetModuleName() string {
	return rpcModuleName
}

func (u *rpcModule) SetBus(bus modules.Bus) {
	u.bus = bus
}

func (u *rpcModule) GetBus() modules.Bus {
	if u.bus == nil {
		log.Fatalf("Bus is not initialized")
	}
	return u.bus
}
