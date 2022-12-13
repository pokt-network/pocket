package addrbook_provider

import (
	"log"

	typesP2P "github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/shared/modules"
)

// GetAddrBook is a helper function that returns the addrBook depending on the availability of the persistence module
//
// this is a temporary solution simply used to centralize the logic that is going to be refactored in #331 and #203
func GetAddrBook(bus modules.Bus, addrBookProvider typesP2P.AddrBookProvider) typesP2P.AddrBook {
	var (
		addrBook typesP2P.AddrBook
		err      error
	)
	if bus.GetPersistenceModule() == nil {
		// TODO (#203): refactor this.
		addrBook, err = addrBookProvider.ValidatorMapToAddrBook(bus.GetConsensusModule().ValidatorMap())
	} else {
		addrBook, err = addrBookProvider.GetStakedAddrBookAtHeight(bus.GetConsensusModule().CurrentHeight())
	}
	if err != nil {
		log.Fatalf("[ERROR] Error getting addrBook: %v", err)
	}
	return addrBook
}
