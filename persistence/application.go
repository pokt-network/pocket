package persistence

import (
	"encoding/hex"

	"github.com/pokt-network/pocket/persistence/types"
	"github.com/pokt-network/pocket/shared/modules"
)

func (p PostgresContext) GetAppExists(address []byte, height int64) (exists bool, err error) {
	return p.GetExists(types.ApplicationActor, address, height)
}

func (p PostgresContext) GetApp(address []byte, height int64) (operator, publicKey, stakedTokens, maxRelays, outputAddress string, pauseHeight, unstakingHeight int64, chains []string, err error) {
	actor, err := p.getActor(types.ApplicationActor, address, height)
	if err != nil {
		return
	}
	operator = actor.Address
	publicKey = actor.PublicKey
	stakedTokens = actor.StakedAmount
	maxRelays = actor.GenericParam
	outputAddress = actor.Output
	pauseHeight = actor.PausedHeight
	unstakingHeight = actor.UnstakingHeight
	chains = actor.Chains
	return
}

func (p PostgresContext) InsertApp(address []byte, publicKey []byte, output []byte, _ bool, _ int32, maxRelays string, stakedTokens string, chains []string, pausedHeight int64, unstakingHeight int64) error {
	return p.InsertActor(types.ApplicationActor, &types.Actor{
		ActorType:       types.ActorType_App,
		Address:         hex.EncodeToString(address),
		PublicKey:       hex.EncodeToString(publicKey),
		Chains:          chains,
		GenericParam:    maxRelays,
		StakedAmount:    stakedTokens,
		PausedHeight:    pausedHeight,
		UnstakingHeight: unstakingHeight,
		Output:          hex.EncodeToString(output),
	})
}

func (p PostgresContext) UpdateApp(address []byte, maxRelays string, stakedAmount string, chains []string) error {
	return p.UpdateActor(types.ApplicationActor, &types.Actor{
		ActorType:    types.ActorType_App,
		Address:      hex.EncodeToString(address),
		Chains:       chains,
		GenericParam: maxRelays,
		StakedAmount: stakedAmount,
	})
}

func (p PostgresContext) GetAppStakeAmount(height int64, address []byte) (string, error) {
	return p.getActorStakeAmount(types.ApplicationActor, address, height)
}

func (p PostgresContext) SetAppStakeAmount(address []byte, stakeAmount string) error {
	return p.setActorStakeAmount(types.ApplicationActor, address, stakeAmount)
}

func (p PostgresContext) GetAppsReadyToUnstake(height int64, _ int32) ([]modules.IUnstakingActor, error) {
	return p.GetActorsReadyToUnstake(types.ApplicationActor, height)
}

func (p PostgresContext) GetAppStatus(address []byte, height int64) (int32, error) {
	return p.GetActorStatus(types.ApplicationActor, address, height)
}

func (p PostgresContext) SetAppUnstakingHeightAndStatus(address []byte, unstakingHeight int64, status int32) error {
	return p.SetActorUnstakingHeightAndStatus(types.ApplicationActor, address, unstakingHeight)
}

func (p PostgresContext) GetAppPauseHeightIfExists(address []byte, height int64) (int64, error) {
	return p.GetActorPauseHeightIfExists(types.ApplicationActor, address, height)
}

func (p PostgresContext) SetAppStatusAndUnstakingHeightIfPausedBefore(pausedBeforeHeight, unstakingHeight int64, status int32) error {
	return p.SetActorStatusAndUnstakingHeightIfPausedBefore(types.ApplicationActor, pausedBeforeHeight, unstakingHeight)
}

func (p PostgresContext) SetAppPauseHeight(address []byte, height int64) error {
	return p.SetActorPauseHeight(types.ApplicationActor, address, height)
}

func (p PostgresContext) GetAppOutputAddress(operator []byte, height int64) ([]byte, error) {
	return p.GetActorOutputAddress(types.ApplicationActor, operator, height)
}
