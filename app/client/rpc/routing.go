package rpc

import "github.com/julienschmidt/httprouter"

func Router(routes Routes) *httprouter.Router {
	router := httprouter.New()
	for _, route := range routes {
		router.Handle(route.Method, route.Path, route.HandlerFunc)
	}
	return router
}

type Route struct {
	Name        string
	Method      string
	Path        string
	HandlerFunc httprouter.Handle
}

type Routes []Route

func (s *rpcServer) GetRoutes() Routes {
	routes := Routes{
		// System routes
		Route{Name: "Health", Method: "POST", Path: "/v1/health", HandlerFunc: s.HealthHandler},
		Route{Name: "Version", Method: "POST", Path: "/v1/version", HandlerFunc: s.VersionHandler},

		// Commands
		Route{Name: "BroadcastTxSync", Method: "POST", Path: "/v1/client/broadcast_tx_sync", HandlerFunc: s.BroadcastRawTxSyncHandler},
		Route{Name: "BroadcastTxAsync", Method: "POST", Path: "/v1/client/broadcast_tx_async", HandlerFunc: s.TODOHandler},   //TODO (team)
		Route{Name: "BroadcastTxCommit", Method: "POST", Path: "/v1/client/broadcast_tx_commit", HandlerFunc: s.TODOHandler}, //TODO (team)

		// Queries
		Route{Name: "QueryTx", Method: "POST", Path: "/v1/query/tx", HandlerFunc: s.TODOHandler}, //TODO (team)
	}
	return routes
}
