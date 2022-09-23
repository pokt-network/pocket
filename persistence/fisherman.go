package persistence

import (
	"encoding/hex"
	"log"

	"github.com/pokt-network/pocket/persistence/types"
	"github.com/pokt-network/pocket/shared/modules"
)

func (p PostgresContext) GetFishermanExists(address []byte, height int64) (exists bool, err error) {
	return p.GetExists(types.FishermanActor, address, height)
}

func (p PostgresContext) GetFisherman(address []byte, height int64) (operator, publicKey, stakedTokens, serviceURL, outputAddress string, pausedHeight, unstakingHeight int64, chains []string, err error) {
	actor, err := p.GetActor(types.FishermanActor, address, height)
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

func (p PostgresContext) InsertFisherman(address []byte, publicKey []byte, output []byte, paused bool, status int32, serviceURL string, stakedTokens string, chains []string, pausedHeight int64, unstakingHeight int64) error {
	return p.InsertActor(types.FishermanActor, types.BaseActor{
		Address:            hex.EncodeToString(address),
		PublicKey:          hex.EncodeToString(publicKey),
		StakedTokens:       stakedTokens,
		ActorSpecificParam: serviceURL,
		OutputAddress:      hex.EncodeToString(output),
		PausedHeight:       pausedHeight,
		UnstakingHeight:    unstakingHeight,
		Chains:             chains,
	})
}

func (p PostgresContext) UpdateFisherman(address []byte, serviceURL string, stakedAmount string, chains []string) error {
	return p.UpdateActor(types.FishermanActor, types.BaseActor{
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

func (p PostgresContext) GetFishermanStakeAmount(height int64, address []byte) (string, error) {
	return p.GetActorStakeAmount(types.FishermanActor, address, height)
}

func (p PostgresContext) SetFishermanStakeAmount(address []byte, stakeAmount string) error {
	return p.SetActorStakeAmount(types.FishermanActor, address, stakeAmount)
}

func (p PostgresContext) GetFishermenReadyToUnstake(height int64, status int32) ([]modules.IUnstakingActor, error) {
	return p.GetActorsReadyToUnstake(types.FishermanActor, height)
}

func (p PostgresContext) GetFishermanStatus(address []byte, height int64) (status int32, err error) {
	return p.GetActorStatus(types.FishermanActor, address, height)
}

func (p PostgresContext) SetFishermanUnstakingHeightAndStatus(address []byte, unstakingHeight int64, status int32) error {
	return p.SetActorUnstakingHeightAndStatus(types.FishermanActor, address, unstakingHeight)
}

func (p PostgresContext) GetFishermanPauseHeightIfExists(address []byte, height int64) (int64, error) {
	return p.GetActorPauseHeightIfExists(types.FishermanActor, address, height)
}

func (p PostgresContext) SetFishermanStatusAndUnstakingHeightIfPausedBefore(pausedBeforeHeight, unstakingHeight int64, status int32) error {
	return p.SetActorStatusAndUnstakingHeightIfPausedBefore(types.FishermanActor, pausedBeforeHeight, unstakingHeight)
}

func (p PostgresContext) SetFishermanPauseHeight(address []byte, height int64) error {
	return p.SetActorPauseHeight(types.FishermanActor, address, height)
}

func (p PostgresContext) GetFishermanOutputAddress(operator []byte, height int64) (output []byte, err error) {
	return p.GetActorOutputAddress(types.FishermanActor, operator, height)
}
