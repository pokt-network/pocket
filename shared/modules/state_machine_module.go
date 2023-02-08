package modules

import "context"

//go:generate mockgen -source=$GOFILE -destination=./mocks/state_machine_module_mock.go -aux_files=github.com/pokt-network/pocket/shared/modules=module.go

const StateMachineModuleName = "state_machine"

type StateMachineModule interface {
	Module

	AvailableTransitions() []string
	Can(event string) bool
	Cannot(event string) bool
	Current() string
	DeleteMetadata(key string)
	Event(ctx context.Context, event string, args ...interface{}) error
	Is(state string) bool
	Metadata(key string) (interface{}, bool)
	SetMetadata(key string, dataValue interface{})
	SetState(state string)
	Transition() error
}
