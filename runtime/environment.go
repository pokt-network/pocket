package runtime

import "os"

// GetEnv returns the value of the environment variable named by the key.
//
// It returns the defaultValue if the variable is not present.
func GetEnv(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}

// IsProcessRunningInsideKubernetes returns true if the process is running inside a Kubernetes cluster.
func IsProcessRunningInsideKubernetes() bool {
	_, exists := os.LookupEnv("KUBERNETES_SERVICE_HOST")
	return exists
}
