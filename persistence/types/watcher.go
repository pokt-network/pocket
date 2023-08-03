package types

import coreTypes "github.com/pokt-network/pocket/shared/core/types"

var _ ProtocolActorSchema = &WatcherSchema{}

type WatcherSchema struct {
	BaseProtocolActorSchema
}

const (
	WatcherTableName            = "watcher"
	WatcherChainsTableName      = "watcher_chains"
	WatcherHeightConstraintName = "watcher_height"
	WatcherChainsConstraintName = "watcher_chain_height"
)

var WatcherActor ProtocolActorSchema = &WatcherSchema{
	BaseProtocolActorSchema: BaseProtocolActorSchema{
		actorType: coreTypes.ActorType_ACTOR_TYPE_WATCHER,

		tableName:       WatcherTableName,
		chainsTableName: WatcherChainsTableName,

		actorSpecificColName: ServiceURLCol,

		heightConstraintName:       WatcherHeightConstraintName,
		chainsHeightConstraintName: WatcherChainsConstraintName,
	},
}
