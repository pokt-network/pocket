package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/pokt-network/pocket/libp2p"
	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/p2p"
	"github.com/pokt-network/pocket/p2p/providers/current_height_provider"
	rpcCHP "github.com/pokt-network/pocket/p2p/providers/current_height_provider/rpc"
	"github.com/pokt-network/pocket/p2p/providers/peerstore_provider"
	rpcABP "github.com/pokt-network/pocket/p2p/providers/peerstore_provider/rpc"
	"github.com/pokt-network/pocket/runtime"
	"github.com/pokt-network/pocket/runtime/defaults"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/shared/modules"
	sharedP2P "github.com/pokt-network/pocket/shared/p2p"
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

// NOTE: this is required by the linter, otherwise a simple string constant would have been enough
type cliContextKey string

const busCLICtxKey = "bus"

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

			bus := runtimeMgr.GetBus()
			modulesRegistry := bus.GetModulesRegistry()

			rpcURL := fmt.Sprintf("http://%s:%s", rpcHost, defaults.DefaultRPCPort)

			addressBookProvider := rpcABP.NewRPCPeerstoreProvider(
				rpcABP.WithP2PConfig(
					runtimeMgr.GetConfig().P2P,
				),
				rpcABP.WithCustomRPCURL(rpcURL),
			)
			modulesRegistry.RegisterModule(addressBookProvider)

			currentHeightProvider := rpcCHP.NewRPCCurrentHeightProvider(
				rpcCHP.WithCustomRPCURL(rpcURL),
			)
			modulesRegistry.RegisterModule(currentHeightProvider)

			setValueInCLIContext(cmd, busCLICtxKey, bus)

			// TECHDEBT: simplify after P2P module consolidation.
			var err error
			p2pMod, err = getP2PModule(runtimeMgr)
			if err != nil {
				logger.Global.Fatal().Err(err).Msg("Failed to create p2p module")
			}

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

	pstore, err := fetchPeerstore(cmd)
	if err != nil {
		logger.Global.Fatal().Msg("Unable to retrieve the pstore")
	}
	for _, val := range pstore.GetPeerList() {
		addr := val.GetAddress()
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

	pstore, err := fetchPeerstore(cmd)
	if err != nil {
		logger.Global.Fatal().Msg("Unable to retrieve the pstore")
	}

	var validatorAddress []byte
	if pstore.Size() == 0 {
		logger.Global.Fatal().Msg("No validators found")
	}

	// if the message needs to be broadcast, it'll be handled by the business logic of the message handler
	validatorAddress = pstore.GetPeerList()[0].GetAddress()
	if err != nil {
		logger.Global.Fatal().Err(err).Msg("Failed to convert validator address into pocketCrypto.Address")
	}

	if err := p2pMod.Send(validatorAddress, anyProto); err != nil {
		logger.Global.Fatal().Err(err).Msg("Failed to send debug message")
	}
}

// fetchPeerstore retrieves the providers from the CLI context and uses them to retrieve the address book for the current height
func fetchPeerstore(cmd *cobra.Command) (sharedP2P.Peerstore, error) {
	bus, ok := getValueFromCLIContext[modules.Bus](cmd, busCLICtxKey)
	if !ok || bus == nil {
		return nil, errors.New("retrieving bus from CLI context")
	}
	modulesRegistry := bus.GetModulesRegistry()
	pstoreProvider, err := modulesRegistry.GetModule(peerstore_provider.ModuleName)
	if err != nil {
		return nil, errors.New("retrieving peerstore provider")
	}
	currentHeightProvider, err := modulesRegistry.GetModule(current_height_provider.ModuleName)
	if err != nil {
		return nil, errors.New("retrieving currentHeightProvider")
	}

	height := currentHeightProvider.(current_height_provider.CurrentHeightProvider).CurrentHeight()
	pstore, err := pstoreProvider.(peerstore_provider.PeerstoreProvider).GetStakedPeerstoreAtHeight(height)
	if err != nil {
		return nil, fmt.Errorf("retrieving peerstore at height %d", height)
	}
	return pstore, err
}

func getP2PModule(runtimeMgr *runtime.Manager) (p2pModule modules.P2PModule, err error) {
	bus := runtimeMgr.GetBus()

	var mod modules.Module
	if runtimeMgr.GetConfig().UseLibP2P {
		mod, err = libp2p.Create(bus)
	} else {
		mod, err = p2p.Create(bus)
	}
	if err != nil {
		return nil, err
	}

	return mod.(modules.P2PModule), nil
}
