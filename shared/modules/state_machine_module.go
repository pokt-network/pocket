package modules

//go:generate mockgen -source=$GOFILE -destination=./mocks/state_machine_module_mock.go -aux_files=github.com/pokt-network/pocket/shared/modules=module.go

import (
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
)

const StateMachineModuleName = "state_machine"

type StateMachineModule interface {
	Module

	SendEvent(event coreTypes.StateMachineEvent, args ...any) error
}
