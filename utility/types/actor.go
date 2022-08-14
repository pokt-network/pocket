package types

import (
	"log"

	"github.com/pokt-network/pocket/shared/types/genesis"
)

// REFACTOR: Moving this into a proto file enum (impacts everything) and making them `int32` by default
const (
	UnstakingStatus = 1
	StakedStatus    = 2
)

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
		return genesis.AppStakePoolName
	case ActorType_Val:
		return genesis.ValidatorStakePoolName
	case ActorType_Fish:
		return genesis.FishermanStakePoolName
	case ActorType_Node:
		return genesis.ServiceNodeStakePoolName
	default:
		log.Fatalf("unknown actor type: %v", actorType)
	}
	return ""
}

func (at ActorType) GetActorName() string {
	return ActorType_name[int32(at)]
}
