package persistence

import (
	"encoding/hex"

	"github.com/pokt-network/pocket/persistence/schema"
	"github.com/pokt-network/pocket/shared/types"
)

func (p PostgresContext) GetServiceNodeExists(address []byte, height int64) (exists bool, err error) {
	return p.GetExists(address, height, schema.ServiceNodeActor.GetExistsQuery)
}

func (p PostgresContext) GetServiceNode(address []byte, height int64) (operator, publicKey, stakedTokens, serviceURL, outputAddress string, pausedHeight, unstakingHeight int64, chains []string, err error) {
	actor, err := p.GetActor(address, height, schema.ServiceNodeActor.GetQuery, schema.ServiceNodeActor.GetChainsQuery)
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
func (p PostgresContext) InsertServiceNode(address []byte, publicKey []byte, output []byte, paused bool, status int, serviceURL string, stakedTokens string, chains []string, pausedHeight int64, unstakingHeight int64) error {
	return p.InsertActor(schema.GenericActor{
		Address:         hex.EncodeToString(address),
		PublicKey:       hex.EncodeToString(publicKey),
		StakedTokens:    stakedTokens,
		GenericParam:    serviceURL,
		OutputAddress:   hex.EncodeToString(output),
		PausedHeight:    pausedHeight,
		UnstakingHeight: unstakingHeight,
		Chains:          chains,
	}, schema.ServiceNodeActor.InsertQuery)
}

// TODO(Andrew): change amount to add, to the amount to be SET
func (p PostgresContext) UpdateServiceNode(address []byte, serviceURL string, stakedTokens string, chains []string) error {
	return p.UpdateActor(schema.GenericActor{
		Address:      hex.EncodeToString(address),
		StakedTokens: stakedTokens,
		GenericParam: serviceURL,
		Chains:       chains,
	}, schema.ServiceNodeActor.UpdateQuery, schema.ServiceNodeActor.UpdateChainsQuery, schema.ServiceNodeActor.GetChainsTableName())
}

func (p PostgresContext) DeleteServiceNode(address []byte) error {
	return nil // no op
}

func (p PostgresContext) GetServiceNodeCount(chain string, height int64) (int, error) {
	panic("GetServiceNodeCount not implemented") // TODO (andrew) implement
}

// TODO(Andrew): remove status - not needed
func (p PostgresContext) GetServiceNodesReadyToUnstake(height int64, status int) (ServiceNodes []*types.UnstakingActor, err error) {
	return p.ActorReadyToUnstakeWithChains(height, schema.ServiceNodeActor.GetReadyToUnstakeQuery)
}

func (p PostgresContext) GetServiceNodeStatus(address []byte, height int64) (status int, err error) {
	return p.GetActorStatus(address, height, schema.ServiceNodeActor.GetUnstakingHeightQuery)
}

// TODO(Andrew): remove status - no longer needed
func (p PostgresContext) SetServiceNodeUnstakingHeightAndStatus(address []byte, unstakingHeight int64, status int) error {
	return p.SetActorUnstakingHeightAndStatus(address, unstakingHeight, schema.ServiceNodeActor.UpdateUnstakingHeightQuery)
}

func (p PostgresContext) GetServiceNodePauseHeightIfExists(address []byte, height int64) (int64, error) {
	return p.GetActorPauseHeightIfExists(address, height, schema.ServiceNodeActor.GetPausedHeightQuery)
}

// TODO(Andrew): remove status - it's not needed
func (p PostgresContext) SetServiceNodesStatusAndUnstakingHeightPausedBefore(pausedBeforeHeight, unstakingHeight int64, status int) error {
	return p.SetActorStatusAndUnstakingHeightPausedBefore(pausedBeforeHeight, unstakingHeight, schema.ServiceNodeActor.UpdatePausedBefore)
}

func (p PostgresContext) SetServiceNodePauseHeight(address []byte, height int64) error {
	return p.SetActorPauseHeight(address, height, schema.ServiceNodeActor.UpdatePausedHeightQuery)
}

func (p PostgresContext) GetServiceNodeOutputAddress(operator []byte, height int64) (output []byte, err error) {
	return p.GetActorOutputAddress(operator, height, schema.ServiceNodeActor.GetOutputAddressQuery)
}
