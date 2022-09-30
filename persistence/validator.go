package persistence

import (
	"encoding/hex"

	"github.com/pokt-network/pocket/persistence/types"
	"github.com/pokt-network/pocket/shared/modules"
)

func (p PostgresContext) GetValidatorExists(address []byte, height int64) (exists bool, err error) {
	return p.GetExists(types.ValidatorActor, address, height)
}

func (p PostgresContext) GetValidator(address []byte, height int64) (operator, publicKey, stakedTokens, serviceURL, outputAddress string, pausedHeight, unstakingHeight int64, err error) {
	actor, err := p.GetActor(types.ValidatorActor, address, height)
	operator = actor.Address
	publicKey = actor.PublicKey
	stakedTokens = actor.StakedTokens
	serviceURL = actor.ActorSpecificParam
	outputAddress = actor.OutputAddress
	pausedHeight = actor.PausedHeight
	unstakingHeight = actor.UnstakingHeight
	return
}

func (p PostgresContext) InsertValidator(address []byte, publicKey []byte, output []byte, _ bool, _ int32, serviceURL string, stakedTokens string, pausedHeight int64, unstakingHeight int64) error {
	return p.InsertActor(types.ValidatorActor, types.BaseActor{
		Address:            hex.EncodeToString(address),
		PublicKey:          hex.EncodeToString(publicKey),
		StakedTokens:       stakedTokens,
		ActorSpecificParam: serviceURL,
		OutputAddress:      hex.EncodeToString(output),
		PausedHeight:       pausedHeight,
		UnstakingHeight:    unstakingHeight,
	})
}

func (p PostgresContext) UpdateValidator(address []byte, serviceURL string, stakedAmount string) error {
	return p.UpdateActor(types.ValidatorActor, types.BaseActor{
		Address:            hex.EncodeToString(address),
		StakedTokens:       stakedAmount,
		ActorSpecificParam: serviceURL,
	})
}

func (p PostgresContext) GetValidatorStakeAmount(height int64, address []byte) (string, error) {
	return p.GetActorStakeAmount(types.ValidatorActor, address, height)
}

func (p PostgresContext) SetValidatorStakeAmount(address []byte, stakeAmount string) error {
	return p.SetActorStakeAmount(types.ValidatorActor, address, stakeAmount)
}

func (p PostgresContext) GetValidatorsReadyToUnstake(height int64, status int32) ([]modules.IUnstakingActor, error) {
	return p.GetActorsReadyToUnstake(types.ValidatorActor, height)
}

func (p PostgresContext) GetValidatorStatus(address []byte, height int64) (int32, error) {
	return p.GetActorStatus(types.ValidatorActor, address, height)
}

func (p PostgresContext) SetValidatorUnstakingHeightAndStatus(address []byte, unstakingHeight int64, status int32) error {
	return p.SetActorUnstakingHeightAndStatus(types.ValidatorActor, address, unstakingHeight)
}

func (p PostgresContext) GetValidatorPauseHeightIfExists(address []byte, height int64) (int64, error) {
	return p.GetActorPauseHeightIfExists(types.ValidatorActor, address, height)
}

func (p PostgresContext) SetValidatorsStatusAndUnstakingHeightIfPausedBefore(pausedBeforeHeight, unstakingHeight int64, status int32) error {
	return p.SetActorStatusAndUnstakingHeightIfPausedBefore(types.ValidatorActor, pausedBeforeHeight, unstakingHeight)
}

func (p PostgresContext) SetValidatorPauseHeight(address []byte, height int64) error {
	return p.SetActorPauseHeight(types.ValidatorActor, address, height)
}

// TODO(team): The Get & Update operations need to be made atomic
// TODO(team): Deprecate this functiona altogether and use UpdateValidator where applicable
func (p PostgresContext) SetValidatorStakedTokens(address []byte, tokens string) error { //
	height, err := p.GetHeight()
	if err != nil {
		return err
	}
	operator, _, _, serviceURL, _, _, _, err := p.GetValidator(address, height)
	if err != nil {
		return err
	}
	addr, err := hex.DecodeString(operator)
	if err != nil {
		return err
	}
	return p.UpdateValidator(addr, serviceURL, tokens)
}

func (p PostgresContext) GetValidatorStakedTokens(address []byte, height int64) (tokens string, err error) {
	_, _, tokens, _, _, _, _, err = p.GetValidator(address, height)
	return
}

func (p PostgresContext) GetValidatorOutputAddress(operator []byte, height int64) (output []byte, err error) {
	return p.GetActorOutputAddress(types.ValidatorActor, operator, height)
}

// TODO(team): implement missed blocks
func (p PostgresContext) SetValidatorPauseHeightAndMissedBlocks(address []byte, pausedHeight int64, missedBlocks int) error {
	return nil
}

// TODO(team): implement missed blocks
func (p PostgresContext) SetValidatorMissedBlocks(address []byte, missedBlocks int) error {
	return nil
}

// TODO(team): implement missed blocks
func (p PostgresContext) GetValidatorMissedBlocks(address []byte, height int64) (int, error) {
	return 0, nil
}
