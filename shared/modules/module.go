package modules

import (
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"google.golang.org/protobuf/types/known/anypb"
)

type Module interface {
	InitializableModule
	IntegratableModule
	InterruptableModule
}

type IntegratableModule interface {
	SetBus(Bus)
	GetBus() Bus
}

type InterruptableModule interface {
	Start() error
	Stop() error
}

type InitializableModule interface {
	GetModuleName() string
	Create(bus Bus) (Module, error)
}

type KeyholderModule interface {
	GetPrivateKey() (cryptoPocket.PrivateKey, error)
}

type P2PAddressableModule interface {
	GetP2PAddress() cryptoPocket.Address
}

type ObservableModule interface {
	InitLogger()
	GetLogger() Logger
}

// This interface represents functions built for an intermediate solution towards seperation consensus and pacemaker modules
// This functions should be only called by the PaceMaker module.
type PaceMakerAccessModule interface {
	//Pacemaker Consensus interaction modules
	ClearLeaderMessagesPool()
	SetHeight(uint64)
	SetRound(uint64)
	SetStep(uint64)
	ResetForNewHeight()
	ReleaseUtilityContext() error
	BroadcastMessageToNodes(*anypb.Any)
}
