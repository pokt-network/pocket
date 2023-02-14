package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

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
	rpcHost          = "http://v1-validator001:50832"
)

func init() {
	if os.Getenv("RPC_HOST") != "" {
		rpcHost = fmt.Sprintf("http://%s:%s", os.Getenv("RPC_HOST"), defaults.DefaultRPCPort)
	}
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
			log.Printf("Validator %s added to the cluster\n", service.Name)
			if err := stakeValidator(validatorKeysMap[validatorId], "150000000001", []string{"0001"}, fmt.Sprintf("v1-validator%s:8080", validatorId)); err != nil {
				log.Printf("Error staking validator: %s", err.Error())
			}
		case watch.Deleted:
			log.Printf("Validator %s deleted from the cluster\n", service.Name)
			if err := unstakeValidator(validatorKeysMap[validatorId]); err != nil {
				log.Printf("Error unstaking validator: %s", err.Error())
			}
		}
	}
}

func stakeValidator(pk crypto.PrivateKey, amount string, chains []string, serviceURL string) error {
	log.Printf("Staking Validator with Address: %s\n", pk.Address())
	if err := os.WriteFile("./pk.json", []byte("\""+pk.String()+"\""), 0o600); err != nil {
		return err
	}

	//nolint:gosec // G204 Dogfooding CLI
	out, err := exec.Command("/usr/local/bin/client", "--not_interactive=true", "--remote_cli_url="+rpcHost, "Validator", "Stake", pk.Address().String(), amount, strings.Join(chains, ","), serviceURL).CombinedOutput()
	if err != nil {
		return err
	}
	log.Println(string(out))
	return nil
}

func unstakeValidator(pk crypto.PrivateKey) error {
	log.Printf("Unstaking Validator with Address: %s\n", pk.Address())
	if err := os.WriteFile("./pk.json", []byte("\""+pk.String()+"\""), 0o600); err != nil {
		return err
	}

	//nolint:gosec // G204 Dogfooding CLI
	out, err := exec.Command("/usr/local/bin/client", "--not_interactive=true", "--remote_cli_url="+rpcHost, "Validator", "Unstake", pk.Address().String()).CombinedOutput()
	if err != nil {
		return err
	}
	log.Println(string(out))
	return nil
}
