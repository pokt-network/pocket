// // CONSIDERATION: Add a debug tag
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
const privateKeysSecretResourceName = "validators-private-keys"
const kubernetesServiceAccountNamespaceFile = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
const defaultNamespace = "default"

var CurrentNamespace = ""

func init() {
	var err error
	CurrentNamespace, err = getNamespace()
	if err != nil {
		logger.Global.Err(err).Msg("could not get namespace, using \"" + defaultNamespace + "\"")
		CurrentNamespace = defaultNamespace
	}

	logger.Global.Info().Str("namespace", CurrentNamespace).Msg("got new namespace")
}

// FetchValidatorPrivateKeys returns a map corresponding to the data section of
// the validator private keys k8s secret (yaml), located at `privateKeysSecretResourceName`.
// NB: depends on running k8s cluster.
func FetchValidatorPrivateKeys(clientset *kubernetes.Clientset) (map[string]string, error) {
	validatorKeysMap := make(map[string]string)

	privateKeysSecret, err := clientset.CoreV1().Secrets(CurrentNamespace).Get(context.TODO(), privateKeysSecretResourceName, metav1.GetOptions{})
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
	_, err := os.Stat(kubernetesServiceAccountNamespaceFile)
	if err != nil {
		return defaultNamespace, nil
	}

	nsBytes, err := os.ReadFile(kubernetesServiceAccountNamespaceFile)
	if err != nil {
		return "", fmt.Errorf("could not read namespace file: %v", err)
	}

	return string(nsBytes), nil
}
