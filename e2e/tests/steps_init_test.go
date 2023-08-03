//go:build e2e

package e2e

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	pocketLogger "github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/runtime/defaults"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	pocketk8s "github.com/pokt-network/pocket/shared/k8s"
	"github.com/regen-network/gocuke"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var e2eLogger = pocketLogger.Global.CreateLoggerForModule("e2e")

const (
	// Each actor is represented e.g. validator-001-pocket:42069 thru validator-999-pocket:42069.
	// Defines the host & port scheme that LocalNet uses for naming actors.
	validatorServiceURLTemplate = "validator-%s-pocket:%d"
	// Mapping from validators to suffix IDs as convienece for some of the tests
	validatorA = "001"
	validatorB = "002"
	// Placeholder chainID
	chainId = "0001"
)

type rootSuite struct {
	gocuke.TestingT

	// validatorKeys is hydrated by the clientset with credentials for all validators.
	// validatorKeys maps validator IDs to their private key as a hex string.
	validatorKeys map[string]string
	// clientset is the kubernetes API we acquire from the user's $HOME/.kube/config
	clientset *kubernetes.Clientset
	// node holds command results between runs and reports errors to the test suite
	node *nodePod
}

func (s *rootSuite) Before() {
	clientSet, err := getClientset(s)
	require.NoErrorf(s, err, "failed to get clientset")

	validatorKeyMap, err := pocketk8s.FetchValidatorPrivateKeys(clientSet)
	if err != nil {
		e2eLogger.Fatal().Err(err).Msg("failed to get validator key map")
	}

	s.node = new(nodePod)
	s.clientset = clientSet
	s.validatorKeys = validatorKeyMap
}

// TestFeatures runs the e2e tests specified in any .features files in this directory
// * This test suite assumes that a LocalNet is running that can be accessed by `kubectl`
func TestFeatures(t *testing.T) {
	e2eTestTags := os.Getenv("POCKET_E2E_TEST_TAGS")
	gocuke.NewRunner(t, &rootSuite{}).Path("*.feature").Tags(e2eTestTags).Run()
}

// InitializeScenario registers step regexes to function handlers

func (s *rootSuite) TheUserHasANode() {
	res, err := s.node.RunCommand("help")
	require.NoErrorf(s, err, res.Stderr)
	s.node.result = res
}

func (s *rootSuite) TheNodeShouldHaveExitedWithoutError() {
	require.NoError(s, s.node.result.Err)
}

func (s *rootSuite) TheUserRunsTheCommand(cmd string) {
	cmds := strings.Split(cmd, " ")
	res, err := s.node.RunCommand(cmds...)
	require.NoError(s, err)
	s.node.result = res
}

// TheDeveloperRunsTheCommand is similar to TheUserRunsTheCommand but exclusive to `Debug` commands
func (s *rootSuite) TheDeveloperRunsTheCommand(cmd string) {
	cmds := strings.Split(cmd, " ")
	cmds = append([]string{"Debug"}, cmds...)
	res, err := s.node.RunCommand(cmds...)
	require.NoError(s, err, fmt.Sprintf("failed to run command: '%s' due to error: %s", cmd, err))
	s.node.result = res
	e2eLogger.Debug().Msgf("TheDeveloperRunsTheCommand: '%s' with result: %s", cmd, res.Stdout)

	// Special case for managing LocalNet config when scaling actors
	if cmds[1] == "ScaleActor" {
		s.syncLocalNetConfigFromHostToLocalFS()
	}
}

func (s *rootSuite) TheNetworkIsAtGenesis() {
	s.TheDeveloperRunsTheCommand("ResetToGenesis")
}

func (s *rootSuite) TheDeveloperWaitsForMilliseconds(millis int64) {
	time.Sleep(time.Duration(millis) * time.Millisecond)
}

func (s *rootSuite) TheNetworkHasActorsOfType(num int64, actor string) {
	// normalize actor to Title case and plural
	caser := cases.Title(language.AmericanEnglish)
	actor = caser.String(strings.ToLower(actor))
	if len(actor) > 0 && actor[len(actor)-1] != 's' {
		actor += "s"
	}
	args := []string{
		"Query",
		actor,
	}

	// Depending on the type of `actor` we're querying, we'll have a different set of expected responses
	// so not all of these fields will be populated, but at least one will be.
	type expectedResponse struct {
		NumValidators *int64 `json:"total_validators"`
		NumApps       *int64 `json:"total_apps"`
		NumFishermen  *int64 `json:"total_fishermen"`
		NumServicers  *int64 `json:"total_servicers"`
		NumAccounts   *int64 `json:"total_accounts"`
	}
	validate := func(res *expectedResponse) bool {
		return res != nil && ((res.NumValidators != nil && *res.NumValidators > 0) ||
			(res.NumApps != nil && *res.NumApps > 0) ||
			(res.NumFishermen != nil && *res.NumFishermen > 0) ||
			(res.NumServicers != nil && *res.NumServicers > 0) ||
			(res.NumAccounts != nil && *res.NumAccounts > 0))
	}

	resRaw, err := s.node.RunCommand(args...)
	require.NoError(s, err)

	res := getResponseFromStdout[expectedResponse](s, resRaw.Stdout, validate)
	require.NotNil(s, res)

	// Validate that at least one of the fields that is populated has the right number of actors
	if res.NumValidators != nil {
		require.Equal(s, num, *res.NumValidators)
	} else if res.NumApps != nil {
		require.Equal(s, num, *res.NumApps)
	} else if res.NumFishermen != nil {
		require.Equal(s, num, *res.NumFishermen)
	} else if res.NumServicers != nil {
		require.Equal(s, num, *res.NumServicers)
	} else if res.NumAccounts != nil {
		require.Equal(s, num, *res.NumAccounts)
	}
}

