package client

import (
	client_types "github.com/pokt-network/pocket/ibc/client/types"
	"github.com/pokt-network/pocket/shared/codec"
	core_types "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/modules"
)

// emitCreateClientEvent emits a create client event
func (c *clientManager) emitCreateClientEvent(clientId string, clientState modules.ClientState) error {
	return c.GetBus().GetEventLogger().EmitEvent(
		&core_types.IBCEvent{
			Topic: client_types.EventTopicCreateClient,
			Attributes: []*core_types.Attribute{
				core_types.NewAttribute(client_types.AttributeKeyClientID, []byte(clientId)),
				core_types.NewAttribute(client_types.AttributeKeyClientType, []byte(clientState.ClientType())),
				core_types.NewAttribute(client_types.AttributeKeyConsensusHeight, []byte(clientState.GetLatestHeight().String())),
			},
		},
	)
}

// emitUpdateClientEvent emits an update client event
func (c *clientManager) emitUpdateClientEvent(
	clientId, clientType string,
	consensusHeight modules.Height,
	clientMessage modules.ClientMessage,
) error {
	// Marshall the client message
	clientMsgBz, err := codec.GetCodec().Marshal(clientMessage)
	if err != nil {
		return err
	}

	return c.GetBus().GetEventLogger().EmitEvent(
		&core_types.IBCEvent{
			Topic: client_types.EventTopicUpdateClient,
			Attributes: []*core_types.Attribute{
				core_types.NewAttribute(client_types.AttributeKeyClientID, []byte(clientId)),
				core_types.NewAttribute(client_types.AttributeKeyClientType, []byte(clientType)),
				core_types.NewAttribute(client_types.AttributeKeyConsensusHeight, []byte(consensusHeight.String())),
				core_types.NewAttribute(client_types.AttributeKeyHeader, clientMsgBz),
			},
		},
	)
}

// emitUpgradeClientEvent emits an upgrade client event
func (c *clientManager) emitUpgradeClientEvent(clientId string, clientState modules.ClientState) error {
	return c.GetBus().GetEventLogger().EmitEvent(
		&core_types.IBCEvent{
			Topic: client_types.EventTopicUpdateClient,
			Attributes: []*core_types.Attribute{
				core_types.NewAttribute(client_types.AttributeKeyClientID, []byte(clientId)),
				core_types.NewAttribute(client_types.AttributeKeyClientType, []byte(clientState.ClientType())),
				core_types.NewAttribute(client_types.AttributeKeyConsensusHeight, []byte(clientState.GetLatestHeight().String())),
			},
		},
	)
}

// emitSubmitMisbehaviourEvent emits a submit misbehaviour event
func (c *clientManager) emitSubmitMisbehaviourEvent(clientId string, clientState modules.ClientState) error {
	return c.GetBus().GetEventLogger().EmitEvent(
		&core_types.IBCEvent{
			Topic: client_types.EventTopicSubmitMisbehaviour,
			Attributes: []*core_types.Attribute{
				core_types.NewAttribute(client_types.AttributeKeyClientID, []byte(clientId)),
				core_types.NewAttribute(client_types.AttributeKeyClientType, []byte(clientState.ClientType())),
			},
		},
	)
}
