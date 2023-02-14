package rpc

import (
	"encoding/hex"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pokt-network/pocket/app"
	"github.com/pokt-network/pocket/shared/codec"
	typesCore "github.com/pokt-network/pocket/shared/core/types"
	typesUtil "github.com/pokt-network/pocket/utility/types"
)

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

	if err = s.GetBus().GetUtilityModule().CheckTransaction(txBz); err != nil {
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

// Broadcast to the entire validator set
func (s *rpcServer) broadcastMessage(msgBz []byte) error {
	utilMsg := &typesUtil.TransactionGossipMessage{
		Tx: msgBz,
	}

	anyUtilityMessage, err := codec.GetCodec().ToAny(utilMsg)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to create Any proto from transaction gossip")
		return err
	}

	if err := s.GetBus().GetP2PModule().Broadcast(anyUtilityMessage); err != nil {
		s.logger.Error().Err(err).Msg("Failed to broadcast utility message")
		return err
	}
	return nil
}

func (s *rpcServer) GetV1P2pAddressBook(ctx echo.Context, params GetV1P2pAddressBookParams) error {
	var height int64
	var actors []Actor

	if params.Height != nil {
		height = *params.Height
	} else {
		height = int64(s.GetBus().GetConsensusModule().CurrentHeight())
	}

	persistenceContext, err := s.GetBus().GetPersistenceModule().NewReadContext(height)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	defer persistenceContext.Close()

	var getter func(height int64) ([]*typesCore.Actor, error)

	if params.ActorType == nil {
		getter = persistenceContext.GetAllStakedActors
	} else {
		switch *params.ActorType {
		case Application:
			getter = persistenceContext.GetAllApps
		case Fisherman:
			getter = persistenceContext.GetAllFishermen
		case ServiceNode:
			getter = persistenceContext.GetAllServiceNodes
		case Validator:
			getter = persistenceContext.GetAllValidators
		}
	}

	coreActors, err := getter(height)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	for _, coreActor := range coreActors {
		actors = append(actors, Actor{
			Address:    coreActor.Address,
			Type:       protocolActorToRPCActorTypeEnum(coreActor.ActorType),
			PublicKey:  coreActor.PublicKey,
			ServiceUrl: coreActor.GenericParam,
		})
	}

	response := P2PStakedActorsResponse{
		Actors: actors,
		Height: height,
	}

	return ctx.JSON(http.StatusOK, response)
}

func protocolActorToRPCActorTypeEnum(coreActorType typesCore.ActorType) ActorTypesEnum {
	switch coreActorType {
	case typesCore.ActorType_ACTOR_TYPE_APP:
		return Application
	case typesCore.ActorType_ACTOR_TYPE_FISH:
		return Fisherman
	case typesCore.ActorType_ACTOR_TYPE_SERVICENODE:
		return ServiceNode
	case typesCore.ActorType_ACTOR_TYPE_VAL:
		return Validator
	default:
		panic("invalid actor type")
	}
}
