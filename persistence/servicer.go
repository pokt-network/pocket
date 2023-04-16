package persistence

import (
	"encoding/hex"

	"github.com/pokt-network/pocket/persistence/types"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	moduleTypes "github.com/pokt-network/pocket/shared/modules/types"
)

func (p *PostgresContext) GetServicerExists(address []byte, height int64) (exists bool, err error) {
	return p.GetExists(types.ServicerActor, address, height)
}

func (p *PostgresContext) GetServicer(address []byte, height int64) (*coreTypes.Actor, error) {
	return p.getActor(types.ServicerActor, address, height)
}

func (p *PostgresContext) InsertServicer(address, publicKey, output []byte, _ bool, _ int32, serviceURL, stakedTokens string, chains []string, pausedHeight, unstakingHeight int64) error {
	return p.InsertActor(types.ServicerActor, &coreTypes.Actor{
		ActorType:       coreTypes.ActorType_ACTOR_TYPE_SERVICER,
		Address:         hex.EncodeToString(address),
		PublicKey:       hex.EncodeToString(publicKey),
		StakedAmount:    stakedTokens,
		ServiceUrl:      serviceURL,
		Output:          hex.EncodeToString(output),
		PausedHeight:    pausedHeight,
		UnstakingHeight: unstakingHeight,
		Chains:          chains,
	})
}

func (p *PostgresContext) UpdateServicer(address []byte, serviceURL, stakedAmount string, chains []string) error {
	return p.UpdateActor(types.ServicerActor, &coreTypes.Actor{
		ActorType:    coreTypes.ActorType_ACTOR_TYPE_SERVICER,
		Address:      hex.EncodeToString(address),
		StakedAmount: stakedAmount,
		ServiceUrl:   serviceURL,
		Chains:       chains,
	})
}

func (p *PostgresContext) GetServicerStakeAmount(height int64, address []byte) (string, error) {
	return p.getActorStakeAmount(types.ServicerActor, address, height)
}

func (p *PostgresContext) SetServicerStakeAmount(address []byte, stakeAmount string) error {
	return p.setActorStakeAmount(types.ServicerActor, address, stakeAmount)
}

func (p *PostgresContext) GetServicerCount(chain string, height int64) (int, error) {
	panic("GetServicerCount not implemented")
}

func (p *PostgresContext) GetServicersReadyToUnstake(height int64, status int32) ([]*moduleTypes.UnstakingActor, error) {
	return p.GetActorsReadyToUnstake(types.ServicerActor, height)
}

func (p *PostgresContext) GetServicerStatus(address []byte, height int64) (int32, error) {
	return p.GetActorStatus(types.ServicerActor, address, height)
}

func (p *PostgresContext) SetServicerUnstakingHeightAndStatus(address []byte, unstakingHeight int64, status int32) error {
	return p.SetActorUnstakingHeightAndStatus(types.ServicerActor, address, unstakingHeight)
}

func (p *PostgresContext) GetServicerPauseHeightIfExists(address []byte, height int64) (int64, error) {
	return p.GetActorPauseHeightIfExists(types.ServicerActor, address, height)
}

func (p *PostgresContext) SetServicerStatusAndUnstakingHeightIfPausedBefore(pausedBeforeHeight, unstakingHeight int64, status int32) error {
	return p.SetActorStatusAndUnstakingHeightIfPausedBefore(types.ServicerActor, pausedBeforeHeight, unstakingHeight)
}

func (p *PostgresContext) SetServicerPauseHeight(address []byte, height int64) error {
	return p.SetActorPauseHeight(types.ServicerActor, address, height)
}

func (p *PostgresContext) GetServicerOutputAddress(operator []byte, height int64) (output []byte, err error) {
	return p.GetActorOutputAddress(types.ServicerActor, operator, height)
}
