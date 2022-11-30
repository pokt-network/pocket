package rpc

import (
	"log"

	"github.com/pokt-network/pocket/shared/modules"
)

var (
	_ modules.RPCModule = &noopRpcModule{}
)

type noopRpcModule struct{}

func (m *noopRpcModule) GetModuleName() string {
	return "noop_rpc_module"
}

func (m *noopRpcModule) Create(runtime modules.RuntimeMgr) (modules.Module, error) {
	return &rpcModule{}, nil
}

func (m *noopRpcModule) SetBus(_ modules.Bus) {}

func (m *noopRpcModule) GetBus() modules.Bus {
	return nil
}

func (m *noopRpcModule) Start() error {
	log.Println("[WARN] RPC server: OFFLINE")
	return nil
}

func (m *noopRpcModule) Stop() error {
	return nil
}

func (m *noopRpcModule) ValidateConfig(_ modules.Config) error {
	return nil
}
