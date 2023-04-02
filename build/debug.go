//go:build debug

package build

import _ "embed"

// DebugKeybaseBackup is a backup of the pre-loaded debug keybase sourced from the manifest file for LocalNet debugging
//
//go:embed debug_keybase/debug_keybase.bak
var DebugKeybaseBackup []byte

func init() {
	if len(DebugKeybaseBackup) == 0 {
		panic("DebugKeybaseBackup is empty")
	}
}
