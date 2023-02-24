package debug

import (
	"fmt"
	"github.com/pokt-network/pocket/shared/converters"
	"os"
	"path/filepath"
	r "runtime"

	"github.com/pokt-network/pocket/app/client/keybase"
	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/runtime"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	pocketk8s "github.com/pokt-network/pocket/shared/k8s"
	"gopkg.in/yaml.v2"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	// NOTE: This is the number of validators in the private-keys.yaml manifest file
	numValidators       = 999
	debugKeybaseSuffix  = "/.pocket/keys"
	privateKeysYamlFile = "../../../../../build/localnet/manifests/private-keys.yaml"
)

var (
	// TODO: Allow users to override this value via `datadir` flag
	debugKeybasePath string
)

// Initialise the debug keybase with the 999 validator keys from the private-keys manifest file
func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Global.Fatal().Err(err).Msg("Cannot find user home directory")
	}
	debugKeybasePath = homeDir + debugKeybaseSuffix

	// Initialise the debug keybase with the 999 validators
	if err := initializeDebugKeybase(); err != nil {
		logger.Global.Fatal().Err(err).Msg("Cannot initialise the keybase with the validator keys")
	}
}

func initializeDebugKeybase() error {
	var (
		validatorKeysPairMap map[string]string
		err                  error
	)

	if runtime.IsProcessRunningInsideKubernetes() {
		validatorKeysPairMap, err = fetchValidatorPrivateKeysFromK8S()
	} else {
		validatorKeysPairMap, err = fetchValidatorPrivateKeysFromFile()
	}
	if err != nil {
		return err
	}

	// Create/Open the keybase at `$HOME/.pocket/keys`
	kb, err := keybase.NewBadgerKeybase(debugKeybasePath)
	if err != nil {
		return err
	}
	db, err := kb.GetBadgerDB()
	if err != nil {
		return err
	}

	// Add the keys if the keybase contains less than 999
	curAddr, _, err := kb.GetAll()
	if err != nil {
		return err
	}

	// Add validator addresses if not present
	if len(curAddr) < numValidators {
		fmt.Println("Rehydrating keybase from private-keys.yaml ...")
		// Use writebatch to speed up bulk insert
		wb := db.NewWriteBatch()
		for _, privHexString := range validatorKeysPairMap {
			// Import the keys into the keybase with no passphrase or hint as these are for debug purposes
			keyPair, err := cryptoPocket.CreateNewKeyFromString(privHexString, "", "")
			if err != nil {
				return err
			}

			// Use key address as key in DB
			addrKey := keyPair.GetAddressBytes()

			// Encode KeyPair into []byte for value
			keypairBz, err := keyPair.Marshal()
			if err != nil {
				return err
			}
			if err := wb.Set(addrKey, keypairBz); err != nil {
				return err
			}
		}
		if err := wb.Flush(); err != nil {
			return err
		}
	}

	// Close DB connection
	if err := kb.Stop(); err != nil {
		return err
	}

	return nil
}

func fetchValidatorPrivateKeysFromK8S() (map[string]string, error) {
	// Initialize Kubernetes client
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Kubernetes config: %w", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Kubernetes client: %w", err)
	}

	// Fetch validator private keys from Kubernetes
	validatorKeysPairMap, err := pocketk8s.FetchValidatorPrivateKeys(clientset)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch validator private keys from Kubernetes: %w", err)
	}
	return validatorKeysPairMap, nil
}

func fetchValidatorPrivateKeysFromFile() (map[string]string, error) {
	// BUG: When running the CLI using the build binary (i.e. `p1`), it searched for the private-keys.yaml file in `github.com/pokt-network/pocket/build/localnet/manifests/private-keys.yaml`
	// Get private keys from manifest file
	_, current, _, _ := r.Caller(0)
	//nolint:gocritic // Use path to find private-keys yaml file from being called in any location in the repo
	yamlFile := filepath.Join(current, privateKeysYamlFile)
	if exists, err := converters.FileExists(yamlFile); !exists || err != nil {
		return nil, fmt.Errorf("unable to find YAML file: %s", yamlFile)
	}

	// Parse the YAML file and load into the config struct
	yamlData, err := os.ReadFile(yamlFile)
	if err != nil {
		return nil, err
	}
	var config struct {
		ApiVersion string            `yaml:"apiVersion"`
		Kind       string            `yaml:"kind"`
		MetaData   map[string]string `yaml:"metadata"`
		Type       string            `yaml:"type"`
		StringData map[string]string `yaml:"stringData"`
	}
	if err := yaml.Unmarshal(yamlData, &config); err != nil {
		return nil, err
	}
	validatorKeysMap := make(map[string]string)

	for id, privHexString := range config.StringData {
		validatorKeysMap[id] = privHexString
	}
	return validatorKeysMap, nil
}
