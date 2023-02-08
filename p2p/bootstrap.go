package p2p

import (
	"fmt"
	"strings"

	"github.com/pokt-network/pocket/runtime/defaults"
)

const (
	validator1EndpointDockerCompose = "node1.consensus"
	validator1EndpointK8S           = "v1-validator001"
)

// defaultBootstrapNodesCsv is a list of nodes to bootstrap the network with. By convention, for now, the first validator will provide bootstrapping facilities.
//
// In LocalNet, the developer will have only one of the two stack online, therefore this is also a poor's man way to simulate the scenario in which a boostrap node is offline.
var defaultBootstrapNodesCsv = strings.Join([]string{
	fmt.Sprintf("http://%s:%s", validator1EndpointDockerCompose, defaults.DefaultRPCPort),
	fmt.Sprintf("http://%s:%s", validator1EndpointK8S, defaults.DefaultRPCPort),
}, ",")
