package main

// TODO(team): discuss & design the long-term solution to this client.

import (
	"log"
	"os"

	"github.com/manifoldco/promptui"
	"github.com/pokt-network/pocket/p2p/pre2p"
	"github.com/pokt-network/pocket/shared/config"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/types"
	typesGenesis "github.com/pokt-network/pocket/shared/types/genesis"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	PromptResetToGenesis      string = "ResetToGenesis"
	PromptPrintNodeState      string = "PrintNodeState"
	PromptTriggerNextView     string = "TriggerNextView"
	PromptTogglePacemakerMode string = "TogglePacemakerMode"
)

var items = []string{
	PromptResetToGenesis,
	PromptPrintNodeState,
	PromptTriggerNextView,
	PromptTogglePacemakerMode,
}

// A P2P module is initialized in order to broadcast a message to the local network
var pre2pMod modules.P2PModule

func main() {
	pk, err := crypto.GeneratePrivateKey()
	if err != nil {
		log.Fatalf("[ERROR] Failed to generate private key: %v", err)
	}

	cfg := &config.Config{
		Genesis: "build/config/genesis.json",

		// Not used - only set to avoid `GetNodeState(_)` from crashing
		PrivateKey: pk.(crypto.Ed25519PrivateKey),

		// Not used - only set to avoid `pre2p.Create()` from crashing
		Pre2P: &config.Pre2PConfig{
			ConsensusPort: 9999,
			UseRainTree:   true,
		},
	}

	// Initialize the state singleton
	_ = typesGenesis.GetNodeState(cfg)

	pre2pMod, err = pre2p.Create(cfg)
	if err != nil {
		log.Fatalf("[ERROR] Failed to create pre2p module: " + err.Error())
	}

	for {
		selection, err := promptGetInput()
		if err == nil {
			handleSelect(selection)
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
		log.Printf("Prompt failed %v\n", err)
		return "", err
	}

	return result, nil
}

func handleSelect(selection string) {
	switch selection {
	case PromptResetToGenesis:
		m := &types.DebugMessage{
			Action:  types.DebugMessageAction_DEBUG_CONSENSUS_RESET_TO_GENESIS,
			Message: nil,
		}
		broadcastDebugMessage(m)
	case PromptPrintNodeState:
		m := &types.DebugMessage{
			Action:  types.DebugMessageAction_DEBUG_CONSENSUS_PRINT_NODE_STATE,
			Message: nil,
		}
		broadcastDebugMessage(m)
	case PromptTriggerNextView:
		m := &types.DebugMessage{
			Action:  types.DebugMessageAction_DEBUG_CONSENSUS_TRIGGER_NEXT_VIEW,
			Message: nil,
		}
		broadcastDebugMessage(m)
	case PromptTogglePacemakerMode:
		m := &types.DebugMessage{
			Action:  types.DebugMessageAction_DEBUG_CONSENSUS_TOGGLE_PACE_MAKER_MODE,
			Message: nil,
		}
		broadcastDebugMessage(m)
	default:
		log.Println("Selection not yet implemented...", selection)
	}
}

func broadcastDebugMessage(debugMsg *types.DebugMessage) {
	anyProto, err := anypb.New(debugMsg)
	if err != nil {
		log.Fatalf("[ERROR] Failed to create Any proto: %v", err)
	}

	pre2pMod.Broadcast(anyProto, types.PocketTopic_DEBUG_TOPIC)
}
