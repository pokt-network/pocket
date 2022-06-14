package persistence

import (
	"encoding/hex"

	"github.com/pokt-network/pocket/persistence/schema"
	"github.com/pokt-network/pocket/shared/types"
)

func (p PostgresContext) GetValidatorExists(address []byte, height int64) (exists bool, err error) {
	return p.GetExists(address, height, schema.ValidatorExistsQuery)
}

func (p PostgresContext) GetValidator(address []byte, height int64) (operator, publicKey, stakedTokens, serviceURL, outputAddress string, pausedHeight, unstakingHeight int64, err error) {
	actor, err := p.GetActor(address, height, schema.ValidatorQuery, nil)
	operator = actor.Address
	publicKey = actor.PublicKey
	stakedTokens = actor.StakedTokens
	serviceURL = actor.GenericParam
	outputAddress = actor.OutputAddress
	pausedHeight = actor.PausedHeight
	unstakingHeight = actor.UnstakingHeight
	return
}

// TODO(Andrew): remove paused and status from the interface
func (p PostgresContext) InsertValidator(address []byte, publicKey []byte, output []byte, paused bool, status int, serviceURL string, stakedTokens string, pausedHeight int64, unstakingHeight int64) error {
	return p.InsertActor(schema.GenericActor{
		Address:         hex.EncodeToString(address),
		PublicKey:       hex.EncodeToString(publicKey),
		StakedTokens:    stakedTokens,
		GenericParam:    serviceURL,
		OutputAddress:   hex.EncodeToString(output),
		PausedHeight:    pausedHeight,
		UnstakingHeight: unstakingHeight,
	}, schema.InsertValidatorQuery)
}

// TODO(Andrew): change amount to add, to the amount to be SET
func (p PostgresContext) UpdateValidator(address []byte, serviceURL string, stakedTokens string) error {
	return p.UpdateActor(schema.GenericActor{
		Address:      hex.EncodeToString(address),
		StakedTokens: stakedTokens,
		GenericParam: serviceURL,
	}, schema.UpdateValidatorQuery, nil, "")
}

// NOTE: Leaving as transaction as I anticipate we'll need more ops in the future
func (p PostgresContext) DeleteValidator(address []byte) error {
	return nil // no op
}

// TODO(Andrew): remove status - not needed
func (p PostgresContext) GetValidatorsReadyToUnstake(height int64, status int) (Validators []*types.UnstakingActor, err error) {
	return p.ActorReadyToUnstakeWithChains(height, schema.ValidatorReadyToUnstakeQuery)
}

func (p PostgresContext) GetValidatorStatus(address []byte, height int64) (status int, err error) {
	return p.GetActorStatus(address, height, schema.ValidatorUnstakingHeightQuery)
}

// TODO(Andrew): remove status - no longer needed
func (p PostgresContext) SetValidatorUnstakingHeightAndStatus(address []byte, unstakingHeight int64, status int) error {
	return p.SetActorUnstakingHeightAndStatus(address, unstakingHeight, schema.UpdateValidatorUnstakingHeightQuery)
}

func (p PostgresContext) GetValidatorPauseHeightIfExists(address []byte, height int64) (int64, error) {
	return p.GetActorPauseHeightIfExists(address, height, schema.ValidatorPauseHeightQuery)
}

// TODO(Andrew): remove status - it's not needed
func (p PostgresContext) SetValidatorsStatusAndUnstakingHeightPausedBefore(pausedBeforeHeight, unstakingHeight int64, status int) error {
	return p.SetActorStatusAndUnstakingHeightPausedBefore(pausedBeforeHeight, unstakingHeight, schema.UpdateValidatorsPausedBefore)
}

func (p PostgresContext) SetValidatorPauseHeight(address []byte, height int64) error {
	return p.SetActorPauseHeight(address, height, schema.UpdateValidatorPausedHeightQuery)
}

func (p PostgresContext) SetValidatorStakedTokens(address []byte, tokens string) error { // TODO deprecate and use update validator
	height, err := p.GetHeight()
	if err != nil {
		return err
	}
	// TODO make atomic
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

func (p PostgresContext) GetValidatorStakedTokens(address []byte, height int64) (tokens string, err error) { // TODO deprecate and use update validator
	_, _, tokens, _, _, _, _, err = p.GetValidator(address, height)
	return
}

func (p PostgresContext) GetValidatorOutputAddress(operator []byte, height int64) (output []byte, err error) {
	return p.GetActorOutputAddress(operator, height, schema.ValidatorOutputAddressQuery)
}

func (p PostgresContext) SetValidatorPauseHeightAndMissedBlocks(address []byte, pausedHeight int64, missedBlocks int) error {
	// TODO implement missed blocks
	return nil
}

func (p PostgresContext) SetValidatorMissedBlocks(address []byte, missedBlocks int) error {
	// TODO implement missed blocks
	return nil
}

func (p PostgresContext) GetValidatorMissedBlocks(address []byte) (int, error) {
	// TODO implement missed blocks
	return 0, nil
}
