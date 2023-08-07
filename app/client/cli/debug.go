package cli

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/pokt-network/pocket/app/client/cli/helpers"
	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/shared/messaging"
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
	dbg := newDebugCommand()
	dbg.AddCommand(newDebugSubCommands()...)
	rootCmd.AddCommand(dbg)

	dbgUI := newDebugUICommand()
	rootCmd.AddCommand(dbgUI)
}

// newDebugCommand returns the cobra CLI for the Debug command.
func newDebugCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "Debug",
		Aliases: []string{"d"},
		Short:   "Debug utility for rapid development",
		Long:    "Debug utility to send fire-and-forget messages to the network for development purposes",
		Args:    cobra.MaximumNArgs(1),
	}
}

// newDebugSubCommands is a list of commands that can be "fired & forgotten" (no selection necessary)
func newDebugSubCommands() []*cobra.Command {
	cmds := []*cobra.Command{
		{
			Use:               "PrintNodeState",
			Aliases:           []string{"print", "state"},
			Short:             "Prints the node state",
			Long:              "Sends a message to all visible nodes to log the current state of their consensus",
			Args:              cobra.ExactArgs(0),
			PersistentPreRunE: helpers.P2PDependenciesPreRunE,
			Run: func(cmd *cobra.Command, args []string) {
				runWithSleep(func() {
					handleSelect(cmd, PromptPrintNodeState)
				})
			},
		},
		{
			Use:               "ResetToGenesis",
			Aliases:           []string{"reset", "genesis"},
			Short:             "Reset to genesis",
			Long:              "Broadcast a message to all visible nodes to reset the state to genesis",
			Args:              cobra.ExactArgs(0),
			PersistentPreRunE: helpers.P2PDependenciesPreRunE,
			Run: func(cmd *cobra.Command, args []string) {
				runWithSleep(func() {
					handleSelect(cmd, PromptResetToGenesis)
				})
			},
		},
		{
			Use:               "TriggerView",
			Aliases:           []string{"next", "trigger", "view"},
			Short:             "Trigger the next view in consensus",
			Long:              "Sends a message to all visible nodes on the network to start the next view (height/step/round) in consensus",
			Args:              cobra.ExactArgs(0),
			PersistentPreRunE: helpers.P2PDependenciesPreRunE,
			Run: func(cmd *cobra.Command, args []string) {
				runWithSleep(func() {
					handleSelect(cmd, PromptTriggerNextView)
				})
			},
		},
		{
			Use:               "TogglePacemakerMode",
			Aliases:           []string{"toggle", "pcm"},
			Short:             "Toggle the pacemaker",
			Long:              "Toggle the consensus pacemaker either on or off so the chain progresses on its own or loses liveness",
			Args:              cobra.ExactArgs(0),
			PersistentPreRunE: helpers.P2PDependenciesPreRunE,
			Run: func(cmd *cobra.Command, args []string) {
				runWithSleep(func() {
					handleSelect(cmd, PromptTogglePacemakerMode)
				})
			},
		},
		{
			Use:               "ScaleActor",
			Aliases:           []string{"scale"},
			Short:             "Scales the number of actors up or down",
			Long:              "Scales the type of actor specified to the number provided",
			Args:              cobra.ExactArgs(2),
			PersistentPreRunE: helpers.P2PDependenciesPreRunE,
			Run: func(cmd *cobra.Command, args []string) {
				actor := args[0]
				numActors := args[1]
				validActors := []string{"fishermen", "full_nodes", "servicers", "validators"}
				if !slices.Contains(validActors, actor) {
					logger.Global.Fatal().Msg("Invalid actor type provided")
				}
				sedReplaceCmd := fmt.Sprintf("/%s:/,/count:/ s/count: [0-9]*/count: %s/", actor, numActors)
				sedCmd := exec.Command("sed", "-i", sedReplaceCmd, "/usr/local/localnet_config.yaml")
				if err := sedCmd.Run(); err != nil {
					log.Fatal(err)
				}
			},
		},
	}
	return cmds
}

// newDebugUICommand returns the cobra CLI for the Debug UI interface.
func newDebugUICommand() *cobra.Command {
	return &cobra.Command{
		Aliases:           []string{"dui", "debug"},
		Use:               "DebugUI",
		Short:             "Debug utility with an interactive UI for development purposes",
		Long:              "Opens a shell-driven selection UI to view and select from a list of debug actions for development purposes",
		Args:              cobra.MaximumNArgs(0),
		PersistentPreRunE: helpers.P2PDependenciesPreRunE,
		RunE:              selectDebugCommand,
	}
}

// selectDebugCommand builds out the list of debug subcommands by matching the
// handleSelect dispatch to the appropriate command.
//   - To add a debug subcommand, you must add it to the `items` array and then
//     write a function handler to match for it in `handleSelect`.
func selectDebugCommand(cmd *cobra.Command, _ []string) error {
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
		logger.Global.Error().Str("selection", selection).Msg("Selection not yet implemented...")
	}
}

// HACK: Because of how the p2p module works, we need to surround it with sleep both BEFORE and AFTER the task.
// - Starting the task too early after the debug client initializes results in a lack of visibility of the nodes in the network
// - Ending the task too early before the debug client completes its task results in a lack of propagation of the message or retrieval of the result
// TECHDEBT: There is likely an event based solution to this but it would require a lot more refactoring of the p2p module.
func runWithSleep(task func()) {
	time.Sleep(1000 * time.Millisecond)
	task()
	time.Sleep(1000 * time.Millisecond)
}

// broadcastDebugMessage broadcasts the debug message to the entire visible network.
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

// sendDebugMessage sends the debug message to just a single (i.e. first) node visible
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
