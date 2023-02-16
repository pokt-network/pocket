package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/manifoldco/promptui"
	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/p2p"
	"github.com/pokt-network/pocket/p2p/providers/addrbook_provider"
	rpcABP "github.com/pokt-network/pocket/p2p/providers/addrbook_provider/rpc"
	"github.com/pokt-network/pocket/p2p/providers/current_height_provider"
	rpcCHP "github.com/pokt-network/pocket/p2p/providers/current_height_provider/rpc"
	"github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/runtime"
	"github.com/pokt-network/pocket/runtime/defaults"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/anypb"
)

// TECHDEBT: Lowercase variables / constants that do not need to be exported.
const (
	PromptResetToGenesis      string = "ResetToGenesis"
	PromptPrintNodeState      string = "PrintNodeState"
	PromptTriggerNextView     string = "TriggerNextView"
	PromptTogglePacemakerMode string = "TogglePacemakerMode"

	PromptShowLatestBlockInStore string = "ShowLatestBlockInStore"

	PromptSendMetadataRequest string = "MetadataRequest"
	PromptSendBlockRequest    string = "BlockRequest"
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
		PromptSendMetadataRequest,
		PromptSendBlockRequest,
	}

	configPath  string = runtime.GetEnv("CONFIG_PATH", "build/config/config1.json")
	genesisPath string = runtime.GetEnv("GENESIS_PATH", "build/config/genesis.json")
	rpcHost     string
)

type ctxKey int

const (
	addrBookProviderCtxKey ctxKey = iota
	currentHeightProviderCtxKey
)

func init() {
	debugCmd := NewDebugCommand()
	rootCmd.AddCommand(debugCmd)

	// by default, we point at the same endpoint used by the CLI but the debug client is used either in docker-compose of K8S, therefore we cater for overriding
	validator1Endpoint := defaults.Validator1EndpointDockerCompose
	if runtime.IsProcessRunningInsideKubernetes() {
		validator1Endpoint = defaults.Validator1EndpointK8S
	}

	rpcHost = runtime.GetEnv("RPC_HOST", validator1Endpoint)
}

func NewDebugCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "debug",
		Short: "Debug utility for rapid development",
		Args:  cobra.ExactArgs(0),
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			runtimeMgr := runtime.NewManagerFromFiles(
				configPath, genesisPath,
				runtime.WithClientDebugMode(),
				runtime.WithRandomPK(),
			)

			rpcUrl := fmt.Sprintf("http://%s:%s", rpcHost, defaults.DefaultRPCPort)

			modulesRegistry := runtimeMgr.GetBus().GetModulesRegistry()
			addressBookProvider := rpcABP.NewRPCAddrBookProvider(
				rpcABP.WithP2PConfig(
					runtimeMgr.GetConfig().P2P,
				),
				rpcABP.WithCustomRPCUrl(rpcUrl),
			)
			modulesRegistry.RegisterModule(addressBookProvider)
			cmd.SetContext(context.WithValue(cmd.Context(), addrBookProviderCtxKey, addressBookProvider))

			currentHeightProvider := rpcCHP.NewRPCCurrentHeightProvider(
				rpcCHP.WithCustomRPCUrl(rpcUrl),
			)
			modulesRegistry.RegisterModule(currentHeightProvider)
			cmd.SetContext(context.WithValue(cmd.Context(), currentHeightProviderCtxKey, currentHeightProvider))

			p2pM, err := p2p.Create(runtimeMgr.GetBus())
			if err != nil {
				logger.Global.Fatal().Err(err).Msg("Failed to create p2p module")
			}
			p2pMod = p2pM.(modules.P2PModule)

			if err := p2pMod.Start(); err != nil {
				logger.Global.Fatal().Err(err).Msg("Failed to start p2p module")
			}
		},
		RunE: runDebug,
	}
}

