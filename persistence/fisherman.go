package persistence

import (
	"encoding/hex"

	"github.com/pokt-network/pocket/persistence/schema"
	"github.com/pokt-network/pocket/shared/types"
)

func (p PostgresContext) GetFishermanExists(address []byte, height int64) (exists bool, err error) {
	return p.GetExists(address, height, schema.FishermanActor.GetExistsQuery)
}

func (p PostgresContext) GetFisherman(address []byte, height int64) (operator, publicKey, stakedTokens, serviceURL, outputAddress string, pausedHeight, unstakingHeight int64, chains []string, err error) {
	actor, err := p.GetActor(address, height, schema.FishermanActor.GetQuery, schema.FishermanActor.GetChainsQuery)
	operator = actor.Address
	publicKey = actor.PublicKey
	stakedTokens = actor.StakedTokens
	serviceURL = actor.GenericParam
	outputAddress = actor.OutputAddress
	pausedHeight = actor.PausedHeight
	unstakingHeight = actor.UnstakingHeight
	chains = actor.Chains
	return
}

// TODO(Andrew): remove paused and status from the interface
func (p PostgresContext) InsertFisherman(address []byte, publicKey []byte, output []byte, paused bool, status int, serviceURL string, stakedTokens string, chains []string, pausedHeight int64, unstakingHeight int64) error {
	return p.InsertActor(schema.GenericActor{
		Address:         hex.EncodeToString(address),
		PublicKey:       hex.EncodeToString(publicKey),
		StakedTokens:    stakedTokens,
		GenericParam:    serviceURL,
		OutputAddress:   hex.EncodeToString(output),
		PausedHeight:    pausedHeight,
		UnstakingHeight: unstakingHeight,
		Chains:          chains,
	}, schema.FishermanActor.InsertQuery)
}

// TODO(Andrew): change amount to add, to the amount to be SET
func (p PostgresContext) UpdateFisherman(address []byte, serviceURL string, stakedTokens string, chains []string) error {
	return p.UpdateActor(schema.GenericActor{
		Address:      hex.EncodeToString(address),
		StakedTokens: stakedTokens,
		GenericParam: serviceURL,
		Chains:       chains,
	}, schema.FishermanActor.UpdateQuery, schema.FishermanActor.UpdateChainsQuery, schema.FishermanActor.GetChainsTableName())
}

func (p PostgresContext) DeleteFisherman(_ []byte) error {
	return nil // no op
}

// TODO(Andrew): remove status - not needed
func (p PostgresContext) GetFishermanReadyToUnstake(height int64, _ int) (Fishermans []*types.UnstakingActor, err error) {
	return p.ActorReadyToUnstakeWithChains(height, schema.FishermanActor.GetReadyToUnstakeQuery)
}

func (p PostgresContext) GetFishermanStatus(address []byte, height int64) (status int, err error) {
	return p.GetActorStatus(address, height, schema.FishermanActor.GetUnstakingHeightQuery)
}

// TODO(Andrew): remove status - no longer needed
func (p PostgresContext) SetFishermanUnstakingHeightAndStatus(address []byte, unstakingHeight int64, _ int) error {
	return p.SetActorUnstakingHeightAndStatus(address, unstakingHeight, schema.FishermanActor.UpdateUnstakingHeightQuery)
}

func (p PostgresContext) GetFishermanPauseHeightIfExists(address []byte, height int64) (int64, error) {
	return p.GetActorPauseHeightIfExists(address, height, schema.FishermanActor.GetPausedHeightQuery)
}

// TODO(Andrew): remove status - it's not needed
func (p PostgresContext) SetFishermansStatusAndUnstakingHeightPausedBefore(pausedBeforeHeight, unstakingHeight int64, _ int) error {
	return p.SetActorStatusAndUnstakingHeightPausedBefore(pausedBeforeHeight, unstakingHeight, schema.FishermanActor.UpdatePausedBefore)
}

func (p PostgresContext) SetFishermanPauseHeight(address []byte, height int64) error {
	return p.SetActorPauseHeight(address, height, schema.FishermanActor.UpdatePausedHeightQuery)
}

func (p PostgresContext) GetFishermanOutputAddress(operator []byte, height int64) (output []byte, err error) {
	return p.GetActorOutputAddress(operator, height, schema.FishermanActor.GetOutputAddressQuery)
}
