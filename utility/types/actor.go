package types

import (
	"log"

	"github.com/pokt-network/pocket/shared/types/genesis"
)

// TODO(andrew): Moving this into a proto file enum (impacts everything) and making them `int32` by default
const (
	UnstakingStatus = 1
	StakedStatus    = 2
)

// TODO(andrew): consolidate with genesis and `types/block.go`
var (
	ActorTypes = []ActorType{
		ActorType_App,
		ActorType_Node,
		ActorType_Fish,
		ActorType_Val,
	}
)

func (actorType ActorType) GetActorPoolName() string {
	switch actorType {
	case ActorType_App:
		return genesis.Pool_Names_AppStakePool.String()
	case ActorType_Val:
		return genesis.Pool_Names_ValidatorStakePool.String()
	case ActorType_Fish:
		return genesis.Pool_Names_FishermanStakePool.String()
	case ActorType_Node:
		return genesis.Pool_Names_ServiceNodeStakePool.String()
	default:
		log.Fatalf("unknown actor type: %v", actorType)
	}
	return ""
}

func (at ActorType) GetActorName() string {
	return ActorType_name[int32(at)]
}
