package types

import (
	"sort"

	"github.com/pokt-network/pocket/shared/types"
)

type NodeId uint64

type ValToIdMap map[string]NodeId // Mapping from hex encoded address to an integer node id.
type IdToValMap map[NodeId]string // Mapping from node id to a hex encoded string address.

type ConsensusNodeState struct {
	NodeId   NodeId
	Height   uint64
	Round    uint8
	Step     uint8
	IsLeader bool
	LeaderId NodeId
}

func GetValToIdMap(valMap types.ValMap) (ValToIdMap, IdToValMap) {
	valAddresses := make([]string, 0, len(valMap))
	for addr := range valMap {
		valAddresses = append(valAddresses, addr)
	}
	sort.Strings(valAddresses)

	valToIdMap := make(ValToIdMap, len(valMap))
	idToValMap := make(IdToValMap, len(valMap))
	for i, addr := range valAddresses {
		valToIdMap[addr] = NodeId(i + 1)
		idToValMap[NodeId(i+1)] = addr
	}

	return valToIdMap, idToValMap
}
