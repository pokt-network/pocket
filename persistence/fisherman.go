package persistence

import (
	"encoding/hex"

	"github.com/pokt-network/pocket/persistence/types"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	moduleTypes "github.com/pokt-network/pocket/shared/modules/types"
)

func (p *PostgresContext) GetFishermanExists(address []byte, height int64) (exists bool, err error) {
	return p.GetExists(types.FishermanActor, address, height)
}

//nolint:gocritic // tooManyResultsChecker This function needs to return many values
func (p *PostgresContext) GetFisherman(address []byte, height int64) (operator, publicKey, stakedTokens, serviceURL, outputAddress string, pausedHeight, unstakingHeight int64, chains []string, err error) {
	actor, err := p.getActor(types.FishermanActor, address, height)
	operator = actor.Address
	publicKey = actor.PublicKey
	stakedTokens = actor.StakedAmount
	serviceURL = actor.ServiceUrl
	outputAddress = actor.Output
	pausedHeight = actor.PausedHeight
	unstakingHeight = actor.UnstakingHeight
	chains = actor.Chains
	return
}

func (p *PostgresContext) InsertFisherman(address, publicKey, output []byte, _ bool, _ int32, serviceURL, stakedTokens string, chains []string, pausedHeight, unstakingHeight int64) error {
	return p.InsertActor(types.FishermanActor, &coreTypes.Actor{
		ActorType:       coreTypes.ActorType_ACTOR_TYPE_FISH,
		Address:         hex.EncodeToString(address),
		PublicKey:       hex.EncodeToString(publicKey),
		Chains:          chains,
		ServiceUrl:      serviceURL,
		StakedAmount:    stakedTokens,
		PausedHeight:    pausedHeight,
		UnstakingHeight: unstakingHeight,
		Output:          hex.EncodeToString(output),
	})
}

func (p *PostgresContext) UpdateFisherman(address []byte, serviceURL, stakedAmount string, chains []string) error {
	return p.UpdateActor(types.FishermanActor, &coreTypes.Actor{
		ActorType:    coreTypes.ActorType_ACTOR_TYPE_FISH,
		Address:      hex.EncodeToString(address),
		StakedAmount: stakedAmount,
		ServiceUrl:   serviceURL,
		Chains:       chains,
	})
}

func (p *PostgresContext) GetFishermanStakeAmount(height int64, address []byte) (string, error) {
	return p.getActorStakeAmount(types.FishermanActor, address, height)
}

func (p *PostgresContext) SetFishermanStakeAmount(address []byte, stakeAmount string) error {
	return p.setActorStakeAmount(types.FishermanActor, address, stakeAmount)
}

func (p *PostgresContext) GetFishermenReadyToUnstake(height int64, status int32) ([]*moduleTypes.UnstakingActor, error) {
	return p.GetActorsReadyToUnstake(types.FishermanActor, height)
}

func (p *PostgresContext) GetFishermanStatus(address []byte, height int64) (status int32, err error) {
	return p.GetActorStatus(types.FishermanActor, address, height)
}

func (p *PostgresContext) SetFishermanUnstakingHeightAndStatus(address []byte, unstakingHeight int64, status int32) error {
	return p.SetActorUnstakingHeightAndStatus(types.FishermanActor, address, unstakingHeight)
}

func (p *PostgresContext) GetFishermanPauseHeightIfExists(address []byte, height int64) (int64, error) {
	return p.GetActorPauseHeightIfExists(types.FishermanActor, address, height)
}

func (p *PostgresContext) SetFishermanStatusAndUnstakingHeightIfPausedBefore(pausedBeforeHeight, unstakingHeight int64, status int32) error {
	return p.SetActorStatusAndUnstakingHeightIfPausedBefore(types.FishermanActor, pausedBeforeHeight, unstakingHeight)
}

func (p *PostgresContext) SetFishermanPauseHeight(address []byte, height int64) error {
	return p.SetActorPauseHeight(types.FishermanActor, address, height)
}

func (p *PostgresContext) GetFishermanOutputAddress(operator []byte, height int64) (output []byte, err error) {
	return p.GetActorOutputAddress(types.FishermanActor, operator, height)
}
