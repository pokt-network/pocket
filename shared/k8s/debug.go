package k8s

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const privateKeysSecretResourceName = "v1-localnet-validators-private-keys"

func FetchValidatorPrivateKeys(clientset *kubernetes.Clientset) (map[string]string, error) {
	validatorKeysMap := make(map[string]string)

	privateKeysSecret, err := clientset.CoreV1().Secrets("default").Get(context.TODO(), privateKeysSecretResourceName, metav1.GetOptions{})
	if err != nil {
		panic(err)
	}

	for id, privHexString := range privateKeysSecret.Data {
		// it's safe to cast []byte to string here
		validatorKeysMap[id] = string(privHexString)
	}
	return validatorKeysMap, nil
}
