//go:build e2e

package e2e

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pokt-network/pocket/runtime/defaults"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	pocketk8s "github.com/pokt-network/pocket/shared/k8s"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/cucumber/godog"
)

var (
	// validatorKeys is hydrated by the clientset with credentials for all validators
	validatorKeys map[string]string
	// clientset is the kubernetes API we acquire from the user's $HOME/.kube/config
	clientset *kubernetes.Clientset
	// validator holds command results between runs and reports errors to the test suite
	validator = &validatorPod{}
	// validatorA maps to suffix ID 001 of the kube pod that we use as our control agent
	validatorA string = "001"
	// validatorB maps to suffix ID 002
	validatorB string = "002"
	chainId           = "0001"
)

func init() {
	cs, err := getClientset()
	if err != nil {
		log.Fatalf("failed to get clientset: %v", err)
	}
	clientset = cs
	vkmap, err := pocketk8s.FetchValidatorPrivateKeys(clientset)
	if err != nil {
		log.Fatalf("failed to get validator keys: %v", err)
	}
	validatorKeys = vkmap
}

// TestFeatures runs the e2e tests specifiedin any .features files in this directory
// * This test suite assumes that a LocalNet is running that can be accessed by `kubectl`
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

// InitializeScenario registers step regexes to function handlers
func InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Step(`^the user runs the command "([^"]*)"$`, theUserRunsTheCommand)
	ctx.Step(`^the user should be able to see standard output containing "([^"]*)"$`, theUserShouldBeAbleToSeeStandardOutputContaining)
	ctx.Step(`^the user has a validator$`, theUserHasAValidator)
	ctx.Step(`^the validator should have exited without error$`, theValidatorShouldHaveExitedWithoutError)
	ctx.Step(`^the user stakes their validator with (\d+) POKT$`, theUserStakesTheirValidatorWithPOKT)
	ctx.Step(`^the user should be able to unstake their wallet$`, theUserShouldBeAbleToUnstakeTheirWallet)
	ctx.Step(`^the user sends (\d+) POKT to another address$`, theUserSendsPOKTToAnotherAddress)
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

func theUserRunsTheCommand(cmd string) error {
	cmds := strings.Split(cmd, " ")
	result, err := validator.RunCommand(cmds...)
	if err != nil {
		validator.result = result
		return err
	}
	validator.result = result
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

func theUserShouldBeAbleToUnstakeTheirWallet() error {
	return unstakeValidator()
}

// sends amount of POKT from v1-validator-001 to v1-validator-002
func theUserSendsPOKTToAnotherAddress(amount int) error {
	privateKey := getPrivateKey(validatorKeys, validatorA)
	valB := getPrivateKey(validatorKeys, validatorB)
	args := []string{
		"--non_interactive=true",
		"--remote_cli_url=" + rpcURL,
		"Account",
		"Send",
		privateKey.Address().String(),
		valB.Address().String(),
		fmt.Sprintf("%d", amount),
	}
	validator.RunCommand(args...)
	res, err := validator.RunCommand(args...)
	if err != nil {
		validator.result = res
		return err
	}
	validator.result = res
	return nil
}

// stakeValidator runs Validator stake command with the address, amount, chains..., and serviceURL provided
func stakeValidator(amount string) error {
	privateKey := getPrivateKey(validatorKeys, validatorA)
	validatorServiceUrl := fmt.Sprintf("v1-validator%s:%d", validatorA, defaults.DefaultP2PPort)
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

// unstakeValidator unstakes the Validator at the same address that stakeValidator uses
func unstakeValidator() error {
	privKey := getPrivateKey(validatorKeys, validatorA)
	args := []string{
		"--non_interactive=true",
		"--remote_cli_url=" + rpcURL,
		"Validator",
		"Unstake",
		privKey.Address().String(),
	}
	res, err := validator.RunCommand(args...)
	if err != nil {
		validator.result = res
		return err
	}
	validator.result = res
	return nil
}

// getPrivateKey generates a new keypair from the private hex key that we get from the clientset
func getPrivateKey(keyMap map[string]string, validatorId string) cryptoPocket.PrivateKey {
	privHexString := keyMap[validatorId]
	keyPair, err := cryptoPocket.CreateNewKeyFromString(privHexString, "", "")
	if err != nil {
		log.Fatalf("failed to extract keypair %+v", err)
	}
	privateKey, err := keyPair.Unarmour("")
	if err != nil {
		log.Fatalf("failed to extract keypair %+v", err)
	}
	return privateKey
}

// getClientset uses the default path `$HOME/.kube/config` to build a kubeconfig
// and then connects to that cluster and returns a *Clientset or an error
func getClientset() (*kubernetes.Clientset, error) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home dir: %w", err)
	}
	kubeConfigPath := filepath.Join(userHomeDir, ".kube", "config")
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to build kubeconfig: %w", err)
	}
	clientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get clientset from config: %w", err)
	}
	return clientset, nil
}
