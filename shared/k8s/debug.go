// // CONSIDERATION: Add a debug tag
package k8s

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/tools/clientcmd"

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

// ADDPR: add the following functions in a separate PR: FetchServicerPrivateKeys and FetchAppPrivateKeys

func getNamespace() (string, error) {
	_, err := os.Stat(kubernetesServiceAccountNamespaceFile)
	if err == nil {
		return getNamespaceSvcAcct()
	}
	if ns, err := getClientNamespace(); err == nil {
		return ns, nil
	}
	return defaultNamespace, nil
}

func getNamespaceSvcAcct() (string, error) {
	nsBytes, err := os.ReadFile(kubernetesServiceAccountNamespaceFile)
	if err != nil {
		return "", fmt.Errorf("could not read namespace file: %v", err)
	}

	return string(nsBytes), nil
}

func getClientNamespace() (string, error) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home dir: %w", err)
	}
	kubeConfigPath := filepath.Join(userHomeDir, ".kube", "config")
	kubeconfig := flag.String("kubeconfig", kubeConfigPath, "(optional) absolute path to the kubeconfig file")
	_, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return "", fmt.Errorf("could not build config from flags: %v", err)
	}

	// use the NewDefaultClientConfigLoadingRules() function to load the kubeconfig file
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()

	// use the NewNonInteractiveDeferredLoadingClientConfig() function to get the config
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeconfigClient := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	// use the Namespace() function to get the current namespace
	namespace, _, err := kubeconfigClient.Namespace()
	return namespace, err
}
