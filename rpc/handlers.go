package rpc

import (
	"encoding/hex"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pokt-network/pocket/app"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/utility"
)

// CONSIDER: Remove all the V1 prefixes from the RPC module

func (s *rpcServer) GetV1Health(ctx echo.Context) error {
	return ctx.NoContent(http.StatusOK)
}

func (s *rpcServer) GetV1Version(ctx echo.Context) error {
	return ctx.String(http.StatusOK, app.AppVersion)
}

func (s *rpcServer) PostV1ClientBroadcastTxSync(ctx echo.Context) error {
	txParams := new(RawTXRequest)
	if err := ctx.Bind(txParams); err != nil {
		return ctx.String(http.StatusBadRequest, "bad request")
	}

	txBz, err := hex.DecodeString(txParams.RawHexBytes)
	if err != nil {
		return ctx.String(http.StatusBadRequest, "cannot decode tx bytes")
	}

	if err = s.GetBus().GetUtilityModule().HandleTransaction(txBz); err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	if err := s.broadcastMessage(txBz); err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	return nil
}

func (s *rpcServer) GetV1ConsensusState(ctx echo.Context) error {
	consensus := s.GetBus().GetConsensusModule()
	return ctx.JSON(200, ConsensusState{
		Height: int64(consensus.CurrentHeight()),
		Round:  int64(consensus.CurrentRound()),
		Step:   int64(consensus.CurrentStep()),
	})
}

func (s *rpcServer) GetV1QueryAllChainParams(ctx echo.Context) error {
	currHeight := s.GetBus().GetConsensusModule().CurrentHeight()
	readCtx, err := s.GetBus().GetPersistenceModule().NewReadContext(int64(currHeight))
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	paramSlice, err := readCtx.GetAllParams()
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	resp := make([]Parameter, 0)
	for i := 0; i < len(paramSlice); i++ {
		resp = append(resp, Parameter{
			ParameterName:  paramSlice[i][0],
			ParameterValue: paramSlice[i][1],
		})
	}
	return ctx.JSON(200, resp)
}

// Broadcast to the entire validator set
func (s *rpcServer) broadcastMessage(msgBz []byte) error {
	utilityMsg, err := utility.PrepareTxGossipMessage(msgBz)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to prepare transaction gossip message")
		return err
	}

	if err := s.GetBus().GetP2PModule().Broadcast(utilityMsg); err != nil {
		s.logger.Error().Err(err).Msg("Failed to broadcast utility message")
		return err
	}

	return nil
}

func (s *rpcServer) GetV1P2pStakedActorsAddressBook(ctx echo.Context, params GetV1P2pStakedActorsAddressBookParams) error {
	var height int64
	var actors []Actor

	if params.Height != nil {
		height = *params.Height
	} else {
		height = int64(s.GetBus().GetConsensusModule().CurrentHeight())
	}

	readCtx, err := s.GetBus().GetPersistenceModule().NewReadContext(height)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	defer readCtx.Release()

	protocolActorGetter := getProtocolActorGetter(readCtx, params)

	protocolActors, err := protocolActorGetter(height)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	for _, protocolActor := range protocolActors {
		actors = append(actors, Actor{
			Address:    protocolActor.Address,
			Type:       protocolActorToRPCActorTypeEnum(protocolActor.ActorType),
			PublicKey:  protocolActor.PublicKey,
			ServiceUrl: protocolActor.ServiceUrl,
		})
	}

	response := P2PStakedActorsResponse{
		Actors: actors,
		Height: height,
	}

	return ctx.JSON(http.StatusOK, response)
}

// protocolActorToRPCActorTypeEnum converts a protocol actor type to the rpc actor type enum
func protocolActorToRPCActorTypeEnum(protocolActorType coreTypes.ActorType) ActorTypesEnum {
	switch protocolActorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		return Application
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		return Fisherman
	case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
		return Servicer
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		return Validator
	default:
		panic("invalid actor type")
	}
}

// getProtocolActorGetter returns the correct protocol actor getter function based on the actor type parameter
func getProtocolActorGetter(persistenceContext modules.PersistenceReadContext, params GetV1P2pStakedActorsAddressBookParams) func(height int64) ([]*coreTypes.Actor, error) {
	var protocolActorGetter func(height int64) ([]*coreTypes.Actor, error) = persistenceContext.GetAllStakedActors
	if params.ActorType == nil {
		return persistenceContext.GetAllStakedActors
	}
	switch *params.ActorType {
	case Application:
		protocolActorGetter = persistenceContext.GetAllApps
	case Fisherman:
		protocolActorGetter = persistenceContext.GetAllFishermen
	case Servicer:
		protocolActorGetter = persistenceContext.GetAllServicers
	case Validator:
		protocolActorGetter = persistenceContext.GetAllValidators
	}
	return protocolActorGetter
}
