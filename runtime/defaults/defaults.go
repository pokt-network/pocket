package defaults

import (
	"fmt"

	"github.com/pokt-network/pocket/runtime/configs/types"
)

const (
	DefaultRPCPort                  = "50832"
	DefaultBusBufferSize            = 100
	DefaultRPCHost                  = "localhost"
	Validator1EndpointDockerCompose = "node1.consensus"
	Validator1EndpointK8S           = "v1-validator001"

	defaultRPCTimeout = 30000
)

var (
	DefaultRemoteCLIURL = fmt.Sprintf("http://%s:%s", DefaultRPCHost, DefaultRPCPort)
	DefaultUseLibp2p    = false

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
	DefaultP2PPort            = uint32(42069)
	DefaultP2PUseRainTree     = true
	DefaultP2PConnectionType  = types.ConnectionType_TCPConnection
	DefaultP2PMaxMempoolCount = uint64(1e5)
	// telemetry
	DefaultTelemetryEnabled  = true
	DefaultTelemetryAddress  = "0.0.0.0:9000"
	DefaultTelemetryEndpoint = "/metrics"
	// logger
	DefaultLoggerLevel  = "debug"
	DefaultLoggerFormat = "pretty"
	// rpc
	DefaultRPCTimeout = uint64(defaultRPCTimeout)
)
