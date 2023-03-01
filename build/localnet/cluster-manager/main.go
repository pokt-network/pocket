package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	pocketLogger "github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/runtime"
	"github.com/pokt-network/pocket/runtime/defaults"
	"github.com/pokt-network/pocket/shared/crypto"
	pocketk8s "github.com/pokt-network/pocket/shared/k8s"
	k8s "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const cliPath = "/usr/local/bin/client"

var (
	rpcURL string
	logger = pocketLogger.Global.CreateLoggerForModule("cluster-manager")
)

func init() {
	rpcURL = fmt.Sprintf("http://%s:%s", runtime.GetEnv("RPC_HOST", "v1-validator001"), defaults.DefaultRPCPort)
}

func main() {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	validatorKeysMap, err := pocketk8s.FetchValidatorPrivateKeys(clientset)
	if err != nil {
		panic(err)
	}

	watcher, err := clientset.CoreV1().Services("default").Watch(context.TODO(), metav1.ListOptions{})
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
			// TODO: consolidate args into constants
			validatorServiceUrl := fmt.Sprintf("v1-validator%s:%d", validatorId, defaults.DefaultP2PPort)
			if err := stakeValidator(privateKey, "150000000001", []string{"0001"}, validatorServiceUrl); err != nil {
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
