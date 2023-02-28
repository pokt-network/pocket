//
// uncomment later //go:build e2e

package e2e

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/cucumber/godog"
)

var (
	client    = &KubeClient{}
	validator = &Validator{}
)

func thePocketClientShouldHaveExitedWithoutError() error {
	return client.result.Err
}

func theUserHasAValidator() error {
	validator, err := validator.RunCommand("help")
	if err != nil {
		log.Printf("validator error: %+v", err)
		return err
	}
	log.Printf("VALIDATOR %s", validator)
	return err
}

func theValidatorShouldHaveExitedWithoutError() error {
	return validator.result.Err
}

func theUserHasAPocketClient() error {
	_, err := client.RunCommand("help")
	return err
}

func theUserRunsTheValidatorCommand(cmd string) error {
	// NB: Handle the split manually because Cucumber
	// doesn't support doing array of strings in test arguments.
	// See https://github.com/cucumber/cucumber-expressions#readme
	cmds := strings.Split(cmd, " ")
	result, err := validator.RunCommand(cmds...)
	if err != nil {
		fmt.Printf("ERROR: %+v", err)
		return err
	}
	if result.Err != nil {
		fmt.Printf("RES ERROR: %+v", result.Err)
		return result.Err
	}
	fmt.Println(result)
	return nil
}

func theUserRunsTheCommand(cmd string) error {
	result, err := client.RunCommand(cmd)
	if err != nil {
		// fmt.Printf("ERROR: %+v", err)
		return err
	}
	if result.Err != nil {
		// fmt.Printf("RES ERROR: %+v", result.Err)
		return result.Err
	}
	fmt.Println(result)
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
	ctx.Step(`^the user has a validator$`, theUserHasAValidator)
	ctx.Step(`^the validator should have exited without error$`, theValidatorShouldHaveExitedWithoutError)
	ctx.Step(`^the user runs the validator command "([^"]*)"$`, theUserRunsTheValidatorCommand)
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
