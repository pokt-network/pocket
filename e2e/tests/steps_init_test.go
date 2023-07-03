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
	"github.com/regen-network/gocuke"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
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

	servicerA = "001"
	appA      = "000"
	serviceA  = "0001"
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

	// servicerKeys is hydrated by the clientset with credentials for all servicers.
	// servicerKeys maps servicer IDs to their private key as a hex string.
	servicerKeys map[string]string

	// appKeys is hydrated by the clientset with credentials for all apps.
	// appKeys maps app IDs to their private key as a hex string.
	appKeys map[string]string
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

	// ADDPR: use pocketk8s to populate
	s.servicerKeys = map[string]string{
		// 000 servicer NOT in session
		"000": "acbca21f295caefdfe480ceba85f3fed31a50915162f94867f9c23d8f474f4c6d1130c5eb920af8edd5b6bfa39d33aa787f421c8ba0786de4ca4e7703553bb97",
		// 001 servicer in session
		"001": "eec4072b095acf60be9d6be4093b14a24e2ddb6e9d385d980a635815961d025856915c1270bc8d9280a633e0be51647f62388a851318381614877ef2ed84a495",
	}

	s.appKeys = map[string]string{
		"000": "468cc03083d72f2440d3d08d12143b9b74cca9460690becaa2499a4f04fddaa805a25e527bf6f51676f61f2f1a96efaa748218ac82f54d3cdc55a4881389eb60",
	}
}

// TestFeatures runs the e2e tests specified in any .features files in this directory
// * This test suite assumes that a LocalNet is running that can be accessed by `kubectl`
func TestFeatures(t *testing.T) {
	runner := gocuke.NewRunner(t, &rootSuite{}).Path("*.feature")
	// DISCUSS: is there a better way to make gocuke pickup the balance, i.e. a hexadecimal, as a string in function argument?
	runner.Step(`^the\srelay\sresponse\scontains\s([[:alnum:]]+)$`, (*rootSuite).TheRelayResponseContains)
	runner.Run()
}

// InitializeScenario registers step regexes to function handlers

func (s *rootSuite) TheUserHasAValidator() {
	res, err := s.validator.RunCommand("help")
	require.NoErrorf(s, err, res.Stderr)
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

// An Application requests the account balance of a specific address at a specific height
func (s *rootSuite) TheApplicationSendsARelayToAServicer() {
	// ADDPR: Add a servicer staked for the Ethereum RelayChain
	// ADDPR: Verify the response: correct id, correct jsonrpc, and the returned balance
	// ADDPR: move the method and account to the feature file

	// ETH
	// Account: 0x8315177aB297bA92A06054cE80a67Ed4DBd7ed3a   (Arbitrum Bridge)
	// Balance: 1,160,126.46817237178258965 ETH  = 0xf5aa94f49d4fd1f8dcd2
	// BlockNumber: 17605670 = 0x10CA426
	checkBalanceRelay := `{"method": "eth_getBalance", "params": ["0x8315177aB297bA92A06054cE80a67Ed4DBd7ed3a", "0x10CA426"], "id": "1", "jsonrpc": "2.0"}`

	servicerPrivateKey := s.getServicerPrivateKey(servicerA)
	appPrivateKey := s.getAppPrivateKey(appA)

	s.sendTrustlessRelay(checkBalanceRelay, servicerPrivateKey.Address().String(), appPrivateKey.Address().String())
}

func (s *rootSuite) TheRelayResponseContains(arg1 string) {
	require.Contains(s, s.validator.result.Stdout, arg1)
}

func (s *rootSuite) sendTrustlessRelay(relayPayload string, servicerAddr, appAddr string) {
	args := []string{
		"Servicer",
		"Relay",
		appAddr,
		servicerAddr,
		// IMPROVE: add ETH_Goerli as a chain/service to genesis
		serviceA,
		relayPayload,
	}

	// DISCUSS: does this need to be run from a client, i.e. not a validator, pod?
	res, err := s.validator.RunCommand(args...)

	require.NoError(s, err)

	s.validator.result = res
}

// getAppPrivateKey generates a new keypair from the application private hex key that we get from the clientset
func (s *rootSuite) getAppPrivateKey(
	appId string,
) cryptoPocket.PrivateKey {
	privHexString := s.appKeys[appId]
	privateKey, err := cryptoPocket.NewPrivateKey(privHexString)
	require.NoErrorf(s, err, "failed to extract privkey")

	return privateKey
}

// getServicerPrivateKey generates a new keypair from the servicer private hex key that we get from the clientset
func (s *rootSuite) getServicerPrivateKey(
	servicerId string,
) cryptoPocket.PrivateKey {
	privHexString := s.servicerKeys[servicerId]
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
