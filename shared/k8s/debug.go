package k8s

import (
	"context"

	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func FetchValidatorPrivateKeys(clientset *kubernetes.Clientset) (map[string]cryptoPocket.KeyPair, error) {
	var validatorKeysMap = make(map[string]cryptoPocket.KeyPair)

	private_keys_secret, err := clientset.CoreV1().Secrets("default").Get(context.TODO(), "v1-localnet-validators-private-keys", metav1.GetOptions{})
	if err != nil {
		panic(err)
	}

	for id, privHexString := range private_keys_secret.Data {
		// Import the keys into the keybase with no passphrase or hint as these are for debug purposes
		keyPair, err := cryptoPocket.CreateNewKeyFromString(string(privHexString), "", "")
		if err != nil {
			return nil, err
		}
		validatorKeysMap[id] = keyPair
	}
	return validatorKeysMap, nil
}
