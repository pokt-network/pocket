package modules

//go:generate mockgen -destination=./mocks/rpc_module_mock.go github.com/pokt-network/pocket/shared/modules RPCModule

const RPCModuleName = "rpc"

type RPCModule interface {
	Module
}
