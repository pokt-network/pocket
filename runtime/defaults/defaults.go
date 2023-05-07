package defaults

import (
	"fmt"
	"os"

	"github.com/pokt-network/pocket/runtime/configs/types"
)

func init() {
	initDefaultRootDirectory()
}

func initDefaultRootDirectory() {
	// use home directory + /.pocket as root directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	DefaultRootDirectory = homeDir + "/.pocket"
	// IMPROVE: this is a hack to get around the fact that we don't know the home directory until after the init function has run
	DefaultKeybaseFilePath = DefaultRootDirectory + "/keys"
}

const (
	DefaultRPCPort                  = "50832"
	DefaultBusBufferSize            = 100
	DefaultRPCHost                  = "localhost"
	Validator1EndpointDockerCompose = "node1.consensus"
	Validator1EndpointK8S           = "validator-001-pocket"
)

var (
	// DefaultRootDirectory is root directory for the pocket node is initialized in the init function to be the home directory + /.pocket
	DefaultRootDirectory = ""

	// NetworkID
	DefaultNetworkID = "localnet"
	// consensus
	DefaultConsensusMaxMempoolBytes = uint64(500000000)
	// pacemaker
	DefaultPacemakerTimeoutMsec               = uint64(10000)
	DefaultPacemakerManual                    = true
	DefaultPacemakerDebugTimeBetweenStepsMsec = uint64(1000)
	// utility
	DefaultUtilityMaxMempoolTransactionBytes = uint64(1024 ^ 3) // 1GB V0 defaults
	DefaultUtilityMaxMempoolTransactions     = uint32(9000)
	// persistence
	DefaultPersistencePostgresURL    = "postgres://postgres:postgres@pocket-db:5432/postgres"
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
	DefaultRPCTimeout = uint64(30000)

	// keybase
	DefaultKeybaseType     = types.KeybaseType_FILE
	DefaultKeybaseFilePath = "" // set in init function
	// vault
	DefaultKeybaseVaultAddr      = ""
	DefaultKeybaseVaultToken     = ""
	DefaultKeybaseVaultMountPath = ""
)

var (
	DefaultRemoteCLIURL = fmt.Sprintf("http://%s:%s", DefaultRPCHost, DefaultRPCPort)
	// DefaultP2PBootstrapNodesCsv is a list of nodes to bootstrap the network with. By convention, for now, the first validator will provide bootstrapping facilities.
	//
	// In LocalNet, the developer will have only one of the two stack online, therefore this is also a poor's man way to simulate the scenario in which a boostrap node is offline.
	DefaultP2PBootstrapNodesCsv = fmt.Sprintf("%s,%s",
		fmt.Sprintf("http://%s:%s", Validator1EndpointDockerCompose, DefaultRPCPort),
		fmt.Sprintf("http://%s:%s", Validator1EndpointK8S, DefaultRPCPort),
	)
)
