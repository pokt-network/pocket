package types

import (
	"sort"

	"github.com/pokt-network/pocket/shared/types"
)

type NodeId uint64

type ValToIdMap map[string]NodeId

type ConsensusNodeState struct {
	NodeId   NodeId
	Height   uint64
	Round    uint8
	Step     uint8
	IsLeader bool
	LeaderId NodeId
}

func GetValToIdMap(valMap types.ValMap) ValToIdMap {
	valAddresses := make([]string, 0, len(valMap))
	for addr := range valMap {
		valAddresses = append(valAddresses, addr)
	}
	sort.Strings(valAddresses)

	valToIdMap := make(ValToIdMap, len(valMap))
	for i, addr := range valAddresses {
		valToIdMap[addr] = NodeId(i + 1)
	}

	return valToIdMap
}
