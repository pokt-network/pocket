package types

import (
	"github.com/pokt-network/pocket/shared/modules"
)

type (
	clientMessage struct {
		Header       *Header       `json:"header,omitempty"`
		Misbehaviour *Misbehaviour `json:"misbehaviour,omitempty"`
	}
	verifyClientMessageInnerPayload struct {
		ClientMessage clientMessage `json:"client_message"`
	}
	verifyClientMessagePayload struct {
		VerifyClientMessage verifyClientMessageInnerPayload `json:"verify_client_message"`
	}
)

// VerifyClientMessage must verify a ClientMessage. A ClientMessage could be a Header,
// Misbehaviour, or batch update. It must handle each type of ClientMessage appropriately.
//
// Calls to CheckForMisbehaviour, UpdateState, and UpdateStateOnMisbehaviour will
// assume that the content of the ClientMessage has been verified and can be trusted
func (cs *ClientState) VerifyClientMessage(clientStore modules.ProvableStore, clientMsg modules.ClientMessage) error {
	clientMsgConcrete := clientMessage{
		Header:       nil,
		Misbehaviour: nil,
	}
	switch clientMsg := clientMsg.(type) {
	case *Header:
		clientMsgConcrete.Header = clientMsg
	case *Misbehaviour:
		clientMsgConcrete.Misbehaviour = clientMsg
	}

	/*
		inner := verifyClientMessageInnerPayload{
			ClientMessage: clientMsgConcrete,
		}
		payload := verifyClientMessagePayload{
			VerifyClientMessage: inner,
		}

		// TODO(#912): implement WASM method calls
	*/

	return nil
}

type (
	updateStateInnerPayload struct {
		ClientMessage clientMessage `json:"client_message"`
	}
	updateStatePayload struct {
		UpdateState updateStateInnerPayload `json:"update_state"`
	}
)

// UpdateState updates and stores as necessary any associated information for an
// IBC client. Upon successful update, a consensus height is returned.
//
// Client state and new consensus states are updated in the store by the contract
// Assumes the ClientMessage has already been verified
func (cs *ClientState) UpdateState(clientStore modules.ProvableStore, clientMsg modules.ClientMessage) (modules.Height, error) {
	/*
		header, ok := clientMsg.(*Header)
		if !ok {
			return nil, errors.New("client message must be a header")
		}

		payload := updateStatePayload{
			UpdateState: updateStateInnerPayload{
				ClientMessage: clientMessage{
					Header: header,
				},
			},
		}

		// TODO(#912): implement WASM method calls
	*/

	return clientMsg.(*Header).Height, nil
}

type (
	updateStateOnMisbehaviourInnerPayload struct {
		ClientMessage clientMessage `json:"client_message"`
	}
	updateStateOnMisbehaviourPayload struct {
		UpdateStateOnMisbehaviour updateStateOnMisbehaviourInnerPayload `json:"update_state_on_misbehaviour"`
	}
)

// UpdateStateOnMisbehaviour should perform appropriate state changes on a
// client state given that misbehaviour has been detected and verified
// Client state is updated in the store by contract.
func (cs *ClientState) UpdateStateOnMisbehaviour(clientStore modules.ProvableStore, clientMsg modules.ClientMessage) error {
	var clientMsgConcrete clientMessage
	switch clientMsg := clientMsg.(type) {
	case *Header:
		clientMsgConcrete.Header = clientMsg
	case *Misbehaviour:
		clientMsgConcrete.Misbehaviour = clientMsg
	}

	/*
		inner := updateStateOnMisbehaviourInnerPayload{
			ClientMessage: clientMsgConcrete,
		}

		payload := updateStateOnMisbehaviourPayload{
			UpdateStateOnMisbehaviour: inner,
		}

		// TODO(#912): implement WASM method calls
	*/

	return nil
}
