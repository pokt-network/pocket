package persistence

import (
	"encoding/hex"

	"github.com/pokt-network/pocket/persistence/schema"
	"github.com/pokt-network/pocket/shared/types"
)

func (p PostgresContext) GetAppExists(address []byte, height int64) (exists bool, err error) {
	return p.GetExists(address, height, schema.ApplicationActor.GetExistsQuery)
}

func (p PostgresContext) GetApp(address []byte, height int64) (operator, publicKey, stakedTokens, maxRelays, outputAddress string, pauseHeight, unstakingHeight int64, chains []string, err error) {
	actor, err := p.GetActor(address, height, schema.ApplicationActor.GetQuery, schema.ApplicationActor.GetChainsQuery)
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
	}, schema.ApplicationActor.InsertQuery)
}

func (p PostgresContext) UpdateApp(address []byte, maxRelays string, stakedTokens string, chains []string) error {
	return p.UpdateActor(schema.GenericActor{
		Address:      hex.EncodeToString(address),
		StakedTokens: stakedTokens,
		GenericParam: maxRelays,
		Chains:       chains,
	}, schema.ApplicationActor.UpdateQuery, schema.ApplicationActor.UpdateChainsQuery, schema.ApplicationActor.GetChainsTableName())
}

func (p PostgresContext) DeleteApp(_ []byte) error {
	// NOOP
	return nil
}

// TODO(Andrew): remove status (second parameter) - not needed
func (p PostgresContext) GetAppsReadyToUnstake(height int64, _ int) (apps []*types.UnstakingActor, err error) {
	return p.ActorReadyToUnstakeWithChains(height, schema.ApplicationActor.GetReadyToUnstakeQuery)
}

func (p PostgresContext) GetAppStatus(address []byte, height int64) (status int, err error) {
	return p.GetActorStatus(address, height, schema.ApplicationActor.GetUnstakingHeightQuery)
}

// TODO(Andrew): remove status (third parameter) - no longer needed
func (p PostgresContext) SetAppUnstakingHeightAndStatus(address []byte, unstakingHeight int64, _ int) error {
	return p.SetActorUnstakingHeightAndStatus(address, unstakingHeight, schema.ApplicationActor.UpdateUnstakingHeightQuery)
}

// DISCUSS(drewsky): Need to create a semantic constant for an error return value, but should it be 0 or -1?
func (p PostgresContext) GetAppPauseHeightIfExists(address []byte, height int64) (pausedHeight int64, err error) {
	return p.GetActorPauseHeightIfExists(address, height, schema.ApplicationActor.GetPausedHeightQuery)
}

// TODO(Andrew): remove status (third parameter) - it's not needed
// DISCUSS(drewsky): This function seems to be doing too much from a naming perspective. Perhaps `SetPausedAppsToStartUnstaking`?
func (p PostgresContext) SetAppsStatusAndUnstakingHeightPausedBefore(pausedBeforeHeight, unstakingHeight int64, _ int) error {
	return p.SetActorStatusAndUnstakingHeightPausedBefore(pausedBeforeHeight, unstakingHeight, schema.ApplicationActor.UpdatePausedBefore)
}

func (p PostgresContext) SetAppPauseHeight(address []byte, height int64) error {
	return p.SetActorPauseHeight(address, height, schema.ApplicationActor.UpdatePausedHeightQuery)
}

func (p PostgresContext) GetAppOutputAddress(operator []byte, height int64) (output []byte, err error) {
	return p.GetActorOutputAddress(operator, height, schema.ApplicationActor.GetOutputAddressQuery)
}
