# CLI

The CLI is meant to be an user but also a machine friendly way for interacting with Pocket Network.

The spirit of the documentation is to continuously update and inform the reader of the general scope of the node binary as breaking, rapid development occurs.

There are two modes of operating the CLI: `Standard` and `Debug`.

- **Standard**: The default mode and is meant for users to interact with the network.
- **Debug**: Intended for developers to interact with the network and debug issues. To enter debug mode, the CLI must be **built** with the `-tags=debug` build tag.

## Commands

Command tree available [here](./commands/client.md)

## Code Organization

```bash
├── cli
│   ├── account.go           # Account subcommand
│   ├── actor.go             # Actor (Application, Node, Fisherman, Validator) subcommands
│   ├── cmd.go               # main (root) command called by the entrypoint
│   ├── debug.go             # Debug subcommand
│   ├── doc
│   │   ├── CHANGELOG.md     # changelog
│   │   ├── README.md        # this file
│   │   └── commands         # commands specific documentation (generated from the commands metadata)
│   ├── docgen
│   │   └── main.go          # commands specific documentation generator
│   ├── gov.go               # Governance subcommand
│   ├── utils.go             # support functions
│   ├── system.go            # System subcommand
│   └── utils_test.go        # tests for the support functions
└── main.go                  # entrypoint
```

## Debug Subcommands

The debug command is terminal utility and set of sub-commands for rapid development and debugging of Pocket validators.
If `debug` is run with no arguments, it drops the user into an interactive prompt session where they can trigger multiple debug message transmissions from the client binary.

```bash
$> client debug
[...]
Use the arrow keys to navigate: ↓ ↑ → ←
? Select an action:
  ▸ PrintNodeState (broadcast)
    TriggerNextView (broadcast)
    TogglePacemakerMode (broadcast)
    ResetToGenesis (broadcast)
    ShowLatestBlockInStore (anycast)
    MetadataRequest (broadcast)
    BlockRequest (broadcast)

```

If it's run with an argument it selects the message type that matches that argument.
The accepted command structure and available sub-commands are shown below.

```bash
Usage:
  client debug [flags]      <-- drops you into a debug command prompt
  client debug [command]    <-- accepts sub commands 

Available Commands:
  BlockRequest
  MetadataRequest
  PrintNodeState
  ResetToGenesis
  ShowLatestBlockInStore
  TogglePacemakerMode
  TriggerNextView
```

This allows Pacemaker mode and other functionality to be toggled externally for testing and performance purposes at runtime and not just configured once at start or build time.

<!-- GITHUB_WIKI: app/client/readme -->
