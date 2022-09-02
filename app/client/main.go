package main

// TODO(team): discuss & design the long-term solution to this client.

import (
	"encoding/json"
	"github.com/pokt-network/pocket/shared/debug"
	"github.com/pokt-network/pocket/shared/test_artifacts"
	"github.com/pokt-network/pocket/telemetry"
	"log"
	"os"

	"github.com/manifoldco/promptui"
	"github.com/pokt-network/pocket/consensus"
	"github.com/pokt-network/pocket/p2p"
	"github.com/pokt-network/pocket/shared"
	pocketCrypto "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	PromptResetToGenesis         string = "ResetToGenesis"
	PromptPrintNodeState         string = "PrintNodeState"
	PromptTriggerNextView        string = "TriggerNextView"
	PromptTogglePacemakerMode    string = "TogglePacemakerMode"
	PromptShowLatestBlockInStore string = "ShowLatestBlockInStore"
)

var items = []string{
	PromptResetToGenesis,
	PromptPrintNodeState,
	PromptTriggerNextView,
	PromptTogglePacemakerMode,
	PromptShowLatestBlockInStore,
}

// A P2P module is initialized in order to broadcast a message to the local network
var p2pMod modules.P2PModule
var consensusMod modules.ConsensusModule

func main() {
	var err error
	config, genesis := test_artifacts.ReadConfigAndGenesisFiles("", "")
	config, err = injectClientPrivateKey(config)
	if err != nil {
		log.Fatalf("[ERROR] Failed to inject a client private key into p2p and consensus module: %v", err.Error())
	}
	consensusMod, err = consensus.Create(config["consensus"], genesis["consensusGenesisState"])
	if err != nil {
		log.Fatalf("[ERROR] Failed to create consensus module: %v", err.Error())
	}
	p2pMod, err = p2p.Create(config["p2p"], genesis["p2PGenesisState"])
	if err != nil {
		log.Fatalf("[ERROR] Failed to create p2p module: %v", err.Error())
	}
	// This telemetry module instance is a NOOP because the 'enable_telemetry' flag in the `cfg` above is set to false.
	// Since this client mimics partial - networking only - functionality of a full node, some of the telemetry-related
	// code paths are executed. To avoid those messages interfering with the telemetry data collected, a non-nil telemetry
	// module that NOOPs (per the configs above) is injected.
	telemetryMod, err := telemetry.Create(config["telemetry"], genesis["telemetryGenesisState"])
	if err != nil {
		log.Fatalf("[ERROR] Failed to create NOOP telemetry module: " + err.Error())
	}

	_ = shared.CreateBusWithOptionalModules(nil, p2pMod, nil, consensusMod, telemetryMod, config, genesis)

	p2pMod.Start()

	for {
		selection, err := promptGetInput()
		if err == nil {
			handleSelect(selection)
		}
	}
}

// inject a random private key so the client may send messages without rain-tree rejecting it as a 'self message'
func injectClientPrivateKey(config map[string]json.RawMessage) (map[string]json.RawMessage, error) {
	pk, err := pocketCrypto.GeneratePrivateKey()
	if err != nil {
		return nil, err
	}
	pkString := pk.String()

	mockConsensusConfig := test_artifacts.MockConsensusConfig{}
	mockP2PConfig := test_artifacts.MockP2PConfig{}
	if err := json.Unmarshal(config["consensus"], &mockConsensusConfig); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(config["p2p"], &mockP2PConfig); err != nil {
		return nil, err
	}
	mockConsensusConfig.PrivateKey = pkString
	mockP2PConfig.PrivateKey = pkString
	if config["consensus"], err = json.Marshal(mockConsensusConfig); err != nil {
		return nil, err
	}
	if config["p2p"], err = json.Marshal(mockP2PConfig); err != nil {
		return nil, err
	}
	return config, nil
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
		m := &debug.DebugMessage{
			Action:  debug.DebugMessageAction_DEBUG_CONSENSUS_RESET_TO_GENESIS,
			Message: nil,
		}
		broadcastDebugMessage(m)
	case PromptPrintNodeState:
		m := &debug.DebugMessage{
			Action:  debug.DebugMessageAction_DEBUG_CONSENSUS_PRINT_NODE_STATE,
			Message: nil,
		}
		broadcastDebugMessage(m)
	case PromptTriggerNextView:
		m := &debug.DebugMessage{
			Action:  debug.DebugMessageAction_DEBUG_CONSENSUS_TRIGGER_NEXT_VIEW,
			Message: nil,
		}
		broadcastDebugMessage(m)
	case PromptTogglePacemakerMode:
		m := &debug.DebugMessage{
			Action:  debug.DebugMessageAction_DEBUG_CONSENSUS_TOGGLE_PACE_MAKER_MODE,
			Message: nil,
		}
		broadcastDebugMessage(m)
	case PromptShowLatestBlockInStore:
		m := &debug.DebugMessage{
			Action:  debug.DebugMessageAction_DEBUG_SHOW_LATEST_BLOCK_IN_STORE,
			Message: nil,
		}
		sendDebugMessage(m)
	default:
		log.Println("Selection not yet implemented...", selection)
	}
}

// Broadcast to the entire validator set
func broadcastDebugMessage(debugMsg *debug.DebugMessage) {
	anyProto, err := anypb.New(debugMsg)
	if err != nil {
		log.Fatalf("[ERROR] Failed to create Any proto: %v", err)
	}

	// TODO(olshansky): Once we implement the cleanup layer in RainTree, we'll be able to use
	// broadcast. The reason it cannot be done right now is because this client is not in the
	// address book of the actual validator nodes, so `node1.consensus` never receives the message.
	// p2pMod.Broadcast(anyProto, types.PocketTopic_DEBUG_TOPIC)

	for _, val := range consensusMod.ValidatorMap() {
		addr, err := pocketCrypto.NewAddress(val.GetAddress())
		if err != nil {
			log.Fatalf("[ERROR] Failed to convert validator address into pocketCrypto.Address: %v", err)
		}
		p2pMod.Send(addr, anyProto, debug.PocketTopic_DEBUG_TOPIC)
	}
}

// Send to just a single (i.e. first) validator in the set
func sendDebugMessage(debugMsg *debug.DebugMessage) {
	anyProto, err := anypb.New(debugMsg)
	if err != nil {
		log.Fatalf("[ERROR] Failed to create Any proto: %v", err)
	}

	var validatorAddress []byte
	for _, val := range consensusMod.ValidatorMap() {
		validatorAddress, err = pocketCrypto.NewAddress(val.GetAddress())
		if err != nil {
			log.Fatalf("[ERROR] Failed to convert validator address into pocketCrypto.Address: %v", err)
		}
		break
	}

	p2pMod.Send(validatorAddress, anyProto, debug.PocketTopic_DEBUG_TOPIC)
}
