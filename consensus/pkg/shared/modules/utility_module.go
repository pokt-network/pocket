package modules

import (
	"pocket/consensus/pkg/shared/context"
	"pocket/consensus/pkg/types/typespb"
)

type UtilityModule interface {
	PocketModule

	// Message Handling
	HandleTransaction(*context.PocketContext, *typespb.Transaction) error
	HandleEvidence(*context.PocketContext, *typespb.Evidence) error

	// Block Application
	ReapMempool(*context.PocketContext) ([]*typespb.Transaction, error)
	BeginBlock(*context.PocketContext) error
	DeliverTx(*context.PocketContext, *typespb.Transaction) error
	EndBlock(*context.PocketContext) error
}
