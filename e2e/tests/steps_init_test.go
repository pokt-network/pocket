//go:build e2e

package e2e

import (
	"fmt"
	"strings"
	"testing"

	"github.com/cucumber/godog"
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
