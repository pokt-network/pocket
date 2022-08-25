package types

import (
	"encoding/hex"
	shared "github.com/pokt-network/pocket/shared/modules"
)

var _ shared.UnstakingActorI = &UnstakingActor{}

func (x *UnstakingActor) SetAddress(address string) { // TODO (team) convert address to string #149
	s, _ := hex.DecodeString(address)
	x.Address = s
}

func (x *UnstakingActor) SetStakeAmount(stakeAmount string) {
	x.StakeAmount = stakeAmount
}

func (x *UnstakingActor) SetOutputAddress(address string) {
	s, _ := hex.DecodeString(address)
	x.OutputAddress = s
}
