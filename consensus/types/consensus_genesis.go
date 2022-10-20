package types

import "github.com/pokt-network/pocket/shared/modules"

var _ modules.ConsensusGenesisState = &ConsensusGenesisState{}

func (x *ConsensusGenesisState) GetVals() []modules.Actor {
	return ActorsToActorsInterface(x.GetValidators())
}

func ActorsToActorsInterface(vals []*Validator) (actorI []modules.Actor) {
	actorI = make([]modules.Actor, len(vals))
	for i, actor := range vals {
		actorI[i] = actor
	}
	return
}

var _ modules.Actor = &Validator{}
