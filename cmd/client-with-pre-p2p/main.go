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
	p2p "pocket/prep2p"
	p2p_types "pocket/prep2p/pre_p2p_types"
	"pocket/shared"
	"pocket/shared/messages"

	"github.com/manifoldco/promptui"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	PromptOptionTriggerNextView           string = "TriggerNextView"
	PromptOptionResetToGenesis            string = "ResetToGenesis"
	PromptOptionPrintNodeState            string = "PrintNodeState"
	PromptOptionSendTx                    string = "SendTx"
	PromptOptionTogglePaceMakerManualMode string = "TogglePaceMakerManualMode"
	PromptOptionTriggerDKG                string = "TriggerDKG"
	PromptOptionDumpToNeo4j               string = "DumpToNeo4j"
)

const defaultGenesisFile = "config/genesis.json"

var items = []string{
	PromptOptionTriggerNextView,
	PromptOptionTriggerDKG,
	PromptOptionTogglePaceMakerManualMode,
	PromptOptionSendTx,
	PromptOptionResetToGenesis,
	PromptOptionPrintNodeState,
	PromptOptionDumpToNeo4j,
}

func main() {
	cfg := &config.Config{
		Genesis:    defaultGenesisFile,
		PrivateKey: types.GeneratePrivateKey(uint32(0)), // Not used
	}

	gob.Register(&consensus.DebugMessage{})
	gob.Register(&consensus.HotstuffMessage{})
	gob.Register(&statesync.StateSyncMessage{})
	gob.Register(&dkg.DKGMessage{})
	gob.Register(&leader_election.LeaderElectionMessage{})
	gob.Register(&consensus.TxWrapperMessage{})

	state := shared.GetPocketState()
	state.LoadStateFromConfig(cfg)

	network := p2p.ConnectToNetwork(state.ValidatorMap)

	log.Println("[CLIENT] Toggling paceMaker into manual mode...")
	handleSelect(PromptOptionTogglePaceMakerManualMode, network)

	for {
		selection, err := promptGetInput()
		if err == nil {
			handleSelect(selection, network)
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

func handleSelect(selection string, network p2p_types.Network) {
	switch selection {
	case PromptOptionTriggerNextView:
		log.Println("[CLIENT] Broadcasting TriggerNextView...")
		m := &consensus.DebugMessage{
			Action: consensus.TriggerNextView,
		}
		broadcastMessage(m, network)
	case PromptOptionSendTx:
		log.Println("[CLIENT] Trigger a SendTx...")
		m := &consensus.DebugMessage{
			Action: consensus.SendTx,
		}
		broadcastMessage(m, network)
	case PromptOptionTriggerDKG:
		log.Println("[CLIENT] Broadcasting DKG...")
		m := &consensus.DebugMessage{
			Action: consensus.TriggerDKG,
		}
		broadcastMessage(m, network)
	case PromptOptionTogglePaceMakerManualMode:
		log.Println("[CLIENT] Broadcasting Toggle PaceMaker...")
		m := &consensus.DebugMessage{
			Action: consensus.TogglePaceMakerManualMode,
		}
		broadcastMessage(m, network)
	case PromptOptionResetToGenesis:
		log.Println("[CLIENT] Broadcasting ResetToGenesis...")
		m := &consensus.DebugMessage{
			Action: consensus.ResetToGenesis,
		}
		broadcastMessage(m, network)
	case PromptOptionPrintNodeState:
		log.Println("[CLIENT] Broadcasting PrintNodeState...")
		m := &consensus.DebugMessage{
			Action: consensus.PrintNodeState,
		}
		broadcastMessage(m, network)
	case PromptOptionDumpToNeo4j:
		log.Println("[CLIENT] Dumping to Neo4j...")
		DumpToNeo4j(network)
	default:
		log.Println("Invalid selection")
	}
}

func broadcastMessage(m consensus_types.GenericConsensusMessage, network p2p_types.Network) {
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

	networkProtoMsg := &messages.NetworkMessage{
		Topic: messages.PocketTopic_CONSENSUS.String(),
		Data:  anyProto,
	}

	bytes, err := proto.Marshal(networkProtoMsg)
	if err != nil {
		log.Println("[ERROR] Failed to encode message: ", err)
		return
	}

	network.NetworkBroadcast(bytes, 0)
}