func runDebug(cmd *cobra.Command, args []string) (err error) {
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
			Message: nil,
		}
		broadcastDebugMessage(cmd, m)
	case PromptPrintNodeState:
		m := &messaging.DebugMessage{
			Action:  messaging.DebugMessageAction_DEBUG_CONSENSUS_PRINT_NODE_STATE,
			Message: nil,
		}
		broadcastDebugMessage(cmd, m)
	case PromptTriggerNextView:
		m := &messaging.DebugMessage{
			Action:  messaging.DebugMessageAction_DEBUG_CONSENSUS_TRIGGER_NEXT_VIEW,
			Message: nil,
		}
		broadcastDebugMessage(cmd, m)
	case PromptTogglePacemakerMode:
		m := &messaging.DebugMessage{
			Action:  messaging.DebugMessageAction_DEBUG_CONSENSUS_TOGGLE_PACE_MAKER_MODE,
			Message: nil,
		}
		broadcastDebugMessage(cmd, m)
	case PromptShowLatestBlockInStore:
		m := &messaging.DebugMessage{
			Action:  messaging.DebugMessageAction_DEBUG_SHOW_LATEST_BLOCK_IN_STORE,
			Message: nil,
		}
		sendDebugMessage(cmd, m)
	case PromptSendMetadataRequest:
		m := &messaging.DebugMessage{
			Action:  messaging.DebugMessageAction_DEBUG_CONSENSUS_SEND_METADATA_REQ,
			Message: nil,
		}
		broadcastDebugMessage(cmd, m)
	case PromptSendBlockRequest:
		m := &messaging.DebugMessage{
			Action:  messaging.DebugMessageAction_DEBUG_CONSENSUS_SEND_BLOCK_REQ,
			Message: nil,
		}
		broadcastDebugMessage(cmd, m)
	default:
		logger.Global.Error().Msg("Selection not yet implemented...")
	}
}

// Broadcast to the entire validator set
func broadcastDebugMessage(cmd *cobra.Command, debugMsg *messaging.DebugMessage) {
	anyProto, err := anypb.New(debugMsg)
	if err != nil {
		logger.Global.Fatal().Err(err).Msg("Failed to create Any proto")
	}

	// TODO(olshansky): Once we implement the cleanup layer in RainTree, we'll be able to use
	// broadcast. The reason it cannot be done right now is because this client is not in the
	// address book of the actual validator nodes, so `node1.consensus` never receives the message.
	// p2pMod.Broadcast(anyProto)

	addrBook, err := fetchAddressBook(cmd)
	for _, val := range addrBook {
		addr := val.Address
		if err != nil {
			logger.Global.Fatal().Err(err).Msg("Failed to convert validator address into pocketCrypto.Address")
		}
		if err := p2pMod.Send(addr, anyProto); err != nil {
			logger.Global.Fatal().Err(err).Msg("Failed to send debug message")
		}
	}

}

// Send to just a single (i.e. first) validator in the set
func sendDebugMessage(cmd *cobra.Command, debugMsg *messaging.DebugMessage) {
	anyProto, err := anypb.New(debugMsg)
	if err != nil {
		logger.Global.Error().Err(err).Msg("Failed to create Any proto")
	}

	addrBook, err := fetchAddressBook(cmd)
	if err != nil {
		logger.Global.Fatal().Msg("Unable to retrieve the addrBook")
	}

	var validatorAddress []byte
	if len(addrBook) == 0 {
		logger.Global.Fatal().Msg("No validators found")
	}

	// if the message needs to be broadcast, it'll be handled by the business logic of the message handler
	validatorAddress = addrBook[0].Address
	if err != nil {
		logger.Global.Fatal().Err(err).Msg("Failed to convert validator address into pocketCrypto.Address")
	}

	if err := p2pMod.Send(validatorAddress, anyProto); err != nil {
		logger.Global.Fatal().Err(err).Msg("Failed to send debug message")
	}
}

// fetchAddressBook retrieves the providers from the CLI context and uses them to retrieve the address book for the current height
func fetchAddressBook(cmd *cobra.Command) (types.AddrBook, error) {
	addrBookProvider := cmd.Context().Value(addrBookProviderCtxKey)
	currentHeightProvider := cmd.Context().Value(currentHeightProviderCtxKey)

	height := currentHeightProvider.(current_height_provider.CurrentHeightProvider).CurrentHeight()
	addrBook, err := addrBookProvider.(addrbook_provider.AddrBookProvider).GetStakedAddrBookAtHeight(height)
	if err != nil {
		logger.Global.Fatal().Msg("Unable to retrieve the addrBook")
	}
	return addrBook, err
}
