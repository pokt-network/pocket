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
	validatorKeysMap = make(map[string]crypto.Ed25519PrivateKey)
	rpcHost          = "http://v1-validator001:8081"
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
			fmt.Printf("Validator %s added to the cluster\n", service.Name)
			stakeValidator(validatorKeysMap[validatorId], "150000000001", []string{"0001"}, fmt.Sprintf("v1-validator%s:8080", validatorId))
		case watch.Deleted:
			fmt.Printf("Validator %s deleted from the cluster\n", service.Name)
			unstakeValidator(validatorKeysMap[validatorId])
		}
	}
}

func stakeValidator(pk crypto.Ed25519PrivateKey, amount string, chains []string, serviceURL string) error {
	fmt.Printf("Staking Validator with Address: %s\n", pk.Address())
	os.WriteFile("./pk.json", []byte("\""+pk.String()+"\""), 0644)

	out, err := exec.Command("/usr/local/bin/client", "--not_interactive=true", "--remote_cli_url="+rpcHost, "Validator", "Stake", pk.Address().String(), amount, strings.Join(chains, ","), serviceURL).CombinedOutput()
	fmt.Println(string(out))
	if err != nil {
		log.Fatal(err)
	}
	return nil
}
func unstakeValidator(pk crypto.Ed25519PrivateKey) error {
	fmt.Printf("Unstaking Validator with Address: %s\n", pk.Address())
	os.WriteFile("./pk.json", []byte("\""+pk.String()+"\""), 0644)

	out, err := exec.Command("/usr/local/bin/client", "--not_interactive=true", "--remote_cli_url="+rpcHost, "Validator", "Unstake", pk.Address().String()).CombinedOutput()
	fmt.Println(string(out))
	if err != nil {
		log.Fatal(err)
	}
	return nil
}
