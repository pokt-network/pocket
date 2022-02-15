package main

import (
	"encoding/gob"
	"log"
	"os"

	"pocket/consensus/pkg/config"
	"pocket/consensus/pkg/consensus"
	"pocket/consensus/pkg/consensus/dkg"
	"pocket/consensus/pkg/consensus/leader_election"
	"pocket/consensus/pkg/consensus/statesync"
	consensus_types "pocket/consensus/pkg/consensus/types"
	"pocket/consensus/pkg/types"
	p2p "pocket/p2p"
	"pocket/shared"
	"pocket/shared/messages"
	"pocket/shared/modules"

	"github.com/manifoldco/promptui"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	PromptOptionTriggerNextView           string = "TriggerNextView"
	PromptOptionTriggerDKG                string = "TriggerDKG"
	PromptOptionTogglePaceMakerManualMode string = "TogglePaceMakerManualMode"
	PromptOptionResetToGenesis            string = "ResetToGenesis"
	PromptOptionPrintNodeState            string = "PrintNodeState"
	PromptOptionDumpToNeo4j               string = "DumpToNeo4j"
)

const defaultGenesisFile = "config/genesis.json"

var items = []string{
	PromptOptionTriggerNextView,
	PromptOptionTriggerDKG,
	PromptOptionTogglePaceMakerManualMode,
	PromptOptionResetToGenesis,
	PromptOptionPrintNodeState,
	PromptOptionDumpToNeo4j,
}

func main() {
	cfg := &config.Config{
		Genesis:    defaultGenesisFile,
		PrivateKey: types.GeneratePrivateKey(uint32(0)), // Not used
		P2P: &config.P2PConfig{
			Protocol:   "tcp",
			Address:    "0.0.0.0:12345",
			ExternalIp: "0.0.0.0:12345",
			Peers: []string{
				"0.0.0.0:8081",
				"0.0.0.0:8082",
				"0.0.0.0:8083",
				"0.0.0.0:8084",
			},
		},
	}

	gob.Register(&consensus.DebugMessage{})
	gob.Register(&consensus.HotstuffMessage{})
	gob.Register(&statesync.StateSyncMessage{})
	gob.Register(&dkg.DKGMessage{})
	gob.Register(&leader_election.LeaderElectionMessage{})

	state := shared.GetPocketState()
	state.LoadStateFromConfig(cfg)

	p2pmod, err := p2p.Create(cfg)

	go p2pmod.Start(nil)

	if err != nil {
		panic(err)
	}

	log.Println("[CLIENT] Toggling paceMaker into manual mode...")
	handleSelect(PromptOptionTogglePaceMakerManualMode, p2pmod)

	for {
		selection, err := promptGetInput()
		if err == nil {
			handleSelect(selection, p2pmod)
		}
	}
}

func promptGetInput() (string, error) {
	prompt := promptui.Select{
		Label: "Select an option",
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

func handleSelect(selection string, p2pmod modules.NetworkModule) {
	switch selection {
	case PromptOptionTriggerNextView:
		log.Println("[CLIENT] Broadcasting TriggerNextView...")
		m := &consensus.DebugMessage{
			Action: consensus.TriggerNextView,
		}
		broadcastMessage(m, p2pmod)
	case PromptOptionTriggerDKG:
		log.Println("[CLIENT] Broadcasting DKG...")
		m := &consensus.DebugMessage{
			Action: consensus.TriggerDKG,
		}
		broadcastMessage(m, p2pmod)
	case PromptOptionTogglePaceMakerManualMode:
		log.Println("[CLIENT] Broadcasting Toggle PaceMaker...")
		m := &consensus.DebugMessage{
			Action: consensus.TogglePaceMakerManualMode,
		}
		broadcastMessage(m, p2pmod)
	case PromptOptionResetToGenesis:
		log.Println("[CLIENT] Broadcasting ResetToGenesis...")
		m := &consensus.DebugMessage{
			Action: consensus.ResetToGenesis,
		}
		broadcastMessage(m, p2pmod)
	case PromptOptionPrintNodeState:
		log.Println("[CLIENT] Broadcasting PrintNodeState...")
		m := &consensus.DebugMessage{
			Action: consensus.PrintNodeState,
		}
		broadcastMessage(m, p2pmod)
	case PromptOptionDumpToNeo4j:
		log.Println("[CLIENT] Dumping to Neo4j...")
		DumpToNeo4j(p2pmod)
	default:
		log.Println("Invalid selection")
	}
}

func broadcastMessage(m consensus_types.GenericConsensusMessage, p2pmod modules.NetworkModule) {
	message := &consensus_types.ConsensusMessage{
		Message: m,
		Sender:  0,
	}
	messageData, err := consensus_types.EncodeConsensusMessage(message)
	if err != nil {
		log.Println("[ERROR] Failed to encode message: ", err)
		return
	}
	consensusProtoMsg := &messages.ConsensusMessage{
		Data: messageData,
	}

	anyProto, err := anypb.New(consensusProtoMsg)
	if err != nil {
		log.Println("[ERROR] Failed to encode message: ", err)
		return
	}

	p2pmsg := &messages.NetworkMessage{
		Nonce: 0,
		Level: 0,
		Topic: messages.PocketTopic_CONSENSUS,
		Data:  anyProto,
	}

	p2pmod.BroadcastMessage(p2pmsg)
}
