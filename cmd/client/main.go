package main

import (
	"fmt"
	"log"
	"os"

	"github.com/manifoldco/promptui"
)

const (
	PromptOptionTriggerNextView           string = "TriggerNextView"
	PromptOptionResetToGenesis            string = "ResetToGenesis"
	PromptOptionPrintNodeState            string = "PrintNodeState"
	PromptOptionSendTx                    string = "SendTx"
	PromptOptionTogglePaceMakerManualMode string = "TogglePaceMakerManualMode"
	PromptOptionTriggerDKG                string = "TriggerDKG"
)

var items = []string{
	PromptOptionTriggerNextView,
	PromptOptionTriggerDKG,
	PromptOptionTogglePaceMakerManualMode,
	PromptOptionSendTx,
	PromptOptionResetToGenesis,
	PromptOptionPrintNodeState,
}

func main() {
	for {
		selection, err := promptGetInput()
		if err == nil {
			fmt.Println("Selection not yet implemented...", selection)
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
