package rpc

import (
	"log"

	"github.com/pokt-network/pocket/shared/modules"
)

var _ modules.RPCModule = &noopRpcModule{}

type noopRpcModule struct {
	modules.BaseIntegratableModule
	modules.BaseInterruptableModule
}

func (m *noopRpcModule) GetModuleName() string {
	return "noop_rpc_module"
}

func (m *noopRpcModule) Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	return &rpcModule{}, nil
}

func (m *noopRpcModule) Start() error {
	log.Println("[WARN] RPC server: OFFLINE")
	return nil
}
