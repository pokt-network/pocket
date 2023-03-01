//go:build e2e

package e2e

import (
	"fmt"
	"strings"
	"testing"

	"github.com/cucumber/godog"
	"github.com/testcontainers/testcontainers-go"
)

var (
	client = &KubeClient{}
)

func thePocketClientShouldHaveExitedWithoutError() error {
	return client.result.Err
}

func theUserHasAPocketClient() error {
	_, err := client.RunCommand("help")
	return err
}

func theUserRunsTheCommand(arg1 string) error {
	result, err := client.RunCommand(arg1)
	if err != nil {
		return err
	}
	if result.Err != nil {
		return result.Err
	}
	return nil
}

func theUserShouldBeAbleToSeeStandardOutputContaining(arg1 string) error {
	if !strings.Contains(client.result.Stdout, arg1) {
		return fmt.Errorf("stdout must contain %s", arg1)
	}
	return nil
}

// InitializeScenario registers step regexes to function handlers
func InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Step(`^the pocket client should have exited without error$`, thePocketClientShouldHaveExitedWithoutError)
	ctx.Step(`^the user has a pocket client$`, theUserHasAPocketClient)
	ctx.Step(`^the user runs the command "([^"]*)"$`, theUserRunsTheCommand)
	ctx.Step(`^the user should be able to see standard output containing "([^"]*)"$`, theUserShouldBeAbleToSeeStandardOutputContaining)
}

// TestFeatures runs the e2e tests specifiedin any .features files in this directory.
// * This test suite assumes that a LocalNet is running that can be accessed by `kubectl`.
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

// buildDockerTestContainer builds a pocket node test container
func buildDockerTestContainer(ctx context.Context) (testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    "../../",
			Dockerfile: "../../build/Dockerfile.m1.proto",
		},
	}
	node, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get container node: %w", err)
	}

	// node can accept commands
	// code, result, err := node.Exec(ctx, "pocket help", nil)

	return node, nil
}

// What if CreateNetwork just created a docker-compose start command and used that?
func createNetwork(ctx context.Context, conf networkConfig) (*Network, error) {
	network := &Network{
		nodes: make([]testcontainers.Container, 0),
	}

	// TODO: start docker containers according to networkConfig and add it to our Network
	node, err := buildDockerTestContainer(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build test container")
	}
	network.nodes = append(network.nodes, node)
	return network, nil
}

// Cleanup cleans up all Network nodes.
func (n *Network) Cleanup(ctx context.Context) error {
	for _, node := range n.nodes {
		// terminate all nodes and report errors
		err := node.Terminate(ctx)
		if err != nil {
			log.Printf("cleanup failed %v", err) // TODO: what to do with this error on cleanup?
		}
	}
	// TODO: collect errors maybe
	return nil
}
