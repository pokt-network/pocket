package modules

//go:generate mockgen -source=$GOFILE -destination=./mocks/rpc_module_mock.go -aux_files=github.com/pokt-network/pocket/internal/shared/modules=module.go

type RPCModule interface {
	Module
	ConfigurableModule
}
