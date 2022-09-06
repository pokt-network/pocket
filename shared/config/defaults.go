package config

import "fmt"

const (
	DefaultRPCPort    = "50832"
	DefaultRPCHost    = "localhost"
	DefaultRPCTimeout = 30000
)

var DefaultRemoteCLIURL = fmt.Sprintf("http://%s:%s", DefaultRPCHost, DefaultRPCPort)
