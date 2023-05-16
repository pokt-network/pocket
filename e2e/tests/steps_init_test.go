//go:build e2e

package e2e

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/regen-network/gocuke"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	pocketLogger "github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/runtime/defaults"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	pocketk8s "github.com/pokt-network/pocket/shared/k8s"
)

var e2eLogger = pocketLogger.Global.CreateLoggerForModule("e2e")

const (
	// defines the host & port scheme that LocalNet uses for naming validators.
	// e.g. validator-001 thru validator-999
	validatorServiceURLTmpl = "validator-%s-pocket:%d"
	// validatorA maps to suffix ID 001 and is also used by the cluster-manager
	// though it has no special permissions.
	validatorA = "001"
	// validatorB maps to suffix ID 002 and receives in the Send test.
	validatorB = "002"
	chainId    = "0001"
)

type rootSuite struct {
	gocuke.TestingT

	// validatorKeys is hydrated by the clientset with credentials for all validators.
	// validatorKeys maps validator IDs to their private key as a hex string.
	validatorKeys map[string]string
	// clientset is the kubernetes API we acquire from the user's $HOME/.kube/config
	clientset *kubernetes.Clientset
	// validator holds command results between runs and reports errors to the test suite
	validator *validatorPod
	// validatorA maps to suffix ID 001 of the kube pod that we use as our control agent
}

func (s *rootSuite) Before() {
	clientSet, err := getClientset(s)
	require.NoErrorf(s, err, "failed to get clientset")

	vkmap, err := pocketk8s.FetchValidatorPrivateKeys(clientSet)
	if err != nil {
		e2eLogger.Fatal().Err(err).Msg("failed to get validator key map")
	}

	s.validator = new(validatorPod)
	s.clientset = clientSet
	s.validatorKeys = vkmap
}

// TestFeatures runs the e2e tests specified in any .features files in this directory
// * This test suite assumes that a LocalNet is running that can be accessed by `kubectl`
func TestFeatures(t *testing.T) {
	gocuke.NewRunner(t, &rootSuite{}).Path("*.feature").Run()
}

// InitializeScenario registers step regexes to function handlers

func (s *rootSuite) TheUserHasAValidator() {
	res, err := s.validator.RunCommand("help")
	require.NoError(s, err)
	s.validator.result = res
}

func (s *rootSuite) TheValidatorShouldHaveExitedWithoutError() {
	require.NoError(s, s.validator.result.Err)
}

func (s *rootSuite) TheUserRunsTheCommand(cmd string) {
	cmds := strings.Split(cmd, " ")
	res, err := s.validator.RunCommand(cmds...)
	require.NoError(s, err)
	s.validator.result = res
}

func (s *rootSuite) TheUserShouldBeAbleToSeeStandardOutputContaining(arg1 string) {
	require.Contains(s, s.validator.result.Stdout, arg1)
}

func (s *rootSuite) TheUserStakesTheirValidatorWithAmountUpokt(amount int64) {
	privKey := s.getPrivateKey(validatorA)
	s.stakeValidator(privKey, fmt.Sprintf("%d", amount))
}

func (s *rootSuite) TheUserShouldBeAbleToUnstakeTheirValidator() {
	s.unstakeValidator()
}

// sends amount from validator-001 to validator-002
func (s *rootSuite) TheUserSendsUpoktToAnotherAddress(amount int64) {
	privKey := s.getPrivateKey(validatorA)
	valB := s.getPrivateKey(validatorB)
	args := []string{
		"--non_interactive=true",
		"--remote_cli_url=" + rpcURL,
		"Account",
		"Send",
		privKey.Address().String(),
		valB.Address().String(),
		fmt.Sprintf("%d", amount),
	}
	res, err := s.validator.RunCommand(args...)
	require.NoError(s, err)

	s.validator.result = res
}

// stakeValidator runs Validator stake command with the address, amount, chains..., and serviceURL provided
func (s *rootSuite) stakeValidator(privKey cryptoPocket.PrivateKey, amount string) {
	validatorServiceUrl := fmt.Sprintf(validatorServiceURLTmpl, validatorA, defaults.DefaultP2PPort)
	args := []string{
		"--non_interactive=true",
		"--remote_cli_url=" + rpcURL,
		"Validator",
		"Stake",
		privKey.Address().String(),
		amount,
		chainId,
		validatorServiceUrl,
	}
	res, err := s.validator.RunCommand(args...)
	require.NoError(s, err)

	s.validator.result = res
}

// unstakeValidator unstakes the Validator at the same address that stakeValidator uses
func (s *rootSuite) unstakeValidator() {
	privKey := s.getPrivateKey(validatorA)
	args := []string{
		"--non_interactive=true",
		"--remote_cli_url=" + rpcURL,
		"Validator",
		"Unstake",
		privKey.Address().String(),
	}
	res, err := s.validator.RunCommand(args...)
	require.NoError(s, err)

	s.validator.result = res
}

// getPrivateKey generates a new keypair from the private hex key that we get from the clientset
func (s *rootSuite) getPrivateKey(
	validatorId string,
) cryptoPocket.PrivateKey {
	privHexString := s.validatorKeys[validatorId]
	privateKey, err := cryptoPocket.NewPrivateKey(privHexString)
	require.NoErrorf(s, err, "failed to extract privkey")

	return privateKey
}

// getClientset uses the default path `$HOME/.kube/config` to build a kubeconfig
// and then connects to that cluster and returns a *Clientset or an error
func getClientset(t gocuke.TestingT) (*kubernetes.Clientset, error) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home dir: %w", err)
	}
	kubeConfigPath := filepath.Join(userHomeDir, ".kube", "config")
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		e2eLogger.Info().Msgf("no default kubeconfig at %s; attempting to load InClusterConfig", kubeConfigPath)
		config := inClusterConfig(t)
		clientset, err := kubernetes.NewForConfig(config)
		require.NoErrorf(t, err, "failed to get clientSet from config")

		return clientset, nil
	}

	e2eLogger.Info().Msgf("e2e tests loaded default kubeconfig located at %s", kubeConfigPath)
	clientSet, err := kubernetes.NewForConfig(kubeConfig)
	require.NoErrorf(t, err, "failed to get clientSet from config")

	return clientSet, nil
}

func inClusterConfig(t gocuke.TestingT) *rest.Config {
	config, err := rest.InClusterConfig()
	require.NoError(t, err)

	return config
}