func (s *rootSuite) ShouldBeUnreachable(pod string) {
	validate := func(res string) bool {
		return strings.Contains(res, "Unable to connect to the RPC")
	}
	args := []string{
		"Query",
		"Height",
	}
	rpcURL := fmt.Sprintf("http://%s-pocket:%s", pod, defaults.DefaultRPCPort)
	resRaw, err := s.node.RunCommandOnHost(rpcURL, args...)
	require.NoError(s, err)

	res := getStrFromStdout(s, resRaw.Stdout, validate)
	require.NotNil(s, res)

	require.Equal(s, fmt.Sprintf("❌ Unable to connect to the RPC @ \x1b[1mhttp://%s-pocket:%s\x1b[0m", pod, defaults.DefaultRPCPort), *res)
}

func (s *rootSuite) ShouldBeAtHeight(pod string, height int64) {
	args := []string{
		"Query",
		"Height",
	}
	type expectedResponse struct {
		Height *int64 `json:"Height"`
	}
	validate := func(res *expectedResponse) bool {
		return res != nil && res.Height != nil
	}

	rpcURL := fmt.Sprintf("http://%s-pocket:%s", pod, defaults.DefaultRPCPort)
	resRaw, err := s.node.RunCommandOnHost(rpcURL, args...)
	require.NoError(s, err)

	res := getResponseFromStdout[expectedResponse](s, resRaw.Stdout, validate)
	require.NotNil(s, res)

	require.Equal(s, height, *res.Height)
}

func (s *rootSuite) TheUserShouldBeAbleToSeeStandardOutputContaining(arg1 string) {
	require.Contains(s, s.node.result.Stdout, arg1)
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
	res, err := s.node.RunCommand(args...)
	require.NoError(s, err)

	s.node.result = res
}

// stakeValidator runs Validator stake command with the address, amount, chains..., and serviceURL provided
func (s *rootSuite) stakeValidator(privKey cryptoPocket.PrivateKey, amount string) {
	validatorServiceUrl := fmt.Sprintf(validatorServiceURLTemplate, validatorA, defaults.DefaultP2PPort)
	args := []string{
		"Validator",
		"Stake",
		privKey.Address().String(),
		amount,
		chainId,
		validatorServiceUrl,
	}
	res, err := s.node.RunCommand(args...)
	require.NoError(s, err)

	s.node.result = res
}

// unstakeValidator unstakes the Validator at the same address that stakeValidator uses
func (s *rootSuite) unstakeValidator() {
	privKey := s.getPrivateKey(validatorA)
	args := []string{
		"Validator",
		"Unstake",
		privKey.Address().String(),
	}
	res, err := s.node.RunCommand(args...)
	require.NoError(s, err)

	s.node.result = res
}

// getPrivateKey generates a new keypair from the private hex key that we get from the clientset
func (s *rootSuite) getPrivateKey(validatorId string) cryptoPocket.PrivateKey {
	privHexString := s.validatorKeys[validatorId]
	privateKey, err := cryptoPocket.NewPrivateKey(privHexString)
	require.NoErrorf(s, err, "failed to extract privkey")

	return privateKey
}

// getClientset uses the default path `$HOME/.kube/config` to build a kubeconfig
// and then connects to that cluster and returns a *Clientset or an error
func getClientset(t gocuke.TestingT) (*kubernetes.Clientset, error) {
	t.Helper()

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

// getResponseFromStdout returns the first output from stdout that passes the validate function provided.
// For example, when running `p1 Query Height`, the output is:
//
//	{"level":"info","module":"e2e","time":"2023-07-11T15:46:07-07:00","message":"..."}
//	{"height":3}
//
// And will return the following map so it can be used by the caller:
//
//	map[height:3]
func getResponseFromStdout[T any](t gocuke.TestingT, stdout string, validate func(res *T) bool) *T {
	t.Helper()

	for _, s := range strings.Split(stdout, "\n") {
		var m T
		if err := json.Unmarshal([]byte(s), &m); err != nil {
			continue
		}
		if !validate(&m) {
			continue
		}
		return &m
	}
	return nil
}

func getStrFromStdout(t gocuke.TestingT, stdout string, validate func(res string) bool) *string {
	t.Helper()
	for _, s := range strings.Split(stdout, "\n") {
		if !validate(s) {
			continue
		}
		return &s
	}
	return nil
}
