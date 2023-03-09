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

// NB: This accesses the kubernetes cluster-manager binary for client, not your local machine's binary.
const cliPath = "/usr/local/bin/client"

var (
	rpcURL string
	logger = pocketLogger.Global.CreateLoggerForModule("cluster-manager")
)

func init() {
	rpcURL = fmt.Sprintf("http://%s:%s", runtime.GetEnv("RPC_HOST", "v1-validator001"), defaults.DefaultRPCPort)
}

// Here's what I think this all looks like at a high level.
//
// TODO: wrap the Observer pattern around the Watcher here to expose a generic interface.
//
// TODO: bring in the Pocket Client interface and wire that up to the shell exec commands that we see below.
//
// TODO: wrap Gherkin tests around the observer assertions like they do with the staking and unstaking below.
//
// From a high level, what if we wrap this in a function and expose generic hooks into this function to assert on
// Kubernetes level events from outside itself?
//
// We should extend the Watcher pattern to the layer above and make tests only interact through observed events.
//
// The cluster can be programmatically booted up or down by adjusting the file configuration but that's out of scope.
// For now, we only want to _assume_ that the network is up and running, but reliably be able to wait on its status.
//
// By hooking into clientset and watcher in this function, we achieve both and can inherit the world for testing.
// * cracks knuckles * Looks like our work is cut out for us, boys.
//
// This file is run by tilt every time any of it changes. So we can work closely in this file with the development
// server running and get our tests in a tight feedback loop right out of the gate on this one.
func main() {
	// get a config for in-cluster services we want to control
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// feed that cluster config to kubernetes
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// gets a map of validator keys
	validatorKeysMap, err := pocketk8s.FetchValidatorPrivateKeys(clientset)
	if err != nil {
		panic(err)
	}

	// watcher spits out events, of what?
	watcher, err := clientset.CoreV1().Services("default").Watch(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	// ranges over watch events so that we get hot-reload for free?
	for event := range watcher.ResultChan() {
		logger.Info().Msgf("/watcher/ResultChan/event: %+v", event)

		// type assert a k8s service out of that watch event
		service, ok := event.Object.(*k8s.Service)
		if !ok {
			continue
		}

		// if it's not a validator node then we can skip the rest of this process.
		if !isValidator(service) {
			continue
		}

		validatorId := extractValidatorId(service.Name)
		privateKey := getPrivateKey(validatorKeysMap, validatorId)

		// TODO: add switch cases for joining, unjoining, jailing, unjailing, etc...
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
