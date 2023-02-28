//
// //go:build e2e

package e2e

import (
	"fmt"
	"os/exec"

	"github.com/pokt-network/pocket/runtime"
	"github.com/pokt-network/pocket/runtime/defaults"
)

func init() {
	// NB: the defaults package exports several common & useful defaults for localnet interaction
	rpcURL = fmt.Sprintf("http://%s:%s", runtime.GetEnv("RPC_HOST", "v1-validator001"), defaults.DefaultRPCPort)
}

// rpcURL of the debug client that the test harness drives.
var rpcURL string

const cliPath = "/usr/local/bin/client"

// CommandResult combines the stdout, stderr, and err of an operation.
type CommandResult struct {
	Stdout string
	Stderr string
	Err    error
}

// PocketClient is a single function interface for interacting with a node.
type PocketClient interface {
	RunCommand(...string) (*CommandResult, error)
}

// TODO: we could collect these into a service map to keep them hidden from global services
var _ PocketClient = &KubeClient{}

// TODO_IN_THIS_COMMIT: take this out into its own file. Adapters should be in their own files.
var _ PocketClient = &Validator{}

// Validator holds the connection information for validator-001 for testing.
type Validator struct {
	result *CommandResult // stores the result of the last command that was run.
}

// RunCommand runs a command on the v1-cli-client binary
func (v *Validator) RunCommand(args ...string) (*CommandResult, error) {
	base := []string{
		"exec", "-it", "deploy/pocket-v1-cli-client",
		"--container", "pocket",
		"--", cliPath,
		"--non_interactive=true",
		"--remote_cli_url=" + rpcURL}

	args = append(base, args...)
	cmd := exec.Command("kubectl", args...)
	r := &CommandResult{}

	out, err := cmd.Output()
	if err != nil {
		fmt.Printf("VALIDATOR DEBUG CMD OUTPUT %s %+v", out, err)
		r.Stderr = err.Error()
		r.Err = err
		v.result = r
		return r, err
	}
	r.Stdout = string(out)
	v.result = r
	return r, nil
}

// KubeClient saves a reference to a command
type KubeClient struct {
	result *CommandResult // stores the result of the last command that was run.
}

// RunCommand runs a command on a KubeClient.
func (k *KubeClient) RunCommand(args ...string) (*CommandResult, error) {
	base := []string{"exec", "-it", "deploy/pocket-v1-cli-client", "--container", "pocket", "--", "client"}
	args = append(base, args...)
	cmd := exec.Command("kubectl", args...)
	r := &CommandResult{}
	out, err := cmd.Output()
	if err != nil {
		r.Stderr = err.Error()
		r.Err = err
		k.result = r
		return r, err
	}
	r.Stdout = string(out)
	k.result = r
	return r, nil
}
