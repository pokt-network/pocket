package configs

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func createTempConfigFile(content string) (string, error) {
	tmpfile, err := ioutil.TempFile("", "test_config_*.json")
	if err != nil {
		return "", err
	}

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		return "", err
	}

	if err := tmpfile.Close(); err != nil {
		return "", err
	}

	return tmpfile.Name(), nil
}

func TestParseConfig(t *testing.T) {

	tests := []struct {
		name          string
		configContent string
		validateFunc  func(t *testing.T, config *Config)
	}{
		{
			name: "Consensus config",
			configContent: `{
		"consensus": {
			"max_mempool_bytes": 1000000,
			"pacemaker_config": {
				"timeout_msec": 5000
			},
			"private_key": "12345",
			"server_mode_enabled": true
		}
	}`,
			validateFunc: func(t *testing.T, config *Config) {
				assert.Equal(t, uint64(1000000), config.Consensus.MaxMempoolBytes)
				assert.Equal(t, uint64(5000), config.Consensus.PacemakerConfig.TimeoutMsec)
				assert.Equal(t, "12345", config.Consensus.PrivateKey)
				assert.Equal(t, true, config.Consensus.ServerModeEnabled)
			},
		},
		{
			name: "Utility config",
			configContent: `{
		"utility": {
			"max_mempool_transaction_bytes": 2000000,
			"max_mempool_transactions": 10000
		}
	}`,
			validateFunc: func(t *testing.T, config *Config) {
				assert.Equal(t, uint64(2000000), config.Utility.MaxMempoolTransactionBytes)
				assert.Equal(t, uint32(10000), config.Utility.MaxMempoolTransactions)
			},
		},
		{
			name: "P2P config",
			configContent: `{
		"p2p": {
			"hostname": "localhost",
			"port": 3000,
			"use_rain_tree": false,
			"private_key": "67890",
			"max_mempool_count": 20000
		}
	}`,
			validateFunc: func(t *testing.T, config *Config) {
				assert.Equal(t, "localhost", config.P2P.Hostname)
				assert.Equal(t, uint32(3000), config.P2P.Port)
				assert.Equal(t, false, config.P2P.UseRainTree)
				assert.Equal(t, "67890", config.P2P.PrivateKey)
				assert.Equal(t, uint64(20000), config.P2P.MaxMempoolCount)
			},
		},
		{
			name: "Telemetry config",
			configContent: `{
		"telemetry": {
			"enabled": true,
			"address": "0.0.0.0:8000",
			"endpoint": "/metrics"
		}
	}`,
			validateFunc: func(t *testing.T, config *Config) {
				assert.Equal(t, true, config.Telemetry.Enabled)
				assert.Equal(t, "0.0.0.0:8000", config.Telemetry.Address)
				assert.Equal(t, "/metrics", config.Telemetry.Endpoint)
			},
		},
		{
			name: "File-based keybase config",
			configContent: `{
		"keybase": {
			"config": {
				"file": {
					"path": "/tmp/keybase"
				}
			}
		}
	}`,
			validateFunc: func(t *testing.T, config *Config) {
				assert.NotNil(t, config.Keybase.Config)
				fileConfig, ok := config.Keybase.Config.(*KeybaseConfig_File)
				assert.True(t, ok, "Expected file-based keybase configuration")
				assert.Equal(t, "/tmp/keybase", fileConfig.File.Path)
			},
		},
		{
			name: "Vault-based keybase config",
			configContent: `{
	"keybase": {
		"config": {
			"vault": {
				"addr": "http://localhost:8200",
				"token": "test_token",
				"mountPath": "secrets"
			}
		}
	}
}`,
			validateFunc: func(t *testing.T, config *Config) {
				assert.NotNil(t, config.Keybase.Config)
				vaultConfig, ok := config.Keybase.Config.(*KeybaseConfig_Vault)
				assert.True(t, ok, "Expected vault-based keybase configuration")
				assert.Equal(t, "http://localhost:8200", vaultConfig.Vault.Addr)
				assert.Equal(t, "test_token", vaultConfig.Vault.Token)
				assert.Equal(t, "secrets", vaultConfig.Vault.MountPath)
			},
		},
		{
			name: "Vault-based keybase config with mount_path variant",
			configContent: `{
	"keybase": {
		"config": {
			"vault": {
				"addr": "http://localhost:8200",
				"token": "test_token",
				"mount_path": "secrets"
			}
		}
	}
}`,
			validateFunc: func(t *testing.T, config *Config) {
				assert.NotNil(t, config.Keybase.Config)
				vaultConfig, ok := config.Keybase.Config.(*KeybaseConfig_Vault)
				assert.True(t, ok, "Expected vault-based keybase configuration")
				assert.Equal(t, "http://localhost:8200", vaultConfig.Vault.Addr)
				assert.Equal(t, "test_token", vaultConfig.Vault.Token)
				assert.Equal(t, "secrets", vaultConfig.Vault.MountPath)
			},
		},
		{
			name: "Invalid keybase config returns default keybase config",
			configContent: `{
		"keybase": {
			"config": {
				"invalid": {
					"addr": "http://localhost:8200"
				}
			}
		}
	}`,
			validateFunc: func(t *testing.T, config *Config) {
				assert.NotNil(t, config.Keybase.Config)
				fileConfig, ok := config.Keybase.Config.(*KeybaseConfig_File)
				assert.True(t, ok, "Expected file-based keybase configuration")
				assert.NotNil(t, fileConfig.File.Path)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmpfile, err := createTempConfigFile(test.configContent)
			assert.NoError(t, err)

			config := ParseConfig(tmpfile)
			test.validateFunc(t, config)

			// Clean up temp file
			os.Remove(tmpfile)
		})
	}
}
