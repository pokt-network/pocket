package types

// TODO: Split this file into multiple types files.
import (
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
)

type NodeId uint64

type (
	ValAddrToIdMap map[string]NodeId // Mapping from hex encoded address to an integer node id.
	IdToValAddrMap map[NodeId]string // Mapping from node id to a hex encoded string address.
	ValidatorMap   map[string]*coreTypes.Actor
)

type ConsensusNodeState struct {
	NodeId NodeId
	Height uint64
	Round  uint8
	Step   uint8

	LeaderId NodeId
	IsLeader bool
}

func ActorListToValidatorMap(actors []*coreTypes.Actor) (m ValidatorMap) {
	m = make(ValidatorMap, len(actors))
	for _, a := range actors {
		m[a.GetAddress()] = a
	}
	return
}
