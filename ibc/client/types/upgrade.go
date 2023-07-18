package types

import (
	"fmt"

	"github.com/pokt-network/pocket/shared/modules"
)

type (
	verifyUpgradeAndUpdateStateInnerPayload struct {
		UpgradeClientState         modules.ClientState    `json:"upgrade_client_state"`
		UpgradeConsensusState      modules.ConsensusState `json:"upgrade_consensus_state"`
		ProofUpgradeClient         []byte                 `json:"proof_upgrade_client"`
		ProofUpgradeConsensusState []byte                 `json:"proof_upgrade_consensus_state"`
	}
	verifyUpgradeAndUpdateStatePayload struct {
		VerifyUpgradeAndUpdateState verifyUpgradeAndUpdateStateInnerPayload `json:"verify_upgrade_and_update_state"`
	}
)

// VerifyUpgradeAndUpdateState, on a successful verification expects the contract
// to update the new client state, consensus state, and any other client metadata.
func (cs *ClientState) VerifyUpgradeAndUpdateState(
	clientStore modules.ProvableStore,
	upgradedClient modules.ClientState,
	upgradedConsState modules.ConsensusState,
	proofUpgradeClient, proofUpgradeConsState []byte,
) error {
	/*
		wasmUpgradeClientState, ok := upgradedClient.(*ClientState)
		if !ok {
			return errors.New("upgraded client state must be Wasm ClientState")
		}

		wasmUpgradeConsState, ok := upgradedConsState.(*ConsensusState)
		if !ok {
			return errors.New("upgraded consensus state must be Wasm ConsensusState")
		}
	*/

	// last height of current counterparty chain must be client's latest height
	lastHeight := cs.GetLatestHeight()

	if !upgradedClient.GetLatestHeight().GT(lastHeight) {
		return fmt.Errorf("upgraded client height %s must be greater than current client height %s",
			upgradedClient.GetLatestHeight(), lastHeight,
		)
	}

	// Must prove against latest consensus state to ensure we are verifying
	// against latest upgrade plan.
	_, err := GetConsensusState(clientStore, lastHeight)
	if err != nil {
		return fmt.Errorf("could not retrieve consensus state for height %s", lastHeight)
	}

	/*
		payload := verifyUpgradeAndUpdateStatePayload{
			VerifyUpgradeAndUpdateState: verifyUpgradeAndUpdateStateInnerPayload{
				UpgradeClientState:         upgradedClient,
				UpgradeConsensusState:      upgradedConsState,
				ProofUpgradeClient:         proofUpgradeClient,
				ProofUpgradeConsensusState: proofUpgradeConsState,
			},
		}

		// TODO(#912): implement WASM contract initialisation
	*/

	return nil
}
