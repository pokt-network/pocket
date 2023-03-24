package e2e

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/pokt-network/pocket/runtime/defaults"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	pocketk8s "github.com/pokt-network/pocket/shared/k8s"

	"github.com/cucumber/godog"
)

var (
	validator = &Validator{}
)

func thePocketClientShouldHaveExitedWithoutError() error {
	return validator.result.Err
}

func theUserHasAValidator() error {
	res, err := validator.RunCommand("help")
	if err != nil {
		log.Printf("validator error: %+v", err)
		return err
	}
	validator.result = res
	return err
}

func theValidatorShouldHaveExitedWithoutError() error {
	return validator.result.Err
}

func theUserHasAPocketClient() error {
	res, err := validator.RunCommand("help")
	validator.result = res
	return err
}

func theUserRunsTheValidatorCommand(cmd string) error {
	cmds := strings.Split(cmd, " ")
	result, err := validator.RunCommand(cmds...)
	if err != nil {
		return err
	}
	if result.Err != nil {
		return result.Err
	}
	return nil
}

func theUserRunsTheCommand(cmd string) error {
	result, err := validator.RunCommand(cmd)
	if err != nil {
		validator.result = result
		return err
	}
	if result.Err != nil {
		return result.Err
	}
	return nil
}

func theUserShouldBeAbleToSeeStandardOutputContaining(arg1 string) error {
	if !strings.Contains(validator.result.Stdout, arg1) {
		return fmt.Errorf("stdout must contain %s", arg1)
	}
	return nil
}

func theUserStakesTheirValidatorWithPOKT(amount int) error {
	return stakeValidator(fmt.Sprintf("%d", amount))
}

// stakeValidator runs Validator stake command with the address, amount, chains..., and serviceURL provided.
func stakeValidator(amount string) error {
	// TODO: pull this out to an init so it only runs once
	clientset, err := getClientset()
	if err != nil {
		return fmt.Errorf("failed to get clientset: %w", err)
	}
	validatorKeysMap, err := pocketk8s.FetchValidatorPrivateKeys(clientset)
	if err != nil {
		return fmt.Errorf("failed to get validator keys: %w", err)
	}

	validatorId := "001"
	chainId := "0001"
	privateKey := getPrivateKey(validatorKeysMap, validatorId)
	validatorServiceUrl := fmt.Sprintf("v1-validator%s:%d", validatorId, defaults.DefaultP2PPort)

	args := []string{
		// NB: ignore passing a --pwd flag because
		// validator keys have empty passwords
		"--non_interactive=true",
		"--remote_cli_url=" + rpcURL,
		"Validator",
		"Stake",
		privateKey.Address().String(),
		amount,
		chainId,
		validatorServiceUrl,
	}

	res, err := validator.RunCommand(args...)
	if err != nil {
		validator.result = res
		return err
	}

	validator.result = res
	return nil
}

// TODO: Create a type for `validatorKeyMap` and document what the expected key-value types contain
func getPrivateKey(validatorKeysMap map[string]string, validatorId string) cryptoPocket.PrivateKey {
	privHexString := validatorKeysMap[validatorId]
	keyPair, err := cryptoPocket.CreateNewKeyFromString(privHexString, "", "")
	if err != nil {
		log.Fatalf("failed to extract keypair %+v", err)
	}
	privateKey, err := keyPair.Unarmour("") // empty passphrase
	if err != nil {
		log.Fatalf("failed to extract keypair %+v", err)
	}
	return privateKey
}

// func unstakeValidator(address string) error {
// 	args := []string{
// 		"--non_interactive=true",
// 		"--remote_cli_url=" + rpcURL,
// 		"Validator", "Unstake", address,
// 	}
// 	res, err := validator.RunCommand(args...)
// 	if err != nil {
// 		validator.result = res
// 		return err
// 	}
// 	validator.result = res
// 	return nil
// }

// InitializeScenario registers step regexes to function handlers
func InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Step(`^the pocket client should have exited without error$`, thePocketClientShouldHaveExitedWithoutError)
	ctx.Step(`^the user has a pocket client$`, theUserHasAPocketClient)
	ctx.Step(`^the user runs the command "([^"]*)"$`, theUserRunsTheCommand)
	ctx.Step(`^the user should be able to see standard output containing "([^"]*)"$`, theUserShouldBeAbleToSeeStandardOutputContaining)
	ctx.Step(`^the user has a validator$`, theUserHasAValidator)
	ctx.Step(`^the validator should have exited without error$`, theValidatorShouldHaveExitedWithoutError)
	ctx.Step(`^the user runs the validator command "([^"]*)"$`, theUserRunsTheValidatorCommand)
	ctx.Step(`^the user stakes their validator with (\d+) POKT$`, theUserStakesTheirValidatorWithPOKT)
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
