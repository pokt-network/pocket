//go:build e2e

package e2e

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	pocketLogger "github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/runtime/defaults"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	pocketk8s "github.com/pokt-network/pocket/shared/k8s"

	"github.com/cucumber/godog"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	logger = pocketLogger.Global.CreateLoggerForModule("e2e")

	// validatorKeys is hydrated by the clientset with credentials for all validators.
	// validatorKeys maps validator IDs to their private key as a hex string.
	validatorKeys map[string]string
	// clientset is the kubernetes API we acquire from the user's $HOME/.kube/config
	clientset *kubernetes.Clientset
	// validator holds command results between runs and reports errors to the test suite
	validator = &validatorPod{}
	// validatorA maps to suffix ID 001 of the kube pod that we use as our control agent
)

const (
	// defines the host & port scheme that LocalNet uses for naming validators.
	// e.g. v1-validator-001 thru v1-validator-999
	validatorServiceURLTmpl = "v1-validator%s:%d"
	// validatorA maps to suffix ID 001 and is also used by the cluster-manager
	// though it has no special permissions.
	validatorA = "001"
	// validatorB maps to suffix ID 002 and receives in the Send test.
	validatorB = "002"
	chainId    = "0001"
)

func init() {
	cs, err := getClientset()
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to get clientset")
	}
	clientset = cs
	vkmap, err := pocketk8s.FetchValidatorPrivateKeys(clientset)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to get validator key map")
	}
	validatorKeys = vkmap
}

// TestFeatures runs the e2e tests specified in any .features files in this directory
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
	ctx.Step(`^the user stakes their validator with amount (\d+) uPOKT$`, theUserStakesTheirValidatorWith)
	ctx.Step(`^the user should be able to unstake their validator$`, theUserShouldBeAbleToUnstakeTheirValidator)
	ctx.Step(`^the user sends (\d+) uPOKT to another address$`, theUserSendsToAnotherAddress)
}

func theUserHasAValidator() error {
	res, err := validator.RunCommand("help")
	validator.result = res
	if err != nil {
		return err
	}
	return nil
}

func theValidatorShouldHaveExitedWithoutError() error {
	return validator.result.Err
}

func theUserRunsTheCommand(cmd string) error {
	cmds := strings.Split(cmd, " ")
	res, err := validator.RunCommand(cmds...)
	validator.result = res
	if err != nil {
		return err
	}
	return nil
}

func theUserShouldBeAbleToSeeStandardOutputContaining(arg1 string) error {
	if !strings.Contains(validator.result.Stdout, arg1) {
		return fmt.Errorf("stdout must contain %s", arg1)
	}
	return nil
}

func theUserStakesTheirValidatorWith(amount int) error {
	return stakeValidator(fmt.Sprintf("%d", amount))
}

func theUserShouldBeAbleToUnstakeTheirValidator() error {
	return unstakeValidator()
}

// sends amount from v1-validator-001 to v1-validator-002
func theUserSendsToAnotherAddress(amount int) error {
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
	res, err := validator.RunCommand(args...)
	validator.result = res
	if err != nil {
		return err
	}
	return nil
}

// stakeValidator runs Validator stake command with the address, amount, chains..., and serviceURL provided
func stakeValidator(amount string) error {
	privateKey := getPrivateKey(validatorKeys, validatorA)
	validatorServiceUrl := fmt.Sprintf(validatorServiceURLTmpl, validatorA, defaults.DefaultP2PPort)
	args := []string{
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
	validator.result = res
	if err != nil {
		return err
	}
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
	validator.result = res
	if err != nil {
		return err
	}
	return nil
}

// getPrivateKey generates a new keypair from the private hex key that we get from the clientset
func getPrivateKey(keyMap map[string]string, validatorId string) cryptoPocket.PrivateKey {
	privHexString := keyMap[validatorId]
	keyPair, err := cryptoPocket.CreateNewKeyFromString(privHexString, "", "")
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to extract keypair")
	}
	privateKey, err := keyPair.Unarmour("")
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to extract privkey")
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
