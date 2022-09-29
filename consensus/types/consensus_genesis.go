package types

import "github.com/pokt-network/pocket/shared/modules"

var _ modules.ConsensusGenesisState = &ConsensusGenesisState{}

func (x *ConsensusGenesisState) GetVals() []modules.Actor {
	return ActorsToActorsInterface(x.GetValidators())
}

func ActorsToActorsInterface(a []*Validator) (actorI []modules.Actor) {
	for _, actor := range a {
		actorI = append(actorI, actor)
	}
	return
}

var _ modules.Actor = &Validator{}
