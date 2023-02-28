package build

import _ "embed"

// PrivateKeysFile is the pre-generated manifest file for LocalNet debugging
//
//go:embed localnet/manifests/private-keys.yaml
var PrivateKeysFile []byte

func init() {
	if len(PrivateKeysFile) == 0 {
		panic("PrivateKeysFile is empty")
	}
}
