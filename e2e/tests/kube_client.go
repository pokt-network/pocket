//go:build e2e

package e2e

import (
	"os/exec"
)

type CommandResult struct {
	Stdout string
	Stderr string
	Err    error
}

var _ PocketClient = &KubeClient{}

// PocketClient is a single function interface for interacting with a node.
type PocketClient interface {
	RunCommand(...string) (*CommandResult, error)
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
