package rpc

import (
	"encoding/hex"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pokt-network/pocket/app"
	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/codec"
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

	height := s.GetBus().GetConsensusModule().CurrentHeight()
	utilityCtx, err := s.GetBus().GetUtilityModule().NewContext(int64(height))
	if err != nil {
		defer func() { log.Fatalf("[ERROR] Failed to create UtilityContext: %v", err) }()
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	err = utilityCtx.CheckTransaction(txBz)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	if err := utilityCtx.Release(); err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	if err := s.GetBus().GetPersistenceModule().ReleaseWriteContext(); err != nil {
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
	utilMsg := &typesCons.UtilityMessage{
		Type: typesCons.UtilityMessageType_UTILITY_MESSAGE_TRANSACTION,
		Data: msgBz,
	}

	// anyProto, err := anypb.New(utilMsg)
	// if err != nil {
	// 	log.Fatalf("[ERROR] Failed to create Any proto: %v", err)
	// }
	anyUtilityMessage, err := codec.GetCodec().ToAny(utilMsg)
	if err != nil {
		log.Printf("[ERROR] Failed to create Any proto: %v", err)
		return err
	}

	if err := s.GetBus().GetP2PModule().Broadcast(anyUtilityMessage); err != nil {
		log.Printf("[ERROR] Failed to broadcast debug message: %v", err)
		return err
	}
	return nil
}
