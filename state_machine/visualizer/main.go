package main

import (
	"os"

	"github.com/looplab/fsm"
	"github.com/pokt-network/pocket/state_machine"
)

func main() {

	stateMachine := state_machine.NewNodeFSM(nil)

	mermaidStateDiagram, err := fsm.VisualizeForMermaidWithGraphType(stateMachine, fsm.StateDiagram)
	if err != nil {
		panic(err)
	}

	header := "# Node Finite State Machine\n\nThe following diagram displays the various states and events that govern the functionality of the node.\n\n```mermaid\n"
	footer := "```"
	if err := os.WriteFile("state_machine/docs/state-machine.diagram.md", []byte(header+mermaidStateDiagram+footer), 0644); err != nil {
		panic(err)
	}
}
