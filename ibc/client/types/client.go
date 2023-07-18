package types

import (
	"errors"
	"fmt"

	core_types "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/modules"
)

const (
	// https://github.com/cosmos/ibc/blob/main/spec/client/ics-008-wasm-client/README.md
	WasmClientType = "08-wasm"
)

var _ modules.ClientState = &ClientState{}

// ClientType returns the client type.
func (cs *ClientState) ClientType() string { return WasmClientType }

// GetLatestHeight returns the latest height stored.
func (cs *ClientState) GetLatestHeight() modules.Height { return cs.RecentHeight }

// Validate performs a basic validation of the client state fields.
func (cs *ClientState) Validate() error {
	if len(cs.Data) == 0 {
		return errors.New("data cannot be empty")
	}

	lenWasmChecksum := len(cs.WasmChecksum)
	if lenWasmChecksum == 0 {
		return errors.New("wasm checksum cannot be empty")
	}
	if lenWasmChecksum > 32 { // sha256 output is 256 bits long
		return fmt.Errorf("expected 32, got %d", lenWasmChecksum)
	}

	return nil
}

type (
	statusInnerPayload struct{}
	statusPayload      struct {
		Status statusInnerPayload `json:"status"`
	}
)

// Status returns the status of the wasm client.
// The client may be:
// - Active: frozen height is zero and client is not expired
// - Frozen: frozen height is not zero
// - Expired: the latest consensus state timestamp + trusting period <= current time
// - Unauthorized: the client type is not registered as an allowed client type
//
// A frozen client will become expired, so the Frozen status
// has higher precedence.
func (cs *ClientState) Status(clientStore modules.ProvableStore) modules.ClientStatus {
	/*
		payload := &statusPayload{Status: statusInnerPayload{}}
		encodedData, err := json.Marshal(payload)
		if err != nil {
			return modules.UnknownStatus
		}

		// TODO(#912): implement WASM contract querying
	*/
	return modules.ActiveStatus
}

// GetTimestampAtHeight returns the timestamp of the consensus state at the given height.
func (cs *ClientState) GetTimestampAtHeight(clientStore modules.ProvableStore, height modules.Height) (uint64, error) {
	consState, err := GetConsensusState(clientStore, height)
	if err != nil {
		return 0, err
	}
	return consState.GetTimestamp(), nil
}

// Initialise checks that the initial consensus state is an 08-wasm consensus
// state and sets the client state, consensus state in the provided client store.
// It also initializes the wasm contract for the client.
func (cs *ClientState) Initialise(clientStore modules.ProvableStore, consensusState modules.ConsensusState) error {
	consState, ok := consensusState.(*ConsensusState)
	if !ok {
		return errors.New("invalid consensus state type")
	}
	if err := setClientState(clientStore, cs); err != nil {
		return fmt.Errorf("failed to set client state: %w", err)
	}
	if err := setConsensusState(clientStore, consState, cs.GetLatestHeight()); err != nil {
		return fmt.Errorf("failed to set consensus state: %w", err)
	}
	// TODO(#912): implement WASM contract initialisation
	return nil
}

type (
	verifyMembershipInnerPayload struct {
		Height           modules.Height            `json:"height"`
		DelayTimePeriod  uint64                    `json:"delay_time_period"`
		DelayBlockPeriod uint64                    `json:"delay_block_period"`
		Proof            []byte                    `json:"proof"`
		Path             core_types.CommitmentPath `json:"path"`
		Value            []byte                    `json:"value"`
	}
	verifyMembershipPayload struct {
		VerifyMembership verifyMembershipInnerPayload `json:"verify_membership"`
	}
)

