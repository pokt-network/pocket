package e2e

import (
	"fmt"
	"io/ioutil"
	"log"
	"testing"

	"github.com/cucumber/godog"
	"github.com/testcontainers/testcontainers-go"
	"golang.org/x/net/context"

	"github.com/pokt-network/pocket/e2e/tests/runner"
)

var (
	// commander is a reference to the container these tests start and issue commands to.
	commander runner.PocketClient
)

func thePocketClientShouldHaveExitedWithoutError() error {
	return godog.ErrPending
}

// debugContainer makes the debug conatiner implement the PocketClient command interface.
type debugContainer struct {
	runner.PocketClient

	c testcontainers.Container
}

func (d *debugContainer) RunCommand(args ...string) (*runner.CommandResult, error) {
	_, reader, err := d.c.Exec(context.Background(), []string{"echo 'hello world'"})
	if err != nil {
		return nil, err
	}
	result, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	log.Printf("result: %+v", result)
	return &runner.CommandResult{
		Stdout: string(result),
		Stderr: err.Error(),
		Err:    nil,
	}, nil
}

func newDebugContainer() (*debugContainer, error) {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image: "pocket/client",
		// WaitingFor: wait.ForAll(wait.ForLog("rainTreeNetwork")), // TODO: do we need to wait at all?
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	if err := container.Start(ctx); err != nil {
		return nil, err
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}
	ip, err := container.ContainerIP(ctx)
	if err != nil {
		return nil, err
	}
	log.Printf("started debug container with %s %s", host, ip)

	return &debugContainer{c: container}, nil
}

func theUserHasAPocketClient() error {
	c, err := newDebugContainer()
	if err != nil {
		return fmt.Errorf("failed to get debug container: %+v", err)
	}
	commander = c
	return nil
}

func theUserRunsTheCommand(arg1 string) error {
	result, err := commander.RunCommand(arg1)
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
