package types

import (
	"github.com/pokt-network/pocket/shared/modules"
	"sort"
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

func GetValAddrToIdMap(validatorMap ValidatorMap) (ValAddrToIdMap, IdToValAddrMap) {
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

func (x *ConsensusConfig) GetPaceMakerConfig() modules.PacemakerConfig {
	return x.GetPacemakerConfig()
}

func (x *PacemakerConfig) SetTimeoutMsec(u uint64) {
	x.TimeoutMsec = u
}

type ValidatorMap map[string]modules.Actor

func ValidatorMapToModulesValidatorMap(validatorMap ValidatorMap) (vm modules.ValidatorMap) {
	vm = make(modules.ValidatorMap)
	for _, v := range validatorMap {
		vm[v.GetAddress()] = v
	}
	return
}

func ValidatorListToMap(validators []*Validator) (m ValidatorMap) {
	m = make(ValidatorMap, len(validators))
	for _, v := range validators {
		m[v.GetAddress()] = v
	}
	return
}

var _ modules.Actor = &Validator{}

func (x *Validator) GetPausedHeight() int64         { panic("not implemented on consensus validator") }
func (x *Validator) GetUnstakingHeight() int64      { panic("not implemented on consensus validator") }
func (x *Validator) GetOutput() string              { panic("not implemented on consensus validator") }
func (x *Validator) GetActorTyp() modules.ActorType { panic("not implemented on consensus validator") }
func (x *Validator) GetChains() []string            { panic("not implemented on consensus validator") }
