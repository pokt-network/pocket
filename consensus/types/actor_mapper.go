package types

import (
	"sort"

	coreTypes "github.com/pokt-network/pocket/shared/core/types"
)

type actorMapper struct {
	valAddrToIdMap ValAddrToIdMap
	idToValAddrMap IdToValAddrMap
	validatorMap   ValidatorMap
}

func NewActorMapper(validators []*coreTypes.Actor) *actorMapper {
	am := &actorMapper{
		valAddrToIdMap: make(ValAddrToIdMap, len(validators)),
		idToValAddrMap: make(IdToValAddrMap, len(validators)),
		validatorMap:   make(ValidatorMap, len(validators)),
	}

	valAddresses := make([]string, 0, len(validators))
	for _, val := range validators {
		addr := val.GetAddress()
		valAddresses = append(valAddresses, addr)
		am.validatorMap[addr] = val
	}
	sort.Strings(valAddresses)

	for i, addr := range valAddresses {
		nodeId := NodeId(i + 1) // TODO(#434): Improve the use of NodeIDs
		am.valAddrToIdMap[addr] = nodeId
		am.idToValAddrMap[nodeId] = addr
	}

	return am
}

func (am *actorMapper) GetValidatorMap() ValidatorMap {
	return am.validatorMap
}

func (am *actorMapper) GetValAddrToIdMap() ValAddrToIdMap {
	return am.valAddrToIdMap
}

func (am *actorMapper) GetIdToValAddrMap() IdToValAddrMap {
	return am.idToValAddrMap
}
