package rpc

import (
	"encoding/hex"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pokt-network/pocket/app"
)

func (s *rpcServer) GetV1Health(ctx echo.Context) error {
	return ctx.NoContent(http.StatusOK)
}

func (s *rpcServer) GetV1Version(ctx echo.Context) error {
	return ctx.String(http.StatusOK, app.AppVersion)
}

func (s *rpcServer) PostV1ClientBroadcastTxSync(ctx echo.Context) error {
	params := new(RawTXRequest)
	if err := ctx.Bind(params); err != nil {
		return ctx.String(http.StatusBadRequest, "bad request")
	}
	bz, err := hex.DecodeString(params.RawHexBytes)
	if err != nil {
		return ctx.String(http.StatusBadRequest, "cannot decode tx bytes")
	}
	height := s.GetBus().GetConsensusModule().CurrentHeight()
	uCtx, err := s.GetBus().GetUtilityModule().NewContext(int64(height))
	if err != nil {
		defer func() { log.Fatalf("[ERROR] Failed to create UtilityContext: %v", err) }()
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	err = uCtx.CheckTransaction(bz)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	if err := uCtx.Release(); err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	if err := s.GetBus().GetPersistenceModule().ReleaseWriteContext(); err != nil {
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
func (s *rpcServer) broadcastDebugMessage(msgBz []byte) {

	m := &debug.DebugMessage{
		Action:  debug.DebugMessageAction_DEBUG_SHOW_LATEST_BLOCK_IN_STORE,
		Message: nil
	}

	anyProto, err := anypb.New(debugMsg)
	if err != nil {
		log.Fatalf("[ERROR] Failed to create Any proto: %v", err)
	}

	// TODO(olshansky): Once we implement the cleanup layer in RainTree, we'll be able to use
	// broadcast. The reason it cannot be done right now is because this client is not in the
	// address book of the actual validator nodes, so `node1.consensus` never receives the message.
	// p2pMod.Broadcast(anyProto, debug.PocketTopic_DEBUG_TOPIC)

	for _, val := range consensusMod.ValidatorMap() {
		addr, err := pocketCrypto.NewAddress(val.GetAddress())
		if err != nil {
			log.Fatalf("[ERROR] Failed to convert validator address into pocketCrypto.Address: %v", err)
		}
		p2pMod.Send(addr, anyProto, debug.PocketTopic_DEBUG_TOPIC)
	}

	s.GetBus().GetP2PModule().Broadcast(anyProto, debug.PocketTopic_DEBUG_TOPIC)
}
