package rpc

import (
	"github.com/julienschmidt/httprouter"
)

const (
	HealthRoute            routeKey = "Health"
	VersionRoute           routeKey = "Version"
	BroadcastTxSyncRoute   routeKey = "BroadcastTxSync"
	BroadcastTxAsyncRoute  routeKey = "BroadcastTxAsync"
	BroadcastTxCommitRoute routeKey = "BroadcastTxCommit"
	QueryTxRoute           routeKey = "QueryTx"
)

var RoutesMap = map[routeKey]Route{
	// System routes
	HealthRoute:  {Method: "GET", Path: "/v1/health", HandlerFunc: func(rs RpcServer) httprouter.Handle { return rs.HealthHandler }},
	VersionRoute: {Method: "GET", Path: "/v1/version", HandlerFunc: func(rs RpcServer) httprouter.Handle { return rs.VersionHandler }},

	// Commands
	BroadcastTxSyncRoute:   {Method: "POST", Path: "/v1/client/broadcast_tx_sync", HandlerFunc: func(rs RpcServer) httprouter.Handle { return rs.BroadcastRawTxSyncHandler }},
	BroadcastTxAsyncRoute:  {Method: "POST", Path: "/v1/client/broadcast_tx_async", HandlerFunc: func(rs RpcServer) httprouter.Handle { return rs.TODOHandler }},  // TODO (team)
	BroadcastTxCommitRoute: {Method: "POST", Path: "/v1/client/broadcast_tx_commit", HandlerFunc: func(rs RpcServer) httprouter.Handle { return rs.TODOHandler }}, // TODO (team)

	// Queries
	QueryTxRoute: {Method: "POST", Path: "/v1/query/tx", HandlerFunc: func(rs RpcServer) httprouter.Handle { return rs.TODOHandler }}, // TODO (team)
}

type (
	routeKey string
	Route    struct {
		Method      string
		Path        string
		HandlerFunc func(RpcServer) httprouter.Handle
	}
	Routes []Route
)

func (s *rpcServer) Router(routes Routes) *httprouter.Router {
	router := httprouter.New()
	for _, route := range routes {
		router.Handle(route.Method, route.Path, route.HandlerFunc(s))
	}
	return router
}

func (s *rpcServer) GetRoutes() Routes {
	routes := Routes{}
	for _, route := range RoutesMap {
		routes = append(routes, route)
	}
	return routes
}
