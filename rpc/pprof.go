//go:build pprof_enabled
// +build pprof_enabled

package rpc

import (
	"log"
	"net/http"
	_ "net/http/pprof"
)

func init() {
	go exposePProfEndpoint()
}

func exposePProfEndpoint() {
	log.Println("PProf endpoint exposed at localhost:6060")
	log.Println(http.ListenAndServe("localhost:6060", nil))
}
