package persistence

import (
	"encoding/hex"

	"github.com/pokt-network/pocket/persistence/schema"
	"github.com/pokt-network/pocket/shared/types"
)

func (p PostgresContext) GetAppExists(address []byte, height int64) (exists bool, err error) {
	return p.GetExists(address, height, schema.AppExistsQuery)
}

func (p PostgresContext) GetApp(address []byte, height int64) (operator, publicKey, stakedTokens, maxRelays, outputAddress string, pauseHeight, unstakingHeight int64, chains []string, err error) {
	actor, err := p.GetActor(address, height, schema.AppQuery, schema.AppChainsQuery)
	operator = actor.Address
	publicKey = actor.PublicKey
	stakedTokens = actor.StakedTokens
	maxRelays = actor.GenericParam
	outputAddress = actor.OutputAddress
	pauseHeight = actor.PausedHeight
	unstakingHeight = actor.UnstakingHeight
	chains = actor.Chains
	return
}

// TODO(Andrew): remove paused and status from the interface
func (p PostgresContext) InsertApp(address []byte, publicKey []byte, output []byte, _ bool, _ int, maxRelays string, stakedTokens string, chains []string, pausedHeight int64, unstakingHeight int64) error {
	return p.InsertActor(schema.GenericActor{
		Address:         hex.EncodeToString(address),
		PublicKey:       hex.EncodeToString(publicKey),
		StakedTokens:    stakedTokens,
		GenericParam:    maxRelays,
		OutputAddress:   hex.EncodeToString(output),
		PausedHeight:    pausedHeight,
		UnstakingHeight: unstakingHeight,
		Chains:          chains,
	}, schema.InsertAppQuery)
}

// TODO(Andrew): change `amountToAdd` to`amountToSET`
// NOTE: originally, we thought we could do arithmetic operations quite easily to just 'bump' the max relays - but since
// it's a bigint (TEXT in Postgres) I don't believe this optimization is possible. Best use new amounts for 'Update'
func (p PostgresContext) UpdateApp(address []byte, maxRelays string, stakedTokens string, chains []string) error {
	return p.UpdateActor(schema.GenericActor{
		Address:      hex.EncodeToString(address),
		StakedTokens: stakedTokens,
		GenericParam: maxRelays,
		Chains:       chains,
	}, schema.UpdateAppQuery, schema.UpdateAppChainsQuery, schema.AppChainsTableName)
}

func (p PostgresContext) DeleteApp(_ []byte) error {
	// No op
	return nil
}

// TODO(Andrew): remove status (second parameter) - not needed
func (p PostgresContext) GetAppsReadyToUnstake(height int64, _ int) (apps []*types.UnstakingActor, err error) {
	return p.ActorReadyToUnstakeWithChains(height, schema.AppsReadyToUnstakeQuery)
}

func (p PostgresContext) GetAppStatus(address []byte, height int64) (status int, err error) {
	return p.GetActorStatus(address, height, schema.AppUnstakingHeightQuery)
}

// TODO(Andrew): remove status (third parameter) - no longer needed
func (p PostgresContext) SetAppUnstakingHeightAndStatus(address []byte, unstakingHeight int64, _ int) error {
	return p.SetActorUnstakingHeightAndStatus(address, unstakingHeight, schema.UpdateAppUnstakingHeightQuery)
}

// DISCUSS(drewsky): Need to create a semantic constant for an error return value, but should it be 0 or -1?
func (p PostgresContext) GetAppPauseHeightIfExists(address []byte, height int64) (pausedHeight int64, err error) {
	return p.GetActorPauseHeightIfExists(address, height, schema.AppPausedHeightQuery)
}

// TODO(Andrew): remove status (third parameter) - it's not needed
// DISCUSS(drewsky): This function seems to be doing too much from a naming perspective. Perhaps `SetPausedAppsToStartUnstaking`?
func (p PostgresContext) SetAppsStatusAndUnstakingHeightPausedBefore(pausedBeforeHeight, unstakingHeight int64, _ int) error {
	return p.SetActorStatusAndUnstakingHeightPausedBefore(pausedBeforeHeight, unstakingHeight, schema.UpdateAppsPausedBefore)
}

func (p PostgresContext) SetAppPauseHeight(address []byte, height int64) error {
	return p.SetActorPauseHeight(address, height, schema.UpdateAppPausedHeightQuery)
}

func (p PostgresContext) GetAppOutputAddress(operator []byte, height int64) (output []byte, err error) {
	return p.GetActorOutputAddress(operator, height, schema.AppOutputAddressQuery)
}
