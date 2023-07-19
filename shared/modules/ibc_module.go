package modules

import "google.golang.org/protobuf/types/known/anypb"

//go:generate mockgen -destination=./mocks/ibc_module_mock.go github.com/pokt-network/pocket/shared/modules IBCModule

const IBCModuleName = "ibc"

type IBCModule interface {
	Module

	HandleEvent(*anypb.Any) error
}
