package persistence

import (
	"encoding/hex"

	"github.com/pokt-network/pocket/persistence/types"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	moduleTypes "github.com/pokt-network/pocket/shared/modules/types"
)

func (p *PostgresContext) GetWatcherExists(address []byte, height int64) (exists bool, err error) {
	return p.GetExists(types.WatcherActor, address, height)
}

func (p *PostgresContext) GetWatcher(address []byte, height int64) (*coreTypes.Actor, error) {
	return p.getActor(types.WatcherActor, address, height)
}

func (p *PostgresContext) InsertWatcher(address, publicKey, output []byte, _ bool, _ int32, serviceURL, stakedTokens string, chains []string, pausedHeight, unstakingHeight int64) error {
	return p.InsertActor(types.WatcherActor, &coreTypes.Actor{
		ActorType:       coreTypes.ActorType_ACTOR_TYPE_WATCHER,
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

func (p *PostgresContext) UpdateWatcher(address []byte, serviceURL, stakedAmount string, chains []string) error {
	return p.UpdateActor(types.WatcherActor, &coreTypes.Actor{
		ActorType:    coreTypes.ActorType_ACTOR_TYPE_WATCHER,
		Address:      hex.EncodeToString(address),
		StakedAmount: stakedAmount,
		ServiceUrl:   serviceURL,
		Chains:       chains,
	})
}

func (p *PostgresContext) GetWatcherStakeAmount(height int64, address []byte) (string, error) {
	return p.getActorStakeAmount(types.WatcherActor, address, height)
}

func (p *PostgresContext) SetWatcherStakeAmount(address []byte, stakeAmount string) error {
	return p.setActorStakeAmount(types.WatcherActor, address, stakeAmount)
}

func (p *PostgresContext) GetWatchersReadyToUnstake(height int64, status int32) ([]*moduleTypes.UnstakingActor, error) {
	return p.GetActorsReadyToUnstake(types.WatcherActor, height)
}

func (p *PostgresContext) GetWatcherStatus(address []byte, height int64) (status int32, err error) {
	return p.GetActorStatus(types.WatcherActor, address, height)
}

func (p *PostgresContext) SetWatcherUnstakingHeightAndStatus(address []byte, unstakingHeight int64, status int32) error {
	return p.SetActorUnstakingHeightAndStatus(types.WatcherActor, address, unstakingHeight)
}

func (p *PostgresContext) GetWatcherPauseHeightIfExists(address []byte, height int64) (int64, error) {
	return p.GetActorPauseHeightIfExists(types.WatcherActor, address, height)
}

func (p *PostgresContext) SetWatcherStatusAndUnstakingHeightIfPausedBefore(pausedBeforeHeight, unstakingHeight int64, status int32) error {
	return p.SetActorStatusAndUnstakingHeightIfPausedBefore(types.WatcherActor, pausedBeforeHeight, unstakingHeight)
}

func (p *PostgresContext) SetWatcherPauseHeight(address []byte, height int64) error {
	return p.SetActorPauseHeight(types.WatcherActor, address, height)
}

func (p *PostgresContext) GetWatcherOutputAddress(operator []byte, height int64) (output []byte, err error) {
	return p.GetActorOutputAddress(types.WatcherActor, operator, height)
}
