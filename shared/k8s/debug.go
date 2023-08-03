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
const (
	privateKeysSecretResourceNameValidators   = "validators-private-keys"
	privateKeysSecretResourceNameServicers    = "servicers-private-keys"
	privateKeysSecretResourceNameWatchers     = "watchers-private-keys"
	privateKeysSecretResourceNameApplications = "applications-private-keys"
	kubernetesServiceAccountNamespaceFile     = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
	defaultNamespace                          = "default"
)

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
// the validator private keys Kubernetes secret.
func FetchValidatorPrivateKeys(clientset *kubernetes.Clientset) (map[string]string, error) {
	return fetchPrivateKeys(clientset, privateKeysSecretResourceNameValidators)
}

// FetchServicerPrivateKeys returns a map corresponding to the data section of
// the servicer private keys Kubernetes secret.
func FetchServicerPrivateKeys(clientset *kubernetes.Clientset) (map[string]string, error) {
	return fetchPrivateKeys(clientset, privateKeysSecretResourceNameServicers)
}

// FetchWatcherPrivateKeys returns a map corresponding to the data section of
// the watcher private keys Kubernetes secret.
func FetchWatcherPrivateKeys(clientset *kubernetes.Clientset) (map[string]string, error) {
	return fetchPrivateKeys(clientset, privateKeysSecretResourceNameWatchers)
}

// FetchApplicationPrivateKeys returns a map corresponding to the data section of
// the application private keys Kubernetes secret.
func FetchApplicationPrivateKeys(clientset *kubernetes.Clientset) (map[string]string, error) {
	return fetchPrivateKeys(clientset, privateKeysSecretResourceNameApplications)
}

// fetchPrivateKeys returns a map corresponding to the data section of
// the private keys Kubernetes secret for the specified resource name and actor.
func fetchPrivateKeys(clientset *kubernetes.Clientset, resourceName string) (map[string]string, error) {
	privateKeysMap := make(map[string]string)
	privateKeysSecret, err := clientset.CoreV1().Secrets(CurrentNamespace).Get(context.TODO(), resourceName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	for id, privHexString := range privateKeysSecret.Data {
		// It's safe to cast []byte to string here
		privateKeysMap[id] = string(privHexString)
	}
	return privateKeysMap, nil
}

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
