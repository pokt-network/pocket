package cli

import (
	"os"

	"github.com/manifoldco/promptui"
	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/p2p"
	debugABP "github.com/pokt-network/pocket/p2p/providers/addrbook_provider/debug"
	debugCHP "github.com/pokt-network/pocket/p2p/providers/current_height_provider/debug"
	"github.com/pokt-network/pocket/runtime"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	pocketCrypto "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/anypb"
)

// TECHDEBT: Lowercase variables / constants that do not need to be exported.
const (
	PromptResetToGenesis         string = "ResetToGenesis"
	PromptPrintNodeState         string = "PrintNodeState"
	PromptTriggerNextView        string = "TriggerNextView"
	PromptTogglePacemakerMode    string = "TogglePacemakerMode"
	PromptShowLatestBlockInStore string = "ShowLatestBlockInStore"
)

var (
	// A P2P module is initialized in order to broadcast a message to the local network
	p2pMod modules.P2PModule

	items = []string{
		PromptResetToGenesis,
		PromptPrintNodeState,
		PromptTriggerNextView,
		PromptTogglePacemakerMode,
		PromptShowLatestBlockInStore,
	}

	// validators holds the list of the validators at genesis time so that we can use it to create a debug address book provider.
	// Its purpose is to allow the CLI to "discover" the nodes in the network. Since currently we don't have churn and we run nodes only in LocalNet, we can rely on the genesis state.
	// HACK(#416): This is a temporary solution that guarantees backward compatibility while we implement peer discovery
	validators []*coreTypes.Actor

	configPath  string = getEnv("CONFIG_PATH", "build/config/config1.json")
	genesisPath string = getEnv("GENESIS_PATH", "build/config/genesis.json")
)

func init() {
	debugCmd := NewDebugCommand()
	rootCmd.AddCommand(debugCmd)
}

func NewDebugCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "debug",
		Short: "Debug utility for rapid development",
		Args:  cobra.ExactArgs(0),
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			runtimeMgr := runtime.NewManagerFromFiles(configPath, genesisPath, runtime.WithClientDebugMode(), runtime.WithRandomPK())

			// HACK(#416): this is a temporary solution that guarantees backward compatibility while we implement peer discovery.
			validators = runtimeMgr.GetGenesis().Validators

			debugAddressBookProvider := debugABP.NewDebugAddrBookProvider(
				runtimeMgr.GetConfig().P2P,
				debugABP.WithActorsByHeight(
					map[int64][]*coreTypes.Actor{
						debugABP.ANY_HEIGHT: validators,
					},
				),
			)

			debugCurrentHeightProvider := debugCHP.NewDebugCurrentHeightProvider(0)

			// TODO(#429): refactor injecting the dependencies into the bus so that they can be consumed in an updated `P2PModule.Create()` implementation
			p2pM, err := p2p.CreateWithProviders(runtimeMgr.GetBus(), debugAddressBookProvider, debugCurrentHeightProvider)
			if err != nil {
				logger.Global.Fatal().Err(err).Msg("Failed to create p2p module")
			}
			p2pMod = p2pM.(modules.P2PModule)

			p2pMod.Start()
		},
		RunE: runDebug,
	}
}

func runDebug(cmd *cobra.Command, args []string) (err error) {
	for {
		if selection, err := promptGetInput(); err == nil {
			handleSelect(selection)
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

func handleSelect(selection string) {
	switch selection {
	case PromptResetToGenesis:
		m := &messaging.DebugMessage{
			Action:  messaging.DebugMessageAction_DEBUG_CONSENSUS_RESET_TO_GENESIS,
			Message: nil,
		}
		broadcastDebugMessage(m)
	case PromptPrintNodeState:
		m := &messaging.DebugMessage{
			Action:  messaging.DebugMessageAction_DEBUG_CONSENSUS_PRINT_NODE_STATE,
			Message: nil,
		}
		broadcastDebugMessage(m)
	case PromptTriggerNextView:
		m := &messaging.DebugMessage{
			Action:  messaging.DebugMessageAction_DEBUG_CONSENSUS_TRIGGER_NEXT_VIEW,
			Message: nil,
		}
		broadcastDebugMessage(m)
	case PromptTogglePacemakerMode:
		m := &messaging.DebugMessage{
			Action:  messaging.DebugMessageAction_DEBUG_CONSENSUS_TOGGLE_PACE_MAKER_MODE,
			Message: nil,
		}
		broadcastDebugMessage(m)
	case PromptShowLatestBlockInStore:
		m := &messaging.DebugMessage{
			Action:  messaging.DebugMessageAction_DEBUG_SHOW_LATEST_BLOCK_IN_STORE,
			Message: nil,
		}
		sendDebugMessage(m)
	default:
		logger.Global.Error().Msg("Selection not yet implemented...")
	}
}

// Broadcast to the entire validator set
func broadcastDebugMessage(debugMsg *messaging.DebugMessage) {
	anyProto, err := anypb.New(debugMsg)
	if err != nil {
		logger.Global.Fatal().Err(err).Msg("Failed to create Any proto")
	}

	// TODO(olshansky): Once we implement the cleanup layer in RainTree, we'll be able to use
	// broadcast. The reason it cannot be done right now is because this client is not in the
	// address book of the actual validator nodes, so `node1.consensus` never receives the message.
	// p2pMod.Broadcast(anyProto, messaging.PocketTopic_DEBUG_TOPIC)

	for _, valAddr := range validators {
		addr, err := pocketCrypto.NewAddress(valAddr.GetAddress())
		if err != nil {
			logger.Global.Fatal().Err(err).Msg("Failed to convert validator address into pocketCrypto.Address")
		}
		p2pMod.Send(addr, anyProto)
	}
}

// Send to just a single (i.e. first) validator in the set
func sendDebugMessage(debugMsg *messaging.DebugMessage) {
	anyProto, err := anypb.New(debugMsg)
	if err != nil {
		logger.Global.Error().Err(err).Msg("Failed to create Any proto")
	}

	var validatorAddress []byte
	if len(validators) == 0 {
		logger.Global.Fatal().Msg("No validators found")
	}

	// if the message needs to be broadcast, it'll be handled by the business logic of the message handler
	validatorAddress, err = pocketCrypto.NewAddress(validators[0].GetAddress())
	if err != nil {
		logger.Global.Fatal().Err(err).Msg("Failed to convert validator address into pocketCrypto.Address")
	}

	p2pMod.Send(validatorAddress, anyProto)
}

func getEnv(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}
