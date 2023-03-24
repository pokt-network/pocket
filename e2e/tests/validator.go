//go:build e2e

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

// rpcURL of the pod that the test harness drives
var rpcURL string

func init() {
	rpcURL = fmt.Sprintf("http://%s:%s", runtime.GetEnv("RPC_HOST", "v1-validator001"), defaults.DefaultRPCPort)
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

// getClientset uses the default path `$HOME/.kube/config` to build a kubeconfig
// and then connects to that cluster and returns a *Clientset or an error
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
	if err != nil {
		v.result = r
		return r, err
	}
	r.Stdout = string(out)
	v.result = r
	return r, nil
}
