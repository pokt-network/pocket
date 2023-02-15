# CLI

The CLI is meant to be an user but also a machine friendly way for interacting with Pocket Network.

The spirit of the documentation is to continuously update and inform the reader of the general scope of the node binary as breaking, rapid development occurs.

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

<!-- GITHUB_WIKI: app/client/readme -->
