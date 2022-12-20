package rpc

import (
	"encoding/hex"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	app "github.com/pokt-network/pocket/cmd"
	"github.com/pokt-network/pocket/internal/shared/codec"
	typesUtil "github.com/pokt-network/pocket/internal/utility/types"
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
		log.Printf("[ERROR] Failed to create Any proto from transaction gossip: %v", err)
		return err
	}

	if err := s.GetBus().GetP2PModule().Broadcast(anyUtilityMessage); err != nil {
		log.Printf("[ERROR] Failed to broadcast utility message: %v", err)
		return err
	}
	return nil
}