// VerifyMembership is a generic proof verification method which verifies a proof
// of the existence of a value at a given CommitmentPath at the specified height.
// The caller is expected to construct the full CommitmentPath from a CommitmentPrefix
// and a standardized path (as defined in ICS 24).
//
// If a zero proof height is passed in, it will fail to retrieve the associated consensus state.
func (cs *ClientState) VerifyMembership(
	clientStore modules.ProvableStore,
	height modules.Height,
	delayTimePeriod, delayBlockPeriod uint64,
	proof, key, value []byte,
) error {
	if cs.GetLatestHeight().LT(height) {
		return fmt.Errorf("client state height < proof height (%d < %d)", cs.GetLatestHeight(), height)
	}

	if _, err := GetConsensusState(clientStore, height); err != nil {
		return errors.New("consensus state not found for proof height")
	}

	/*
		payload := verifyMembershipPayload{
			VerifyMembership: verifyMembershipInnerPayload{
				Height:           height,
				DelayTimePeriod:  delayTimePeriod,
				DelayBlockPeriod: delayBlockPeriod,
				Proof:            proof,
				Path:             key,
				Value:            value,
			},
		}

		// TODO(#912): implement WASM contract method calls
	*/

	return nil
}

type (
	verifyNonMembershipInnerPayload struct {
		Height           modules.Height            `json:"height"`
		DelayTimePeriod  uint64                    `json:"delay_time_period"`
		DelayBlockPeriod uint64                    `json:"delay_block_period"`
		Proof            []byte                    `json:"proof"`
		Path             core_types.CommitmentPath `json:"path"`
	}
	verifyNonMembershipPayload struct {
		VerifyNonMembership verifyNonMembershipInnerPayload `json:"verify_non_membership"`
	}
)

// VerifyNonMembership is a generic proof verification method which verifies
// the absence of a given CommitmentPath at a specified height.
// The caller is expected to construct the full CommitmentPath from a
// CommitmentPrefix and a standardized path (as defined in ICS 24).
//
// If a zero proof height is passed in, it will fail to retrieve the associated consensus state.
func (cs *ClientState) VerifyNonMembership(
	clientStore modules.ProvableStore,
	height modules.Height,
	delayTimePeriod, delayBlockPeriod uint64,
	proof, key []byte,
) error {
	if cs.GetLatestHeight().LT(height) {
		return fmt.Errorf("client state height < proof height (%d < %d)", cs.GetLatestHeight(), height)
	}

	if _, err := GetConsensusState(clientStore, height); err != nil {
		return errors.New("consensus state not found for proof height")
	}

	/*
		payload := verifyNonMembershipPayload{
			VerifyNonMembership: verifyNonMembershipInnerPayload{
				Height:           height,
				DelayTimePeriod:  delayTimePeriod,
				DelayBlockPeriod: delayBlockPeriod,
				Proof:            proof,
				Path:             key,
			},
		}

		// TODO(#912): implement WASM contract method calls
	*/

	return nil
}

type (
	checkForMisbehaviourInnerPayload struct {
		ClientMessage clientMessage `json:"client_message"`
	}
	checkForMisbehaviourPayload struct {
		CheckForMisbehaviour checkForMisbehaviourInnerPayload `json:"check_for_misbehaviour"`
	}
)

// CheckForMisbehaviour detects misbehaviour in a submitted Header message and
// verifies the correctness of a submitted Misbehaviour ClientMessage
func (cs *ClientState) CheckForMisbehaviour(clientStore modules.ProvableStore, clientMsg modules.ClientMessage) bool {
	clientMsgConcrete := clientMessage{
		Header:       nil,
		Misbehaviour: nil,
	}
	switch msg := clientMsg.(type) {
	case *Header:
		clientMsgConcrete.Header = msg
	case *Misbehaviour:
		clientMsgConcrete.Misbehaviour = msg
	}

	if clientMsgConcrete.Header == nil && clientMsgConcrete.Misbehaviour == nil {
		return false
	}

	/*
		inner := checkForMisbehaviourInnerPayload{
			ClientMessage: clientMsgConcrete,
		}
		payload := checkForMisbehaviourPayload{
			CheckForMisbehaviour: inner,
		}

		// TODO(#912): implement WASM contract method calls
	*/

	return true
}
