package modules

//go:generate mockgen -destination=./mocks/servicer_module_mock.go github.com/pokt-network/pocket/shared/modules ServicerModule

import (
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
)

const (
	ServicerModuleName = "servicer"
)

type ServicerModule interface {
	Module
	HandleRelay(*coreTypes.Relay) (*coreTypes.RelayResponse, error)
}
