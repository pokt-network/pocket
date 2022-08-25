package rpc

import (
	"log"
	"net/http"
	"time"

	"github.com/pokt-network/pocket/app"
	"github.com/pokt-network/pocket/shared"
)

var APIVersion = app.AppVersion

type rpcServer struct {
	node *shared.Node
}

func NewRPCServer(pocketNode *shared.Node) *rpcServer {
	return &rpcServer{
		node: pocketNode,
	}
}

func (s *rpcServer) StartRPC(port string, timeout uint64) {
	log.Printf("Starting RPC on port %s...\n", port)

	routes := s.GetRoutes()

	srv := &http.Server{
		ReadTimeout:       30 * time.Second,
		ReadHeaderTimeout: 20 * time.Second,
		WriteTimeout:      60 * time.Second,
		Addr:              ":" + port,
		Handler:           http.TimeoutHandler(Router(routes), time.Duration(timeout)*time.Millisecond, "Server Timeout Handling Request"),
	}
	log.Fatal(srv.ListenAndServe())
}
