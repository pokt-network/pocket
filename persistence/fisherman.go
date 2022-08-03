package persistence

import (
	"encoding/hex"
	"log"

	"github.com/pokt-network/pocket/persistence/schema"
	"github.com/pokt-network/pocket/shared/types"
)

func (p PostgresContext) GetFishermanExists(address []byte, height int64) (exists bool, err error) {
	return p.GetExists(schema.FishermanActor, address, height)
}

func (p PostgresContext) GetFisherman(address []byte, height int64) (operator, publicKey, stakedTokens, serviceURL, outputAddress string, pausedHeight, unstakingHeight int64, chains []string, err error) {
	actor, err := p.GetActor(schema.FishermanActor, address, height)
	operator = actor.Address
	publicKey = actor.PublicKey
	stakedTokens = actor.StakedTokens
	serviceURL = actor.ActorSpecificParam
	outputAddress = actor.OutputAddress
	pausedHeight = actor.PausedHeight
	unstakingHeight = actor.UnstakingHeight
	chains = actor.Chains
	return
}

func (p PostgresContext) InsertFisherman(address []byte, publicKey []byte, output []byte, _ bool, _ int, serviceURL string, stakedAmount string, chains []string, pausedHeight int64, unstakingHeight int64) error {
	return p.InsertActor(schema.FishermanActor, schema.BaseActor{
		Address:            hex.EncodeToString(address),
		PublicKey:          hex.EncodeToString(publicKey),
		StakedTokens:       stakedAmount,
		ActorSpecificParam: serviceURL,
		OutputAddress:      hex.EncodeToString(output),
		PausedHeight:       pausedHeight,
		UnstakingHeight:    unstakingHeight,
		Chains:             chains,
	})
}

func (p PostgresContext) UpdateFisherman(address []byte, serviceURL string, stakedAmount string, chains []string) error {
	return p.UpdateActor(schema.FishermanActor, schema.BaseActor{
		Address:            hex.EncodeToString(address),
		StakedTokens:       stakedAmount,
		ActorSpecificParam: serviceURL,
		Chains:             chains,
	})
}

func (p PostgresContext) DeleteFisherman(_ []byte) error {
	log.Println("[DEBUG] DeleteFisherman is a NOOP")
	return nil
}

func (p PostgresContext) GetFishermenReadyToUnstake(height int64, _ int) ([]*types.UnstakingActor, error) {
	return p.GetActorsReadyToUnstake(schema.FishermanActor, height)
}

func (p PostgresContext) GetFishermanStatus(address []byte, height int64) (status int, err error) {
	return p.GetActorStatus(schema.FishermanActor, address, height)
}

func (p PostgresContext) SetFishermanUnstakingHeightAndStatus(address []byte, unstakingHeight int64, _ int) error {
	return p.SetActorUnstakingHeightAndStatus(schema.FishermanActor, address, unstakingHeight)
}

func (p PostgresContext) GetFishermanPauseHeightIfExists(address []byte, height int64) (int64, error) {
	return p.GetActorPauseHeightIfExists(schema.FishermanActor, address, height)
}

func (p PostgresContext) SetFishermanStatusAndUnstakingHeightIfPausedBefore(pausedBeforeHeight, unstakingHeight int64, _ int) error {
	return p.SetActorStatusAndUnstakingHeightIfPausedBefore(schema.FishermanActor, pausedBeforeHeight, unstakingHeight)
}

func (p PostgresContext) SetFishermanPauseHeight(address []byte, height int64) error {
	return p.SetActorPauseHeight(schema.FishermanActor, address, height)
}

func (p PostgresContext) GetFishermanOutputAddress(operator []byte, height int64) (output []byte, err error) {
	return p.GetActorOutputAddress(schema.FishermanActor, operator, height)
}
