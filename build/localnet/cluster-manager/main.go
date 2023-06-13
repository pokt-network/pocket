package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/pokt-network/pocket/app/client/cli/flags"
	pocketLogger "github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/runtime/defaults"
	"github.com/pokt-network/pocket/shared/crypto"
	pocketk8s "github.com/pokt-network/pocket/shared/k8s"
	k8s "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const cliPath = "/usr/local/bin/p1"
const validatorServiceUrlFormat = "validator-%s-pocket:%d"

var (
	rpcURL string
	logger = pocketLogger.Global.CreateLoggerForModule("cluster-manager")

	autoStakeAmount = "150000000001"
	autoStakeChains = []string{"0001"}
	// autoStakeSkipStakeForValidatorIds is a list of validator ids that should not be auto-staked
	// it is used to avoid auto-staking the validators that are already staked as part of genesis.
	autoStakeSkipStakeForValidatorIds = []string{"001", "002", "003", "004"}

	clusterManagerCmd = &cobra.Command{
		Use:   "cluster-manager",
		Short: "Start the Pocket Network Cluster Manager service",
		Long: `Start the Pocket Network Cluster Manager service which listens for and reacts to events coming over the k8s.io API's watch.Interface#ResultChan().

See the following k8s.io documentation for more information:
- https://pkg.go.dev/k8s.io/client-go/kubernetes/typed/core/v1@v0.26.1#ServiceInterface
- https://pkg.go.dev/k8s.io/apimachinery/pkg/watch#Interface,
- https://pkg.go.dev/k8s.io/apimachinery/pkg/watch#Event`,
		Run:  runClusterManagerCmd,
		Args: cobra.ExactArgs(0),
	}
)

func init() {
	// setup the `remote_cli_url` flag to be consistent with the CLI except for
	// the default, which is set to the URL of the first k8s validator's RPC
	// endpoint.
	clusterManagerCmd.PersistentFlags().StringVar(
		&flags.RemoteCLIURL,
		"remote_cli_url",
		defaults.Validator1EndpointK8SHostname,
		"takes a remote endpoint in the form of <protocol>://<host>:<port> (uses RPC Port)",
	)

	// ensure that the env var can override the flag
	if err := viper.BindPFlag("remote_cli_url", clusterManagerCmd.PersistentFlags().Lookup("remote_cli_url")); err != nil {
		log.Fatalf("Error binding remote_cli_url flag: %v", err)
	}
}

func main() {
	if err := clusterManagerCmd.Execute(); err != nil {
		log.Fatalf("Error executing cluster-manager command: %v", err)
	}
}

func runClusterManagerCmd(_ *cobra.Command, _ []string) {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// Monitor for crashed pods and delete them
	go initCrashedPodsDeleter(clientset)

	validatorKeysMap, err := pocketk8s.FetchValidatorPrivateKeys(clientset)
	if err != nil {
		panic(err)
	}

	watcher, err := clientset.CoreV1().Services(pocketk8s.CurrentNamespace).Watch(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	for event := range watcher.ResultChan() {
		service, ok := event.Object.(*k8s.Service)
		if !ok {
			continue
		}

		if !isValidator(service) {
			continue
		}

		validatorId := extractValidatorId(service.Name)
		privateKey := getPrivateKey(validatorKeysMap, validatorId)

		switch event.Type {
		case watch.Added:
			logger.Info().Str("validator", service.Name).Msg("Validator added to the cluster")
			if shouldSkipAutoStaking(validatorId) {
				logger.Info().Str("validator", service.Name).Msg("autoStakeSkipStakeForValidarIds includes this validatorId. Skipping auto-staking")
				continue
			}

			validatorServiceUrl := fmt.Sprintf(validatorServiceUrlFormat, validatorId, defaults.DefaultP2PPort)
			if err := stakeValidator(privateKey, autoStakeAmount, autoStakeChains, validatorServiceUrl); err != nil {
				logger.Err(err).Msg("Error staking validator")
			}
		case watch.Deleted:
			logger.Info().Str("validator", service.Name).Msg("Validator deleted from the cluster")
			if err := unstakeValidator(privateKey); err != nil {
				logger.Err(err).Msg("Error unstaking validator")
			}
		}
	}
}

func stakeValidator(pk crypto.PrivateKey, amount string, chains []string, serviceURL string) error {
	logger.Info().Str("address", pk.Address().String()).Msg("Staking Validator")
	if err := os.WriteFile("./pk.json", []byte("\""+pk.String()+"\""), 0o600); err != nil {
		return err
	}

	args := []string{
		"--non_interactive=true",
		"--remote_cli_url=" + rpcURL,
		"Validator",
		"Stake",
		pk.Address().String(),
		amount,
		strings.Join(chains, ","),
		serviceURL,
	}
	logger.Debug().Str("command", cliPath+" "+strings.Join(args, " ")).Msg("Invoking CLI")

	//nolint:gosec // G204 Dogfooding CLI
	out, err := exec.Command(cliPath, args...).CombinedOutput()
	logger.Info().Str("output", string(out)).Msg("CLI command")
	if err != nil {
		return err
	}
	return nil
}

func unstakeValidator(pk crypto.PrivateKey) error {
	logger.Info().Str("address", pk.Address().String()).Msg("Unstaking Validator")
	if err := os.WriteFile("./pk.json", []byte("\""+pk.String()+"\""), 0o600); err != nil {
		return err
	}

	args := []string{
		"--non_interactive=true",
		"--remote_cli_url=" + rpcURL,
		"Validator",
		"Unstake",
		pk.Address().String(),
	}
	logger.Debug().Str("command", cliPath+" "+strings.Join(args, " ")).Msg("Invoking CLI")

	//nolint:gosec // G204 Dogfooding CLI
	out, err := exec.Command(cliPath, args...).CombinedOutput()
	logger.Info().Str("output", string(out)).Msg("CLI command")
	if err != nil {
		return err
	}
	return nil
}
