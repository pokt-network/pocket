package types

import (
	"sort"

	"github.com/pokt-network/pocket/shared/types"
)

type NodeId uint64

type ValAddrToIdMap map[string]NodeId // Mapping from hex encoded address to an integer node id.
type IdToValAddrMap map[NodeId]string // Mapping from node id to a hex encoded string address.

type ConsensusNodeState struct {
	NodeId   NodeId
	Height   uint64
	Round    uint8
	Step     uint8
	IsLeader bool
	LeaderId NodeId
}

func GetValAddrToIdMap(valMap types.ValMap) (ValAddrToIdMap, IdToValAddrMap) {
	valAddresses := make([]string, 0, len(valMap))
	for addr := range valMap {
		valAddresses = append(valAddresses, addr)
	}
	sort.Strings(valAddresses)

	valToIdMap := make(ValAddrToIdMap, len(valMap))
	idToValMap := make(IdToValAddrMap, len(valMap))
	for i, addr := range valAddresses {
		nodeId := NodeId(i + 1)
		valToIdMap[addr] = nodeId
		idToValMap[nodeId] = addr
	}

	return valToIdMap, idToValMap
}
