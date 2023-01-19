package modules

//go:generate mockgen -source=$GOFILE -destination=./mocks/rpc_module_mock.go -aux_files=github.com/pokt-network/pocket/shared/modules=module.go

const RPCModuleName = "rpc"

type RPCModule interface {
	Module
}
