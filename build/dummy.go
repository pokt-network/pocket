//go:build !debug && !test

// build dummy exists to prevent the gopls from complaining about an empty package without extra configuration. This
// dummy is included when the debug build tag is not included.
package build

// PrivateKeysFile is the pre-generated manifest file for LocalNet debugging
var PrivateKeysFile []byte
