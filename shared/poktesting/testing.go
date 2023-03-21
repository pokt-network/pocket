//go:build test

package poktesting

import (
	"github.com/pokt-network/pocket/build"
	"gopkg.in/yaml.v2"
)

// ParseValidatorPrivateKeysFromEmbeddedYaml fetches the validator private keys from the embedded build/localnet/manifests/private-keys.yaml manifest file.
func ParseValidatorPrivateKeysFromEmbeddedYaml() (map[string]string, error) {

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
