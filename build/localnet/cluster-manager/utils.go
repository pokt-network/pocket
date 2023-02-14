package main

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/pokt-network/pocket/shared/crypto"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func fetchValidatorPrivateKeys(clientset *kubernetes.Clientset, validatorKeysMap map[string]crypto.PrivateKey) {
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
	if _, err := buf.ReadFrom(reader); err != nil {
		return nil, err
	}

	priv := &crypto.Ed25519PrivateKey{}
	err = priv.UnmarshalText(buf.Bytes())
	if err != nil {
		return
	}
	pk = priv.Bytes()
	return
}
