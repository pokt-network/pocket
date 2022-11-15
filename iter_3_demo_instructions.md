# Iteration 3 Demo

## Setup LocalNet (shell 1)

```
$̶ ̶m̶a̶k̶e̶ ̶d̶o̶c̶k̶e̶r̶_̶w̶i̶p̶e̶ // Clear everything; takes a long time
$ make // show all the commands
$ make docker_wipe_nodes // clear all 4 validator nodes
$ make db_drop // clear the existing database
$ make compose_and_watch // Start 4 node LocalNet environment
```

## Setup LocalNet debugger (shell 2)

```
$ make client_start && make client_connect // start the debugger
$ PrintNodeState
```

## Inspect the data in the database (shell 3)

```
$ make db_show_schemas // show 4 nodes
$ make db_cli_node // connect to the default node 1
```

## Inspect the data in the database in another node (shell 4)

```
$ psqlSchema=node3 make db_cli_node // connect to node 3
```

Trigger command via client (shell 5)

```
$ go run app/client/_.go // show all the commands
$ go run app/client/_.go --path_to_private_key_file=/Users/olshansky/workspace/pocket/pocket/pkeys/node1.json Account Send 6f66574e1f50f0ef72dff748c3f11b9e0e89d32a 67eb3f0a50ae459fecf666be0e93176e92441317 1000
```

## What corners did we cut for this demo?

**A lot.**

Persistence - state hash implementation & design

- Not merged to main
- May have big changes incoming as a result of state sync

CLI

- the source for the private key

Keybase

- Not implemented yet (hardcoded in the configs & genesis)

Infra

- K8s operator not merged to main yet

Trust vs proof

- Tooling to prove to you (user/audience) that integrity is maintained DNE yet
- Still heavily dependent on trust that things work
  - Requires more tooling
  - Requires more testing

Improving Demos:

- https://github.com/pokt-network/pocket/issues/349
- https://github.com/pokt-network/pocket/issues/350
