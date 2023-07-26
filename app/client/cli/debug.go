package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/pokt-network/pocket/app/client/cli/helpers"
	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/p2p/providers/peerstore_provider"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/shared/modules"
)

// TECHDEBT: Lowercase variables / constants that do not need to be exported.
const (
	PromptResetToGenesis         string = "ResetToGenesis (broadcast)"
	PromptPrintNodeState         string = "PrintNodeState (broadcast)"
	PromptTriggerNextView        string = "TriggerNextView (broadcast)"
	PromptTogglePacemakerMode    string = "TogglePacemakerMode (broadcast)"
	PromptShowLatestBlockInStore string = "ShowLatestBlockInStore (anycast)"
	PromptSendMetadataRequest    string = "MetadataRequest (broadcast)"
	PromptSendBlockRequest       string = "BlockRequest (broadcast)"
)

var items = []string{
	PromptPrintNodeState,
	PromptTriggerNextView,
	PromptTogglePacemakerMode,
	PromptResetToGenesis,
	PromptShowLatestBlockInStore,
	PromptSendMetadataRequest,
	PromptSendBlockRequest,
}

func init() {
	dbgUI := newDebugUICommand()
	dbgUI.AddCommand(newDebugUISubCommands()...)
	rootCmd.AddCommand(dbgUI)
}

// newDebugUISubCommands builds out the list of debug subcommands by matching the
// handleSelect dispatch to the appropriate command.
// * To add a debug subcommand, you must add it to the `items` array and then
// write a function handler to match for it in `handleSelect`.
func newDebugUISubCommands() []*cobra.Command {
	commands := make([]*cobra.Command, len(items))
	for idx, promptItem := range items {
		commands[idx] = &cobra.Command{
			Use:               promptItem,
			PersistentPreRunE: helpers.P2PDependenciesPreRunE,
			Run: func(cmd *cobra.Command, _ []string) {
				handleSelect(cmd, cmd.Use)
			},
			ValidArgs: items,
		}
	}
	return commands
}

// newDebugUICommand returns the cobra CLI for the Debug UI interface.
func newDebugUICommand() *cobra.Command {
	return &cobra.Command{
		Aliases:           []string{"dui"},
		Use:               "DebugUI",
		Short:             "Debug selection ui for rapid development",
		Args:              cobra.MaximumNArgs(0),
		PersistentPreRunE: helpers.P2PDependenciesPreRunE,
		RunE:              runDebug,
	}
}

func runDebug(cmd *cobra.Command, _ []string) (err error) {
	for {
		if selection, err := promptGetInput(); err == nil {
			handleSelect(cmd, selection)
		} else {
			return err
		}
	}
}

func promptGetInput() (string, error) {
	prompt := promptui.Select{
		Label: "Select an action",
		Items: items,
		Size:  len(items),
	}

	_, result, err := prompt.Run()

	if err == promptui.ErrInterrupt {
		os.Exit(0)
	}

	if err != nil {
		logger.Global.Error().Err(err).Msg("Prompt failed")
		return "", err
	}

	return result, nil
}

func handleSelect(cmd *cobra.Command, selection string) {
	switch selection {
	case PromptResetToGenesis:
		m := &messaging.DebugMessage{
			Action:  messaging.DebugMessageAction_DEBUG_CONSENSUS_RESET_TO_GENESIS,
			Type:    messaging.DebugMessageRoutingType_DEBUG_MESSAGE_TYPE_BROADCAST,
			Message: nil,
		}
		broadcastDebugMessage(cmd, m)
	case PromptPrintNodeState:
		m := &messaging.DebugMessage{
			Action:  messaging.DebugMessageAction_DEBUG_CONSENSUS_PRINT_NODE_STATE,
			Type:    messaging.DebugMessageRoutingType_DEBUG_MESSAGE_TYPE_BROADCAST,
			Message: nil,
		}
		broadcastDebugMessage(cmd, m)
	case PromptTriggerNextView:
		m := &messaging.DebugMessage{
			Action:  messaging.DebugMessageAction_DEBUG_CONSENSUS_TRIGGER_NEXT_VIEW,
			Type:    messaging.DebugMessageRoutingType_DEBUG_MESSAGE_TYPE_BROADCAST,
			Message: nil,
		}
		broadcastDebugMessage(cmd, m)
	case PromptTogglePacemakerMode:
		m := &messaging.DebugMessage{
			Action:  messaging.DebugMessageAction_DEBUG_CONSENSUS_TOGGLE_PACE_MAKER_MODE,
			Type:    messaging.DebugMessageRoutingType_DEBUG_MESSAGE_TYPE_BROADCAST,
			Message: nil,
		}
		broadcastDebugMessage(cmd, m)
	case PromptShowLatestBlockInStore:
		m := &messaging.DebugMessage{
			Action: messaging.DebugMessageAction_DEBUG_SHOW_LATEST_BLOCK_IN_STORE,
			// NB: Anycast because we technically accept any node but we arbitrarily choose the first in our address book.
			Type:    messaging.DebugMessageRoutingType_DEBUG_MESSAGE_TYPE_ANYCAST,
			Message: nil,
		}
		sendDebugMessage(cmd, m)
	case PromptSendMetadataRequest:
		m := &messaging.DebugMessage{
			Action:  messaging.DebugMessageAction_DEBUG_CONSENSUS_SEND_METADATA_REQ,
			Type:    messaging.DebugMessageRoutingType_DEBUG_MESSAGE_TYPE_BROADCAST,
			Message: nil,
		}
		broadcastDebugMessage(cmd, m)
	case PromptSendBlockRequest:
		m := &messaging.DebugMessage{
			Action:  messaging.DebugMessageAction_DEBUG_CONSENSUS_SEND_BLOCK_REQ,
			Type:    messaging.DebugMessageRoutingType_DEBUG_MESSAGE_TYPE_BROADCAST,
			Message: nil,
		}
		broadcastDebugMessage(cmd, m)
	default:
		logger.Global.Error().Msg("Selection not yet implemented...")
	}
}

