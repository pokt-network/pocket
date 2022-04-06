package types

import (
	"sort"

	typesGenesis "github.com/pokt-network/pocket/shared/types/genesis"
)

type NodeId uint64

type ValAddrToIdMap map[string]NodeId // Mapping from hex encoded address to an integer node id.
type IdToValAddrMap map[NodeId]string // Mapping from node id to a hex encoded string address.

type ConsensusNodeState struct {
	NodeId NodeId
	Height uint64
	Round  uint8
	Step   uint8

	LeaderId NodeId
	IsLeader bool
}

func GetValAddrToIdMap(validatorMap map[string]*typesGenesis.Validator) (ValAddrToIdMap, IdToValAddrMap) {
	valAddresses := make([]string, 0, len(validatorMap))
	for addr := range validatorMap {
		valAddresses = append(valAddresses, addr)
	}
	sort.Strings(valAddresses)

	valToIdMap := make(ValAddrToIdMap, len(valAddresses))
	idToValMap := make(IdToValAddrMap, len(valAddresses))
	for i, addr := range valAddresses {
		nodeId := NodeId(i + 1)
		valToIdMap[addr] = nodeId
		idToValMap[nodeId] = addr
	}

	return valToIdMap, idToValMap
}
