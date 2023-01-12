package defaults

import (
	"fmt"
)

const (
	defaultRPCPort       = "50832"
	defaultRPCHost       = "localhost"
	defaultRPCTimeout    = 30000
	DefaultBusBufferSize = 100
)

var (
	DefaultRemoteCLIURL = fmt.Sprintf("http://%s:%s", defaultRPCHost, defaultRPCPort)

	// consensus
	DefaultConsensusMaxMempoolBytes = uint64(500000000)
	// pacemaker
	DefaultPacemakerTimeoutMsec               = uint64(5000)
	DefaultPacemakerManual                    = true
	DefaultPacemakerDebugTimeBetweenStepsMsec = uint64(1000)
	// utility
	DefaultUtilityMaxMempoolTransactionBytes = uint64(1024 ^ 3) // 1GB V0 defaults
	DefaultUtilityMaxMempoolTransactions     = uint32(9000)
	// persistence
	DefaultPersistencePostgresUrl    = "postgres://postgres:postgres@pocket-db:5432/postgres"
	DefaultPersistenceBlockStorePath = "/var/blockstore"
	// p2p
	DefaultP2PConsensusPort         = uint32(8080)
	DefaultP2PUseRainTree           = true
	DefaultP2PIsEmptyConnectionType = false
	DefaultP2PMaxMempoolCount       = uint64(1e5)
	// telemetry
	DefaultTelemetryEnabled  = true
	DefaultTelemetryAddress  = "0.0.0.0:9000"
	DefaultTelemetryEndpoint = "/metrics"
	// logger
	DefaultLoggerLevel  = "debug"
	DefaultLoggerFormat = "pretty"
	// rpc
	DefaultRpcPort    = defaultRPCPort
	DefaultRpcTimeout = uint64(defaultRPCTimeout)
)
