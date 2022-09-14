package types

import (
	"log"
)

// TODO(andrew): Moving this into a proto file enum (impacts everything) and making them `int32` by default
const (
	UnstakingStatus = 1
	StakedStatus    = 2
)

// TODO(andrew): consolidate with genesis and `types/block.go`
var (
	ActorTypes = []UtilActorType{ // TODO (andrew) consolidate with genesis
		UtilActorType_App,
		UtilActorType_Node,
		UtilActorType_Fish,
		UtilActorType_Val,
	}
)

func (x UtilActorType) GetActorPoolName() string {
	switch x {
	case UtilActorType_App:
		return "AppStakePool"
	case UtilActorType_Val:
		return "ValidatorStakePool"
	case UtilActorType_Fish:
		return "FishermanStakePool"
	case UtilActorType_Node:
		return "ServiceNodeStakePool"
	default:
		log.Fatalf("unknown actor type: %v", x)
	}
	return ""
}

func (x UtilActorType) GetActorName() string {
	return UtilActorType_name[int32(x)]
}
