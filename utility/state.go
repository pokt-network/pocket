package utility

import (
	"fmt"
	"log"

	"github.com/pokt-network/pocket/shared/types"
	typesUtil "github.com/pokt-network/pocket/utility/types"
)

func (u *UtilityContext) updateStateCommitment() types.Error {
	// Update the Merkle Tree associated with each actor
	for _, actorType := range typesUtil.ActorTypes {
		// Need to get all the actors updated at this height
		switch actorType {
		case typesUtil.ActorType_App:
			apps, err := u.Context.GetAppsUpdated(u.LatestHeight) // shouldn't need to pass in a height here
			if err != nil {
				return types.NewError(types.Code(42), "Couldn't figure out apps updated")
			}
			fmt.Println("apps: ", apps)
		case typesUtil.ActorType_Val:
			fallthrough
		case typesUtil.ActorType_Fish:
			fallthrough
		case typesUtil.ActorType_Node:
			fallthrough
		default:
			log.Fatalf("Not supported yet")
		}
	}

	// TODO: Update Merkle Tree for Accounts

	// TODO: Update Merkle Tree for Pools

	// TODO:Update Merkle Tree for Blocks

	// TODO: Update Merkle Tree for Params

	// TODO: Update Merkle Tree for Flags

	return nil
}
