package runtime

import (
	"encoding/json"
	"os"

	"github.com/pokt-network/pocket/runtime/genesis"
)

func parseGenesisJSON(genesisPath string) (g *genesis.GenesisState, err error) {
	data, err := os.ReadFile(genesisPath)
	if err != nil {
		return
	}

	// general genesis file
	g = new(genesis.GenesisState)
	err = json.Unmarshal(data, &g)
	return
}
