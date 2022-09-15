package test_artifacts

import "fmt"

const (
	genesisStatePostfix = "_genesis_state"
)

func GetGenesisFileName(moduleName string) string {
	return fmt.Sprintf("%s%s", moduleName, genesisStatePostfix)
}
