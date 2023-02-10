package rpc

import (
	"log"

	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
)

var _ modules.RPCModule = &noopRpcModule{}

type noopRpcModule struct {
	base_modules.IntegratableModule
	base_modules.InterruptableModule
}

func (m *noopRpcModule) GetModuleName() string {
	return "noop_rpc_module"
}

func (m *noopRpcModule) Create(bus modules.Bus, _ ...modules.ModuleOption) (modules.Module, error) {
	return &rpcModule{}, nil
}

func (m *noopRpcModule) Start() error {
	log.Println("[WARN] RPC server: OFFLINE")
	return nil
}
