//go:build e2e

package e2e

import (
	"fmt"
	"os/exec"

	"github.com/pokt-network/pocket/runtime"
	"github.com/pokt-network/pocket/runtime/defaults"
)

var (
	// rpcURL used by targetPod to build commands
	rpcURL string
	// targetPod is the kube pod that executes calls to the pocket binary under test
	targetPod = "deploy/dev-cli-client"
)

func init() {
	// set the rpcURL based on the environment while still supporting overriding of the RPC_HOST.
	var rpcHost string
	// TECHDEBT: if we intend to support running the e2e tests in both tilt/k8s and
	// docker compose, then we need to understand why the `KUBERNETES_SERVICE_HOST`
	// isn't set even when running in k8s. See: `test_go('e2e-tests', ...` in the
	// Tiltfile.
	// if runtime.IsProcessRunningInsideKubernetes() {
	rpcHost = runtime.GetEnv("RPC_HOST", defaults.RandomValidatorEndpointK8SHostname)
	// } else {
	// 	rpcHost = runtime.GetEnv("RPC_HOST", defaults.Validator1EndpointDockerComposeHostname)
	// }

	rpcURL = fmt.Sprintf("http://%s:%s", rpcHost, defaults.DefaultRPCPort)
}

// cliPath is the path of the binary installed and is set by the Tiltfile
const cliPath = "/usr/local/bin/p1"

// commandResult combines the stdout, stderr, and err of an operation
type commandResult struct {
	Stdout string
	Stderr string
	Err    error
}

// PocketClient is a single function interface for interacting with a node
type PocketClient interface {
	RunCommand(...string) (*commandResult, error)
}

// Ensure that Validator fulfills PocketClient
var _ PocketClient = &validatorPod{}

// validatorPod holds the connection information to pod validator-001 for testing
type validatorPod struct {
	result *commandResult // stores the result of the last command that was run
}

// RunCommand runs a command on a target kube pod
func (v *validatorPod) RunCommand(args ...string) (*commandResult, error) {
	base := []string{
		"exec", "-i", targetPod,
		"--container", "pocket",
		"--", cliPath,
		"--non_interactive=true",
		"--remote_cli_url=" + rpcURL,
	}
	args = append(base, args...)
	cmd := exec.Command("kubectl", args...)
	r := &commandResult{}
	out, err := cmd.Output()
	r.Stdout = string(out)
	v.result = r
	if err != nil {
		return r, err
	}
	return r, nil
}
