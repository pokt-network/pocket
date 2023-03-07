package persistence

import (
	"encoding/hex"

	"github.com/pokt-network/pocket/persistence/types"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	moduleTypes "github.com/pokt-network/pocket/shared/modules/types"
)

func (p *PostgresContext) GetValidatorExists(address []byte, height int64) (exists bool, err error) {
	return p.GetExists(types.ValidatorActor, address, height)
}

//nolint:gocritic // tooManyResultsChecker This function needs to return many values
func (p *PostgresContext) GetValidator(address []byte, height int64) (operator, publicKey, stakedTokens, serviceURL, outputAddress string, pausedHeight, unstakingHeight int64, err error) {
	actor, err := p.getActor(types.ValidatorActor, address, height)
	operator = actor.Address
	publicKey = actor.PublicKey
	stakedTokens = actor.StakedAmount
	serviceURL = actor.ServiceUrl
	outputAddress = actor.Output
	pausedHeight = actor.PausedHeight
	unstakingHeight = actor.UnstakingHeight
	return
}

func (p *PostgresContext) InsertValidator(address, publicKey, output []byte, _ bool, _ int32, serviceURL, stakedTokens string, pausedHeight, unstakingHeight int64) error {
	return p.InsertActor(types.ValidatorActor, &coreTypes.Actor{
		ActorType:       coreTypes.ActorType_ACTOR_TYPE_VAL,
		Address:         hex.EncodeToString(address),
		PublicKey:       hex.EncodeToString(publicKey),
		StakedAmount:    stakedTokens,
		ServiceUrl:      serviceURL,
		Output:          hex.EncodeToString(output),
		PausedHeight:    pausedHeight,
		UnstakingHeight: unstakingHeight,
	})
}

func (p *PostgresContext) UpdateValidator(address []byte, serviceURL, stakedAmount string) error {
	return p.UpdateActor(types.ValidatorActor, &coreTypes.Actor{
		ActorType:    coreTypes.ActorType_ACTOR_TYPE_VAL,
		Address:      hex.EncodeToString(address),
		StakedAmount: stakedAmount,
		ServiceUrl:   serviceURL,
	})
}

func (p *PostgresContext) GetValidatorStakeAmount(height int64, address []byte) (string, error) {
	return p.getActorStakeAmount(types.ValidatorActor, address, height)
}

func (p *PostgresContext) SetValidatorStakeAmount(address []byte, stakeAmount string) error {
	return p.setActorStakeAmount(types.ValidatorActor, address, stakeAmount)
}

func (p *PostgresContext) GetValidatorsReadyToUnstake(height int64, status int32) ([]*moduleTypes.UnstakingActor, error) {
	return p.GetActorsReadyToUnstake(types.ValidatorActor, height)
}

func (p *PostgresContext) GetValidatorStatus(address []byte, height int64) (int32, error) {
	return p.GetActorStatus(types.ValidatorActor, address, height)
}

func (p *PostgresContext) SetValidatorUnstakingHeightAndStatus(address []byte, unstakingHeight int64, status int32) error {
	return p.SetActorUnstakingHeightAndStatus(types.ValidatorActor, address, unstakingHeight)
}

func (p *PostgresContext) GetValidatorPauseHeightIfExists(address []byte, height int64) (int64, error) {
	return p.GetActorPauseHeightIfExists(types.ValidatorActor, address, height)
}

func (p *PostgresContext) SetValidatorsStatusAndUnstakingHeightIfPausedBefore(pausedBeforeHeight, unstakingHeight int64, status int32) error {
	return p.SetActorStatusAndUnstakingHeightIfPausedBefore(types.ValidatorActor, pausedBeforeHeight, unstakingHeight)
}

func (p *PostgresContext) SetValidatorPauseHeight(address []byte, height int64) error {
	return p.SetActorPauseHeight(types.ValidatorActor, address, height)
}

func (p *PostgresContext) GetValidatorOutputAddress(operator []byte, height int64) (output []byte, err error) {
	return p.GetActorOutputAddress(types.ValidatorActor, operator, height)
}

// TODO: implement missed blocks
func (p *PostgresContext) SetValidatorMissedBlocks(address []byte, missedBlocks int) error {
	return nil
}

// TODO: implement missed blocks
func (p *PostgresContext) GetValidatorMissedBlocks(address []byte, height int64) (int, error) {
	return 0, nil
}
