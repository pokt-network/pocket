//go:build e2e

package e2e

import (
	"fmt"
	"os/exec"

	"github.com/pokt-network/pocket/runtime"
	"github.com/pokt-network/pocket/runtime/defaults"
)

// rpcURL of the pod that the test harness drives
var rpcURL string

func init() {
	rpcURL = fmt.Sprintf("http://%s:%s", runtime.GetEnv("RPC_HOST", "pockets"), defaults.DefaultRPCPort)
}

// cliPath is the path of the binary installed and is set by the Tiltfile
const cliPath = "/usr/local/bin/client"

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

// RunCommand runs a command on the pocket binary
func (v *validatorPod) RunCommand(args ...string) (*commandResult, error) {
	base := []string{
		"exec", "-i", "deploy/pocket-v1-cli-client",
		"--container", "pocket",
		"--", cliPath,
		"--non_interactive=true",
		"--remote_cli_url=" + rpcURL,
	}
	args = append(base, args...)
	cmd := exec.Command("kubectl", args...)
	r := &commandResult{}
	out, err := cmd.Output()
	v.result = r
	r.Stdout = string(out)
	if err != nil {
		return r, err
	}
	return r, nil
}
