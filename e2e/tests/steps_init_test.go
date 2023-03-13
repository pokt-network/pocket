package e2e

import (
	"fmt"
	"log"
	"os/exec"
	"testing"

	"github.com/cucumber/godog"

	"github.com/pokt-network/pocket/e2e/tests/runner"
)

var (
	client runner.PocketClient
)

// KubeClient saves a reference to a command
type KubeClient struct {
	cmd *exec.Cmd
}

// RunCommand runs a command on a KubeClient.
func (k *KubeClient) RunCommand(args ...string) (*runner.CommandResult, error) {
	// TODO: wire this up to run arbitrary commands wrapped around the kubectl now that we have the right pod
	// fmt.Printf("k.cmd: %v\n", k.cmd)
	return nil, godog.ErrPending
}

func thePocketClientShouldHaveExitedWithoutError() error {
	return godog.ErrPending
}

func theUserHasAPocketClient() error {
	cmd := exec.Command("kubectl",
		"exec", "-it", "deploy/pocket-v1-cli-client",
		"--container", "pocket",
		"--", "client",
		"help")
	// TODO: for some reason the cluster-manager does not want to play nice for this command?
	// cmd := exec.Command("kubectl", "exec", "-i", "-t", "deploy/pocket-v1-cluster-manager", "--", "/usr/local/bin/cluster-manager")
	// cmd.Wait()
	log.Println(cmd)
	out, err := cmd.Output()
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return err
	}
	fmt.Printf("### OUTPUT: %s\n", out)
	client = &KubeClient{cmd: cmd}
	return nil
}

func theUserRunsTheCommand(arg1 string) error {
	result, err := client.RunCommand(arg1)
	if err != nil {
		return err
	}
	if result.Err != nil {
		return result.Err
	}
	if result.Stdout != "" {
		log.Println(result.Stdout)
	}
	return nil
}

func theUserShouldBeAbleToSeeStandardOutputContaining(arg1 string) error {
	return godog.ErrPending
}

// InitializeScenario registers step regexes to function handlers
func InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Step(`^the pocket client should have exited without error$`, thePocketClientShouldHaveExitedWithoutError)
	ctx.Step(`^the user has a pocket client$`, theUserHasAPocketClient)
	ctx.Step(`^the user runs the command "([^"]*)"$`, theUserRunsTheCommand)
	ctx.Step(`^the user should be able to see standard output containing "([^"]*)"$`, theUserShouldBeAbleToSeeStandardOutputContaining)
}

// TestFeatures runs the e2e tests specifiedin any .features files in this directory.
// * This test suite assumes that you have a local network running.
// * loops over networkConfigs and runs the entire cucumebr suite against that network instance.
// * allows support for multiple seed network configurations in the future.
func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"./"},
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}
