package types

import (
	"github.com/pokt-network/pocket/shared/modules"
)

func actorToValidator(actor modules.Actor) *Validator {
	return &Validator{
		Address:      actor.GetAddress(),
		PublicKey:    actor.GetPublicKey(),
		StakedAmount: actor.GetStakedAmount(),
		GenericParam: actor.GetGenericParam(),
	}
}

func ToConsensusValidators(actors []modules.Actor) (vals []*Validator) {
	vals = make([]*Validator, len(actors))
	for i, actor := range actors {
		vals[i] = actorToValidator(actor)
	}
	return
}
