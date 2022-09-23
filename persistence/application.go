package persistence

import (
	"encoding/hex"
	"log"

	"github.com/pokt-network/pocket/persistence/types"
	"github.com/pokt-network/pocket/shared/modules"
)

func (p PostgresContext) GetAppExists(address []byte, height int64) (exists bool, err error) {
	return p.GetExists(types.ApplicationActor, address, height)
}

func (p PostgresContext) GetApp(address []byte, height int64) (operator, publicKey, stakedTokens, maxRelays, outputAddress string, pauseHeight, unstakingHeight int64, chains []string, err error) {
	actor, err := p.GetActor(types.ApplicationActor, address, height)
	operator = actor.Address
	publicKey = actor.PublicKey
	stakedTokens = actor.StakedTokens
	maxRelays = actor.ActorSpecificParam
	outputAddress = actor.OutputAddress
	pauseHeight = actor.PausedHeight
	unstakingHeight = actor.UnstakingHeight
	chains = actor.Chains
	return
}

func (p PostgresContext) InsertApp(address []byte, publicKey []byte, output []byte, paused bool, status int32, maxRelays string, stakedTokens string, chains []string, pausedHeight int64, unstakingHeight int64) error {
	return p.InsertActor(types.ApplicationActor, types.BaseActor{
		Address:            hex.EncodeToString(address),
		PublicKey:          hex.EncodeToString(publicKey),
		StakedTokens:       stakedTokens,
		ActorSpecificParam: maxRelays,
		OutputAddress:      hex.EncodeToString(output),
		PausedHeight:       pausedHeight,
		UnstakingHeight:    unstakingHeight,
		Chains:             chains,
	})
}

func (p PostgresContext) UpdateApp(address []byte, maxRelays string, stakedAmount string, chains []string) error {
	return p.UpdateActor(types.ApplicationActor, types.BaseActor{
		Address:            hex.EncodeToString(address),
		StakedTokens:       stakedAmount,
		ActorSpecificParam: maxRelays,
		Chains:             chains,
	})
}

func (p PostgresContext) GetAppStakeAmount(height int64, address []byte) (string, error) {
	return p.GetActorStakeAmount(types.ApplicationActor, address, height)
}

func (p PostgresContext) SetAppStakeAmount(address []byte, stakeAmount string) error {
	return p.SetActorStakeAmount(types.ApplicationActor, address, stakeAmount)
}

func (p PostgresContext) DeleteApp(_ []byte) error {
	log.Println("[DEBUG] DeleteApp is a NOOP")
	return nil
}

func (p PostgresContext) GetAppsReadyToUnstake(height int64, status int32) ([]modules.IUnstakingActor, error) {
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
