package e2e

import (
	"testing"

	"github.com/cucumber/godog"
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

// TestFeatures runs the e2e tests specifiedin any .features files in this directory.
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
