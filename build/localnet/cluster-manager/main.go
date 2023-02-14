package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/runtime"
	"github.com/pokt-network/pocket/runtime/defaults"
	"github.com/pokt-network/pocket/shared/crypto"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	validatorKeysMap = make(map[string]crypto.PrivateKey)
	rpcUrl           string
	log              = logger.Global.CreateLoggerForModule("cluster-manager")
)

func init() {
	rpcUrl = fmt.Sprintf("http://%s:%s", runtime.GetEnv("RPC_HOST", "v1-validator001"), defaults.DefaultRPCPort)
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

	fetchValidatorPrivateKeys(clientset, validatorKeysMap)

	watcher, err := clientset.CoreV1().Services("default").Watch(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	for event := range watcher.ResultChan() {
		service, ok := event.Object.(*v1.Service)
		if !ok {
			continue
		}

		if !isValidator(service) {
			continue
		}

		validatorId := extractValidatorId(service.Name)
		switch event.Type {
		case watch.Added:
			log.Info().Str("validator", service.Name).Msg("Validator added to the cluster")
			if err := stakeValidator(validatorKeysMap[validatorId], "150000000001", []string{"0001"}, fmt.Sprintf("v1-validator%s:8080", validatorId)); err != nil {
				log.Err(err).Msg("Error staking validator")
			}
		case watch.Deleted:
			log.Info().Str("validator", service.Name).Msg("Validator deleted from the cluster")
			if err := unstakeValidator(validatorKeysMap[validatorId]); err != nil {
				log.Err(err).Msg("Error unstaking validator")
			}
		}
	}
}

func stakeValidator(pk crypto.PrivateKey, amount string, chains []string, serviceURL string) error {
	log.Info().Str("address", pk.Address().String()).Msg("Staking Validator")
	if err := os.WriteFile("./pk.json", []byte("\""+pk.String()+"\""), 0o600); err != nil {
		return err
	}

	//nolint:gosec // G204 Dogfooding CLI
	out, err := exec.Command("/usr/local/bin/client", "--not_interactive=true", "--remote_cli_url="+rpcUrl, "Validator", "Stake", pk.Address().String(), amount, strings.Join(chains, ","), serviceURL).CombinedOutput()
	log.Info().Str("output", string(out)).Msg("CLI")
	if err != nil {
		return err
	}
	return nil
}

func unstakeValidator(pk crypto.PrivateKey) error {
	log.Info().Str("address", pk.Address().String()).Msg("Unstaking Validator")
	if err := os.WriteFile("./pk.json", []byte("\""+pk.String()+"\""), 0o600); err != nil {
		return err
	}

	//nolint:gosec // G204 Dogfooding CLI
	out, err := exec.Command("/usr/local/bin/client", "--not_interactive=true", "--remote_cli_url="+rpcUrl, "Validator", "Unstake", pk.Address().String()).CombinedOutput()
	log.Info().Str("output", string(out)).Msg("CLI")
	if err != nil {
		return err
	}
	return nil
}
