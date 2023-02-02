package main

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/pokt-network/pocket/shared/crypto"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	validatorKeysMap = make(map[string]crypto.Ed25519PrivateKey)
)

func main() {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	fetchValidatorPrivateKeys(clientset)

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

		switch event.Type {
		case watch.Added:
			fmt.Printf("Validator %s added\n", service.Name)
			fmt.Printf("Staking Validator with Address: %s\n", validatorKeysMap[extractValidatorId(service.Name)].Address())
		case watch.Deleted:
			fmt.Printf("Validator %s deleted\n", service.Name)
			fmt.Printf("Unstaking Validator with Address: %s\n", validatorKeysMap[extractValidatorId(service.Name)].Address())
		}
	}
}

func fetchValidatorPrivateKeys(clientset *kubernetes.Clientset) {
	private_keys_secret, err := clientset.CoreV1().Secrets("default").Get(context.TODO(), "v1-localnet-validators-private-keys", metav1.GetOptions{})
	if err != nil {
		panic(err)
	}

	for key, value := range private_keys_secret.Data {
		pk, err := parsePrivateKey(bytes.NewReader(value))
		if err != nil {
			panic(err)
		}
		validatorKeysMap[key] = pk
	}
}

func isValidator(service *v1.Service) bool {
	return service.Labels["v1-purpose"] == "validator"
}

func extractValidatorId(validatorName string) string {
	if len(validatorName) >= 3 {
		return validatorName[len(validatorName)-3:]
	}
	return validatorName
}

func parsePrivateKey(reader io.Reader) (pk crypto.Ed25519PrivateKey, err error) {
	if reader == nil {
		return nil, fmt.Errorf("cannot read from reader %v", reader)
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(reader)

	priv := &crypto.Ed25519PrivateKey{}
	err = priv.UnmarshalText(buf.Bytes())
	if err != nil {
		return
	}
	pk = priv.Bytes()
	return
}
