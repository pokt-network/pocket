package k8s

import (
	"context"
	"fmt"
	"os"

	"github.com/pokt-network/pocket/logger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

//nolint:gosec // G101 Not a credential
const privateKeysSecretResourceName = "v1-localnet-validators-private-keys"

var CurrentNamespace = ""

func init() {
	var err error
	CurrentNamespace, err = getNamespace()
	if err != nil {
		logger.Global.Err(err).Msg("could not get namespace, using default")
		CurrentNamespace = "default"
	}

	logger.Global.Info().Str("namespace", CurrentNamespace).Msg("using namespace")
}

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

func getNamespace() (string, error) {
	nsFile := "/var/run/secrets/kubernetes.io/serviceaccount/namespace"

	if _, err := os.Stat(nsFile); err == nil {
		nsBytes, err := os.ReadFile(nsFile)
		if err != nil {
			return "", fmt.Errorf("could not read namespace file: %v", err)
		}
		return string(nsBytes), nil
	}

	return "default", nil
}
