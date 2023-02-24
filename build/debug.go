package build

import _ "embed"

// PrivateKeysFile is the private keys manifest file for the localnet for debugging
//
//go:embed localnet/manifests/private-keys.yaml
var PrivateKeysFile []byte

func init() {
	if len(PrivateKeysFile) == 0 {
		panic("PrivateKeysFile is empty")
	}
}
