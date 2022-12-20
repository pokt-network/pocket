package types

import (
	"encoding/hex"
	"log"

	shared "github.com/pokt-network/pocket/internal/shared/modules"
)

var _ shared.IUnstakingActor = &UnstakingActor{}

func (x *UnstakingActor) SetAddress(address string) { // TODO (team) convert address to string #149
	s, err := hex.DecodeString(address)
	if err != nil {
		log.Fatal(err)
	}
	x.Address = s
}

func (x *UnstakingActor) SetStakeAmount(stakeAmount string) {
	x.StakeAmount = stakeAmount
}

func (x *UnstakingActor) SetOutputAddress(address string) {
	s, err := hex.DecodeString(address)
	if err != nil {
		log.Fatal(err)
	}
	x.OutputAddress = s
}
