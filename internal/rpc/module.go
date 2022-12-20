package rpc

import (
	"fmt"
	"log"

	// importing because used by code-generated files that are git ignored and to allow go mod tidy and go mod vendor to function properly
	_ "github.com/getkin/kin-openapi/openapi3"
	_ "github.com/labstack/echo/v4"

	"github.com/pokt-network/pocket/internal/shared/modules"
)

var _ modules.RPCModule = &rpcModule{}

type rpcModule struct {
	bus    modules.Bus
	config modules.RPCConfig
}

const (
	rpcModuleName = "rpc"
)

func Create(runtime modules.RuntimeMgr) (modules.Module, error) {
	return new(rpcModule).Create(runtime)
}

func (m *rpcModule) Create(runtime modules.RuntimeMgr) (modules.Module, error) {
	cfg := runtime.GetConfig()
	if err := m.ValidateConfig(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}
	rpcCfg := cfg.GetRPCConfig()

	if !rpcCfg.GetEnabled() {
		return &noopRpcModule{}, nil
	}

	return &rpcModule{
		config: rpcCfg,
	}, nil
}

func (u *rpcModule) Start() error {
	go NewRPCServer(u.GetBus()).StartRPC(u.config.GetPort(), u.config.GetTimeout())
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

func (*rpcModule) ValidateConfig(cfg modules.Config) error {
	// TODO (#334): implement this
	return nil
}
