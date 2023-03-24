package e2e

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pokt-network/pocket/runtime"
	"github.com/pokt-network/pocket/runtime/defaults"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
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

var _ PocketClient = &Validator{}

// Validator holds the connection information for validator-001 for testing.
type Validator struct {
	result *CommandResult // stores the result of the last command that was run.
}

func getClientset() (*kubernetes.Clientset, error) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home dir: %w", err)
	}
	kubeConfigPath := filepath.Join(userHomeDir, ".kube", "config")

	kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to build kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get clientset from config: %w", err)
	}

	return clientset, nil
}

// RunCommand runs a command on the pocket binary
func (v *Validator) RunCommand(args ...string) (*CommandResult, error) {
	base := []string{
		"exec", "-i", "deploy/pocket-v1-cli-client",
		"--container", "pocket",
		"--", cliPath,
		"--non_interactive=true",
		// "--remote_cli_url=" + rpcURL,
	}

	args = append(base, args...)
	cmd := exec.Command("kubectl", args...)
	r := &CommandResult{}

	out, err := cmd.Output()
	if err != nil {
		v.result = r
		return r, err
	}
	r.Stdout = string(out)
	v.result = r
	return r, nil
}
