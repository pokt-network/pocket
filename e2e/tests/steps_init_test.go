package e2e

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/cucumber/godog"
	"github.com/testcontainers/testcontainers-go"
)

func thePocketClientShouldHaveExitedWithoutError() error {
	return godog.ErrPending
}

func theUserHasAPocketClient() error {
	return godog.ErrPending
}

func theUserRunsTheCommand(arg1 string) error {
	return godog.ErrPending
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

// Network is a collection of docker container pocket nodes
type Network struct {
	nodes []testcontainers.Container
}

type networkConfig struct{}

// TestFeatures runs the e2e tests specifiedin any .features files in this directory.
// It loops over networkConfigs and runs the entire cucumebr suite against it.
// This allows support for multiple seed network configurations in the future without
// having to worry about it right now.
func TestFeatures(t *testing.T) {
	// TODO: seed with a basic network config. // REFERENCE: /build/config/ has the files that we should need here
	networkConfigs := []networkConfig{
		{},
	}
	for _, v := range networkConfigs {
		// make a new context for each network
		ctx := context.Background()

		// make a new network with the current config
		network, err := createNetwork(ctx, v)
		if err != nil {
			t.Errorf("failed to start create test network %+v", err)
		}

		// we have network and suite here now, run the cucumber test suite.
		suite := godog.TestSuite{
			ScenarioInitializer: InitializeScenario,
			Options: &godog.Options{
				Format:   "pretty",
				Paths:    []string{"./"},
				TestingT: t,
			},
		}

		t.Logf("network: %v\n", network)
		t.Logf("suite: %v\n", suite)

		if suite.Run() != 0 {
			t.Fatal("non-zero status returned, failed to run feature tests")
		}

		defer func(ctx context.Context) {
			network.Cleanup(ctx)
		}(ctx)
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
