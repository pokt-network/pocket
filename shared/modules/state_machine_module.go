package modules

//go:generate mockgen -destination=./mocks/state_machine_module_mock.go github.com/pokt-network/pocket/shared/modules StateMachineModule

import (
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
)

const StateMachineModuleName = "state_machine"

type StateMachineModule interface {
	Module

	SendEvent(event coreTypes.StateMachineEvent, args ...any) error
}
