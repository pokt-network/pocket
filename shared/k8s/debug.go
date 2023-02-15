package k8s

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func FetchValidatorPrivateKeys(clientset *kubernetes.Clientset) (map[string]string, error) {
	var validatorKeysMap = make(map[string]string)

	private_keys_secret, err := clientset.CoreV1().Secrets("default").Get(context.TODO(), "v1-localnet-validators-private-keys", metav1.GetOptions{})
	if err != nil {
		panic(err)
	}

	for id, privHexString := range private_keys_secret.Data {
		validatorKeysMap[id] = string(privHexString)
	}
	return validatorKeysMap, nil
}
