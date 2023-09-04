package client

import (
	"errors"
	"time"

	light_client_types "github.com/pokt-network/pocket/ibc/client/light_clients/types"
	"github.com/pokt-network/pocket/ibc/client/types"
	ibc_types "github.com/pokt-network/pocket/ibc/types"
	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/modules"
	util_types "github.com/pokt-network/pocket/utility/types"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
)

// GetHostConsensusState returns the ConsensusState at the given height for the
// host chain, the Pocket network. It then serialises this and packs it into a
// ConsensusState object for use in a WASM client
func (c *clientManager) GetHostConsensusState(height modules.Height) (modules.ConsensusState, error) {
	blockStore := c.GetBus().GetPersistenceModule().GetBlockStore()
	block, err := blockStore.GetBlock(height.GetRevisionHeight())
	if err != nil {
		return nil, err
	}
	pocketConsState := &light_client_types.PocketConsensusState{
		Timestamp:       block.BlockHeader.Timestamp,
		StateHash:       block.BlockHeader.StateHash,
		StateTreeHashes: block.BlockHeader.StateTreeHashes,
		NextValSetHash:  block.BlockHeader.NextValSetHash,
	}
	consBz, err := codec.GetCodec().Marshal(pocketConsState)
	if err != nil {
		return nil, err
	}
	return types.NewConsensusState(consBz, uint64(pocketConsState.Timestamp.AsTime().UnixNano())), nil
}

// GetHostClientState returns the ClientState at the given height for the host
// chain, the Pocket network.
//
// This function is used to validate the state of a client running on a
// counterparty chain.
func (c *clientManager) GetHostClientState(height modules.Height) (modules.ClientState, error) {
	blockStore := c.GetBus().GetPersistenceModule().GetBlockStore()
	block, err := blockStore.GetBlock(height.GetRevisionHeight())
	if err != nil {
		return nil, err
	}
	rCtx, err := c.GetBus().GetPersistenceModule().NewReadContext(int64(height.GetRevisionHeight()))
	if err != nil {
		return nil, err
	}
	defer rCtx.Release()
	unbondingBlocks, err := rCtx.GetIntParam(util_types.ValidatorUnstakingBlocksParamName, int64(height.GetRevisionHeight()))
	if err != nil {
		return nil, err
	}
	// TODO_AFTER(#705): use the actual MinimumBlockTime once set
	blockTime := time.Minute * 15
	unbondingPeriod := blockTime * time.Duration(unbondingBlocks) // approx minutes per block * blocks
	pocketClient := &light_client_types.PocketClientState{
		NetworkId:       block.BlockHeader.NetworkId,
		TrustLevel:      &light_client_types.Fraction{Numerator: 2, Denominator: 3},
		TrustingPeriod:  durationpb.New(unbondingPeriod),
		UnbondingPeriod: durationpb.New(unbondingPeriod),
		MaxClockDrift:   durationpb.New(blockTime), // DISCUSS: What is a reasonable MaxClockDrift?
		LatestHeight: &types.Height{
			RevisionNumber: height.GetRevisionNumber(),
			RevisionHeight: height.GetRevisionHeight(),
		},
		ProofSpec: ibc_types.SmtSpec,
	}
	clientBz, err := codec.GetCodec().Marshal(pocketClient)
	if err != nil {
		return nil, err
	}
	return &types.ClientState{
		Data:         clientBz,
		RecentHeight: pocketClient.LatestHeight,
	}, nil
}

// VerifyHostClientState verifies that a ClientState for a light client running
// on a counterparty chain is valid, by checking it against the result of
// GetHostClientState(counterpartyClientState.GetLatestHeight())
func (c *clientManager) VerifyHostClientState(counterparty modules.ClientState) error {
	height, err := c.GetCurrentHeight()
	if err != nil {
		return err
	}
	hostState, err := c.GetHostClientState(height)
	if err != nil {
		return err
	}
	poktHost := new(light_client_types.PocketClientState)
	err = codec.GetCodec().Unmarshal(hostState.GetData(), poktHost)
	if err != nil {
		return err
	}
	poktCounter := new(light_client_types.PocketClientState)
	err = codec.GetCodec().Unmarshal(counterparty.GetData(), poktCounter)
	if err != nil {
		return errors.New("counterparty client state is not a PocketClientState")
	}

	if poktCounter.FrozenHeight > 0 {
		return errors.New("counterparty client state is frozen")
	}
	if poktCounter.NetworkId != poktHost.NetworkId {
		return errors.New("counterparty client state has different network id")
	}
	if poktCounter.LatestHeight.RevisionNumber != poktHost.LatestHeight.RevisionNumber {
		return errors.New("counterparty client state has different revision number")
	}
	if poktCounter.GetLatestHeight().GTE(poktHost.GetLatestHeight()) {
		return errors.New("counterparty client state has a height greater than or equal to the host client state")
	}
	if poktCounter.TrustLevel.LT(&light_client_types.Fraction{Numerator: 2, Denominator: 3}) ||
		poktCounter.TrustLevel.GT(&light_client_types.Fraction{Numerator: 1, Denominator: 1}) {
		return errors.New("counterparty client state trust level is not in the accepted range")
	}
	if !proto.Equal(poktCounter.ProofSpec, poktHost.ProofSpec) {
		return errors.New("counterparty client state has different proof spec")
	}
	if poktCounter.UnbondingPeriod != poktHost.UnbondingPeriod {
		return errors.New("counterparty client state has different unbonding period")
	}
	if poktCounter.UnbondingPeriod.AsDuration().Nanoseconds() < poktHost.TrustingPeriod.AsDuration().Nanoseconds() {
		return errors.New("counterparty client state unbonding period is less than trusting period")
	}

	// RESEARCH: Look into upgrade paths, their use and if they should just be equal

	return nil
}

// GetCurrentHeight returns the current IBC client height of the network
// TODO_AFTER(#882): Use actual revision number
func (h *clientManager) GetCurrentHeight() (modules.Height, error) {
	currHeight := h.GetBus().GetConsensusModule().CurrentHeight()
	rCtx, err := h.GetBus().GetPersistenceModule().NewReadContext(int64(currHeight))
	if err != nil {
		return nil, err
	}
	defer rCtx.Release()
	revNum := rCtx.GetRevisionNumber(int64(currHeight))
	return &types.Height{
		RevisionNumber: revNum,
		RevisionHeight: currHeight,
	}, nil
}
