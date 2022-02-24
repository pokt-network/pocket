package main

import (
	"encoding/gob"
	"log"
	"os"
	"pocket/consensus"
	"pocket/consensus/dkg"
	"pocket/consensus/leader_election"
	"pocket/consensus/statesync"
	consensus_types "pocket/consensus/types"
	"pocket/p2p/pre_p2p"
	p2p_types "pocket/p2p/pre_p2p/types"
	p2ptypes "pocket/p2p/types"
	"pocket/shared/config"
	"pocket/shared/crypto"

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

const defaultGenesisFile = "build/config/genesis.json"

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
	pk, _ := crypto.GeneratePrivateKey()
	cfg := &config.Config{
		Genesis:    defaultGenesisFile,
		PrivateKey: pk.String(), // Not used
	}

	gob.Register(&consensus.DebugMessage{})
	gob.Register(&consensus.HotstuffMessage{})
	gob.Register(&statesync.StateSyncMessage{})
	gob.Register(&dkg.DKGMessage{})
	gob.Register(&leader_election.LeaderElectionMessage{})
	gob.Register(&consensus.TxWrapperMessage{})

	state := pre_p2p.GetTestState()
	state.LoadStateFromConfig(cfg)

	network := pre_p2p.ConnectToNetwork(state.ValidatorMap)

	log.Println("[CLIENT] Toggling paceMaker into manual mode...")
	handleSelect(PromptOptionTogglePaceMakerManualMode, network, state)

	for {
		selection, err := promptGetInput()
		if err == nil {
			handleSelect(selection, network, state)
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

func handleSelect(selection string, network p2p_types.Network, state *pre_p2p.TestState) {
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
			Action:  consensus.SendTx,
			Payload: NewSendTxBytes(state),
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
	consensusProtoMsg := &consensus_types.Message{
		Data: messageData,
	}

	anyProto, err := anypb.New(consensusProtoMsg)
	if err != nil {
		log.Println("[ERROR] Failed to encode message: ", err)
		return
	}

	networkProtoMsg := &p2ptypes.NetworkMessage{
		Topic: p2ptypes.PocketTopic_CONSENSUS,
		Data:  anyProto,
	}
	log.Println("Sending a network message with topic", p2ptypes.PocketTopic_CONSENSUS)

	bytes, err := proto.Marshal(networkProtoMsg)
	if err != nil {
		log.Println("[ERROR] Failed to encode message: ", err)
		return
	}

	network.NetworkBroadcast(bytes, 0)
}
