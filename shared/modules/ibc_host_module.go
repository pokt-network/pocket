package modules

import (
	"github.com/pokt-network/pocket/runtime/configs"
)

//go:generate mockgen -destination=./mocks/ibc_host_module_mock.go github.com/pokt-network/pocket/shared/modules IBCHostSubmodule

const IBCHostSubmoduleName = "ibc_host"

type IBCHostOption func(IBCHostSubmodule)

type ibcHostFactory = FactoryWithConfigAndOptions[IBCHostSubmodule, *configs.IBCHostConfig, IBCHostOption]

// IBCHost is the interface used by the host machine (a Pocket node) to interact with the IBC module
// the host is responsible for managing the IBC state and interacting with consensus in order for
// any IBC packets to be sent to another host on a different chain (via an IBC relayer). The hosts
// are also responsible for receiving any IBC packets from another chain and verifying them through
// the IBC light clients they manage
// https://github.com/cosmos/ibc/tree/main/spec/core/ics-024-host-requirements
type IBCHostSubmodule interface {
	Submodule
	ibcHostFactory

	// GetTimestamp returns the current unix timestamp for the host machine
	GetTimestamp() uint64

	// GetProvableStore returns an instance of a ProvableStore managed by the StoreManager
	GetProvableStore(name string) (ProvableStore, error)
}
