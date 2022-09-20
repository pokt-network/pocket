package types

import "github.com/pokt-network/pocket/shared/modules"

func actorToValidator(actor modules.Actor) *Validator {
	return &Validator{
		Address:      actor.GetAddress(),
		PublicKey:    actor.GetPublicKey(),
		StakedAmount: actor.GetStakedAmount(),
		GenericParam: actor.GetGenericParam(),
	}
}

func ToConsensusValidators(actors []modules.Actor) []*Validator {
	r := make([]*Validator, 0)
	for _, a := range actors {
		r = append(r, actorToValidator(a))
	}
	return r
}
