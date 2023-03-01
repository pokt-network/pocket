package e2e

import (
	"context"
	"fmt"
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

// Realm is a collection of docker container pocket nodes
type Realm struct {
	nodes []testcontainers.Container
}

type networkConfig struct{} // a placeholder for networkConfig info

// TestFeatures runs the e2e tests specifiedin any .features files in this directory.
// * loops over networkConfigs and runs the entire cucumebr suite against that network instance.
// * allows support for multiple seed network configurations in the future.
func TestFeatures(t *testing.T) {
	// TODO: seed with a basic network config. // REFERENCE: /build/config/ has the files that we should need here
	networkConfigs := []networkConfig{{}} // NOTE: force with empty placeholder once to make test run and fail
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

func createNetwork(ctx context.Context, conf networkConfig) (*Realm, error) {
	network := &Realm{
		nodes: make([]testcontainers.Container, 0),
	}

	// TODO: create and return a docker-compose network here and add them to the Network

	return network, fmt.Errorf("not impl")
}

func (n *Realm) Cleanup(ctx context.Context) error {
	return fmt.Errorf("not impl")
}
