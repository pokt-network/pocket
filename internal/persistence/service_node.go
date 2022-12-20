package persistence

import (
	"encoding/hex"

	"github.com/pokt-network/pocket/internal/persistence/types"
	"github.com/pokt-network/pocket/internal/shared/modules"
)

func (p PostgresContext) GetServiceNodeExists(address []byte, height int64) (exists bool, err error) {
	return p.GetExists(types.ServiceNodeActor, address, height)
}

func (p PostgresContext) GetServiceNode(address []byte, height int64) (operator, publicKey, stakedTokens, serviceURL, outputAddress string, pausedHeight, unstakingHeight int64, chains []string, err error) {
	actor, err := p.getActor(types.ServiceNodeActor, address, height)
	operator = actor.Address
	publicKey = actor.PublicKey
	stakedTokens = actor.StakedAmount
	serviceURL = actor.GenericParam
	outputAddress = actor.Output
	pausedHeight = actor.PausedHeight
	unstakingHeight = actor.UnstakingHeight
	chains = actor.Chains
	return
}

func (p PostgresContext) InsertServiceNode(address []byte, publicKey []byte, output []byte, _ bool, _ int32, serviceURL string, stakedTokens string, chains []string, pausedHeight int64, unstakingHeight int64) error {
	return p.InsertActor(types.ServiceNodeActor, &types.Actor{
		ActorType:       types.ActorType_Node,
		Address:         hex.EncodeToString(address),
		PublicKey:       hex.EncodeToString(publicKey),
		StakedAmount:    stakedTokens,
		GenericParam:    serviceURL,
		Output:          hex.EncodeToString(output),
		PausedHeight:    pausedHeight,
		UnstakingHeight: unstakingHeight,
		Chains:          chains,
	})
}

func (p PostgresContext) UpdateServiceNode(address []byte, serviceURL string, stakedAmount string, chains []string) error {
	return p.UpdateActor(types.ServiceNodeActor, &types.Actor{
		ActorType:    types.ActorType_Node,
		Address:      hex.EncodeToString(address),
		StakedAmount: stakedAmount,
		GenericParam: serviceURL,
		Chains:       chains,
	})
}

func (p PostgresContext) GetServiceNodeStakeAmount(height int64, address []byte) (string, error) {
	return p.getActorStakeAmount(types.ServiceNodeActor, address, height)
}

func (p PostgresContext) SetServiceNodeStakeAmount(address []byte, stakeAmount string) error {
	return p.setActorStakeAmount(types.ServiceNodeActor, address, stakeAmount)
}

func (p PostgresContext) GetServiceNodeCount(chain string, height int64) (int, error) {
	panic("GetServiceNodeCount not implemented")
}

func (p PostgresContext) GetServiceNodesReadyToUnstake(height int64, status int32) ([]modules.IUnstakingActor, error) {
	return p.GetActorsReadyToUnstake(types.ServiceNodeActor, height)
}

func (p PostgresContext) GetServiceNodeStatus(address []byte, height int64) (int32, error) {
	return p.GetActorStatus(types.ServiceNodeActor, address, height)
}

func (p PostgresContext) SetServiceNodeUnstakingHeightAndStatus(address []byte, unstakingHeight int64, status int32) error {
	return p.SetActorUnstakingHeightAndStatus(types.ServiceNodeActor, address, unstakingHeight)
}

func (p PostgresContext) GetServiceNodePauseHeightIfExists(address []byte, height int64) (int64, error) {
	return p.GetActorPauseHeightIfExists(types.ServiceNodeActor, address, height)
}

func (p PostgresContext) SetServiceNodeStatusAndUnstakingHeightIfPausedBefore(pausedBeforeHeight, unstakingHeight int64, status int32) error {
	return p.SetActorStatusAndUnstakingHeightIfPausedBefore(types.ServiceNodeActor, pausedBeforeHeight, unstakingHeight)
}

func (p PostgresContext) SetServiceNodePauseHeight(address []byte, height int64) error {
	return p.SetActorPauseHeight(types.ServiceNodeActor, address, height)
}

func (p PostgresContext) GetServiceNodeOutputAddress(operator []byte, height int64) (output []byte, err error) {
	return p.GetActorOutputAddress(types.ServiceNodeActor, operator, height)
}
