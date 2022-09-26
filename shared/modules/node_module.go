package modules

//go:generate mockgen -source=$GOFILE -destination=./mocks/node_module_mock.go -aux_files=github.com/pokt-network/pocket/shared/modules=module.go

type NodeModule interface {
	Module
	P2PAddressableModule
}
