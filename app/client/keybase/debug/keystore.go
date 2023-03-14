//go:build debug

package debug

import (
	"fmt"
	"os"
	"sync"

	"github.com/pokt-network/pocket/app/client/keybase"
	"github.com/pokt-network/pocket/build"
	"github.com/pokt-network/pocket/logger"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"gopkg.in/yaml.v2"
)

const (
	// NOTE: This is the number of validators in the private-keys.yaml manifest file
	numValidators      = 999
	debugKeybaseSuffix = "/.pocket/keys"
)

var (
	// TODO: Allow users to override this value via `datadir` flag or env var or config file
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

	validatorKeysPairMap, err = parseValidatorPrivateKeysFromEmbeddedYaml()

	if err != nil {
		return err
	}

	// Create/Open the keybase at `$HOME/.pocket/keys`
	kb, err := keybase.NewKeybase(debugKeybasePath)
	if err != nil {
		return err
	}
	db := kb.GetBadgerDB()

	// Add the keys if the keybase contains less than 999
	curAddr, _, err := kb.GetAll()
	if err != nil {
		return err
	}

	// Add validator addresses if not present
	if len(curAddr) < numValidators {
		logger.Global.Debug().Msgf(fmt.Sprintf("Debug keybase initializing... Adding %d validator keys to the keybase", numValidators-len(curAddr)))

		// Use writebatch to speed up bulk insert
		wb := db.NewWriteBatch()

		// Create a channel to receive errors from goroutines
		errCh := make(chan error, numValidators)

		// Create a WaitGroup to wait for all goroutines to finish
		var wg sync.WaitGroup
		wg.Add(numValidators)

		for _, privHexString := range validatorKeysPairMap {
			go func(privHexString string) {
				defer wg.Done()

				// Import the keys into the keybase with no passphrase or hint as these are for debug purposes
				keyPair, err := cryptoPocket.CreateNewKeyFromString(privHexString, "", "")
				if err != nil {
					errCh <- err
					return
				}

				// Use key address as key in DB
				addrKey := keyPair.GetAddressBytes()

				// Encode KeyPair into []byte for value
				keypairBz, err := keyPair.Marshal()
				if err != nil {
					errCh <- err
					return
				}
				if err := wb.Set(addrKey, keypairBz); err != nil {
					errCh <- err
					return
				}
			}(privHexString)
		}

		// Wait for all goroutines to finish
		wg.Wait()

		// Check if any goroutines returned an error
		select {
		case err := <-errCh:
			return err
		default:
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

// parseValidatorPrivateKeysFromEmbeddedYaml fetches the validator private keys from the embedded build/localnet/manifests/private-keys.yaml manifest file.
func parseValidatorPrivateKeysFromEmbeddedYaml() (map[string]string, error) {

	// Parse the YAML file and load into the config struct
	var config struct {
		ApiVersion string            `yaml:"apiVersion"`
		Kind       string            `yaml:"kind"`
		MetaData   map[string]string `yaml:"metadata"`
		Type       string            `yaml:"type"`
		StringData map[string]string `yaml:"stringData"`
	}
	if err := yaml.Unmarshal(build.PrivateKeysFile, &config); err != nil {
		return nil, err
	}
	validatorKeysMap := make(map[string]string)

	for id, privHexString := range config.StringData {
		validatorKeysMap[id] = privHexString
	}
	return validatorKeysMap, nil
}
