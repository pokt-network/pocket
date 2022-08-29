package rpc

import (
	"encoding/hex"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type RpcServer interface {
	TODOHandler(http.ResponseWriter, *http.Request, httprouter.Params)
	HealthHandler(http.ResponseWriter, *http.Request, httprouter.Params)
	VersionHandler(http.ResponseWriter, *http.Request, httprouter.Params)
	BroadcastRawTxSyncHandler(http.ResponseWriter, *http.Request, httprouter.Params)
}

var _ RpcServer = &rpcServer{}

func (s *rpcServer) TODOHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	WriteErrorResponse(w, 503, "not implemented")
}

func (s *rpcServer) HealthHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	WriteOKResponse(w)
}

func (s *rpcServer) VersionHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	WriteResponse(w, APIVersion, r.URL.Path, r.Host)
}

func (s *rpcServer) BroadcastRawTxSyncHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	params := SendRawTxParams{}
	if err := PopModel(w, r, ps, &params); err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	bz, err := hex.DecodeString(params.RawHexBytes)
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	bus := s.node.GetBus()
	height := bus.GetConsensusModule().CurrentHeight()
	uCtx, err := bus.GetUtilityModule().NewContext(int64(height))
	if err != nil {
		log.Fatalf("[ERROR] Failed to create UtilityContext: %v", err)
	}
	err = uCtx.CheckTransaction(bz)
	if err != nil {
		WriteErrorResponse(w, 500, err.Error())
		return
	}

	WriteOKResponse(w)
}
