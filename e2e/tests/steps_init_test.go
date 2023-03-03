package e2e

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/cucumber/godog"
	"github.com/pokt-network/pocket/e2e/tests/runner"
	"github.com/stretchr/testify/assert"

	tc "github.com/testcontainers/testcontainers-go/modules/compose"
)

var (
	// TODO: how could we not use global state here?
	// we could consider something like badger?
	realm *Realm
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
	// nodes holds the list of booted containers in this realm. They fulfill
	// the PocketClient interface and the Realm may only interact with them
	// through that interface.
	nodes []runner.PocketClient
	// suite is the test suite for this realm
	suite godog.TestSuite
}

type networkConfig struct{} // a placeholder for networkConfig info

// TestFeatures runs the e2e tests specifiedin any .features files in this directory.
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

	realm = &Realm{
		nodes: make([]runner.PocketClient, 0),
		suite: suite,
	}

	// TODO: initialize containers for testing here
	// - docker-compose
	// - individual containers manually wired together
	// TODO: wrap testcontainers.Container with dockerClient to fulfill RunCommand / PocketClinet interface

	compose, err := tc.NewDockerCompose("../../build/deployments/docker-compose.yaml")
	assert.NoError(t, err, "NewDockerComposeAPI()")

	t.Cleanup(func() {
		err := compose.Down(context.Background(), tc.RemoveOrphans(true), tc.RemoveImagesLocal)
		if err != nil {
			t.Errorf("failed to tear down compose instance: %+v", err)
		}
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	// Okay, so we get to here - progress.
	assert.NoError(t, compose.Up(ctx, tc.Wait(true)), "compose.Up()")

	// do some testing here

	for _, n := range realm.nodes {
		log.Printf("node: %+v", n)
		result, err := n.RunCommand("echo 'hello world'")
		if err != nil {
			log.Fatalf("RunCommand() failed: %+v", err)
		}
		fmt.Printf("result: %v\n", result)
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}