// Broadcast to the entire network.
func broadcastDebugMessage(cmd *cobra.Command, debugMsg *messaging.DebugMessage) {
	anyProto, err := anypb.New(debugMsg)
	if err != nil {
		logger.Global.Fatal().Err(err).Msg("Failed to create Any proto")
	}

	bus, err := helpers.GetBusFromCmd(cmd)
	if err != nil {
		logger.Global.Fatal().Err(err).Msg("Failed to retrieve bus from command")
	}
	if err := bus.GetP2PModule().Broadcast(anyProto); err != nil {
		logger.Global.Error().Err(err).Msg("Failed to broadcast debug message")
	}
}

// Send to just a single (i.e. first) validator in the set
func sendDebugMessage(cmd *cobra.Command, debugMsg *messaging.DebugMessage) {
	anyProto, err := anypb.New(debugMsg)
	if err != nil {
		logger.Global.Error().Err(err).Msg("Failed to create Any proto")
	}

	pstore, err := helpers.FetchPeerstore(cmd)
	if err != nil {
		logger.Global.Fatal().Err(err).Msg("Unable to retrieve the pstore")
	}

	if pstore.Size() == 0 {
		logger.Global.Fatal().Msg("No validators found")
	}

	// if the message needs to be broadcast, it'll be handled by the business logic of the message handler
	//
	// TODO(#936): The statement above is false. Using `#Send()` will only
	// be unicast with no opportunity for further propagation.
	firstStakedActorAddress := pstore.GetPeerList()[0].GetAddress()
	if err != nil {
		logger.Global.Fatal().Err(err).Msg("Failed to convert validator address into pocketCrypto.Address")
	}

	bus, err := helpers.GetBusFromCmd(cmd)
	if err != nil {
		logger.Global.Fatal().Err(err).Msg("Failed to retrieve bus from command")
	}
	if err := bus.GetP2PModule().Send(firstStakedActorAddress, anyProto); err != nil {
		logger.Global.Error().Err(err).Msg("Failed to send debug message")
	}
}

// fetchPeerstore retrieves the providers from the CLI context and uses them to retrieve the address book for the current height
func fetchPeerstore(cmd *cobra.Command) (typesP2P.Peerstore, error) {
	bus, ok := helpers.GetValueFromCLIContext[modules.Bus](cmd, helpers.BusCLICtxKey)
	if !ok || bus == nil {
		return nil, errors.New("retrieving bus from CLI context")
	}
	// TECHDEBT(#810, #811): use `bus.GetPeerstoreProvider()` after peerstore provider
	// is retrievable as a proper submodule
	pstoreProvider, err := bus.GetModulesRegistry().GetModule(peerstore_provider.PeerstoreProviderSubmoduleName)
	if err != nil {
		return nil, errors.New("retrieving peerstore provider")
	}
	currentHeightProvider := bus.GetCurrentHeightProvider()

	height := currentHeightProvider.CurrentHeight()
	pstore, err := pstoreProvider.(peerstore_provider.PeerstoreProvider).GetStakedPeerstoreAtHeight(height)
	if err != nil {
		return nil, fmt.Errorf("retrieving peerstore at height %d", height)
	}
	// Inform the client's main P2P that a the blockchain is at a new height so it can, if needed, update its view of the validator set
	err = sendConsensusNewHeightEventToP2PModule(height, bus)
	if err != nil {
		return nil, errors.New("sending consensus new height event")
	}
	return pstore, nil
}

// sendConsensusNewHeightEventToP2PModule mimicks the consensus module sending a ConsensusNewHeightEvent to the p2p module
// This is necessary because the debug client is not a validator and has no consensus module but it has to update the peerstore
// depending on the changes in the validator set, which is based on the on-chain state.
// TODO(#613): Make the debug client mimic a full node.
func sendConsensusNewHeightEventToP2PModule(height uint64, bus modules.Bus) error {
	newHeightEvent, err := messaging.PackMessage(&messaging.ConsensusNewHeightEvent{Height: height})
	if err != nil {
		logger.Global.Fatal().Err(err).Msg("Failed to pack consensus new height event")
	}
	return bus.GetP2PModule().HandleEvent(newHeightEvent.Content)
}
