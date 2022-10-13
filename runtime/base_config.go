package runtime

import "github.com/pokt-network/pocket/shared/modules"

var _ modules.BaseConfig = &BaseConfig{}

type BaseConfig struct {
	RootDirectory string `json:"root_directory"`
	PrivateKey    string `json:"private_key"` // TODO (pocket/issues/150) better architecture for key management (keybase, keyfiles, etc.)
}

func (c *BaseConfig) GetRootDirectory() string {
	return c.RootDirectory
}

func (c *BaseConfig) GetPrivateKey() string {
	return c.PrivateKey
}
