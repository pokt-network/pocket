//go:build !debug

// build dummy exists to prevent the gopls from complaining about an empty package without extra configuration. This
// dummy is included when the debug build tag is not included.
package build

// DebugKeybaseBackup is a backup of the pre-loaded debug keybase sourced from the manifest file for LocalNet debugging
var DebugKeybaseBackup []byte
