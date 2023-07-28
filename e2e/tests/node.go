// //go:build e2e

package e2e

import (
	"fmt"
	"os/exec"

	"github.com/pokt-network/pocket/runtime"
	"github.com/pokt-network/pocket/runtime/defaults"
)

// cliPath is the path of the binary installed and is set by the Tiltfile
const cliPath = "/usr/local/bin/p1"

var (
	// defaultRPCURL used by targetPod to build commands
	defaultRPCURL string
	// targetDevClientPod is the kube pod that executes calls to the pocket binary under test
	targetDevClientPod = "deploy/dev-cli-client"
)

func init() {
	defaultRPCHost := runtime.GetEnv("RPC_HOST", defaults.RandomValidatorEndpointK8SHostname)
	defaultRPCURL = fmt.Sprintf("http://%s:%s", defaultRPCHost, defaults.DefaultRPCPort)
}

// commandResult combines the stdout, stderr, and err of an operation
type commandResult struct {
	Stdout string
	Stderr string
	Err    error
}

// PocketClient is a single function interface for interacting with a node
type PocketClient interface {
	RunCommand(...string) (*commandResult, error)
	RunCommandOnHost(string, ...string) (*commandResult, error)
}

// Ensure that Validator fulfills PocketClient
var _ PocketClient = &nodePod{}

// nodePod holds the connection information to a specific pod in between different instructions during testing
type nodePod struct {
	targetPodName string
	result        *commandResult // stores the result of the last command that was run
}

// RunCommand runs a command on a pre-configured kube pod with the given args
func (n *nodePod) RunCommand(args ...string) (*commandResult, error) {
	return n.RunCommandOnHost(defaultRPCURL, args...)
}

// RunCommandOnHost runs a command on specified kube pod with the given args
func (n *nodePod) RunCommandOnHost(rpcUrl string, args ...string) (*commandResult, error) {
	base := []string{
		"exec", "-i", targetDevClientPod,
		"--container", "pocket",
		"--", cliPath,
		"--non_interactive=true",
		"--remote_cli_url=" + rpcUrl,
	}
	args = append(base, args...)
	cmd := exec.Command("kubectl", args...)
	r := &commandResult{}
	out, err := cmd.Output()
	r.Stdout = string(out)
	n.result = r
	// IMPROVE: make targetPodName configurable
	n.targetPodName = targetDevClientPod
	if err != nil {
		return r, err
	}
	return r, nil
}
