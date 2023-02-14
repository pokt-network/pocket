# Iteration 3 Demo: End-to-end LocalNet Tx POC <!-- omit in toc -->

## Table of Contents <!-- omit in toc -->

- [Goals](#goals)
  - [Features](#features)
- [Shell #1: Setup LocalNet](#shell-1-setup-localnet)
- [Shell #2: Setup Consensus debugger](#shell-2-setup-consensus-debugger)
- [Shell #3: Inspect the data in the database for node1](#shell-3-inspect-the-data-in-the-database-for-node1)
- [Shell #4: Inspect the data in the database for node3](#shell-4-inspect-the-data-in-the-database-for-node3)
- [Shell #5: Trigger a send transaction from the CLI](#shell-5-trigger-a-send-transaction-from-the-cli)
  - [Available Commands](#available-commands)
    - [Accounts setup](#accounts-setup)
  - [First Transaction](#first-transaction)
  - [Second Transaction](#second-transaction)
- [\[Optional\] Shell #6: See Swagger UI](#optional-shell-6-see-swagger-ui)

## Goals

The first video of this demo can be accessed [here](https://drive.google.com/file/d/1IOrzq-XJP04BJjyqPPpPu873aSfwrnur/view?usp=sharing).

![Demo Goals](https://user-images.githubusercontent.com/1892194/205820691-26e801e4-ff79-4132-a7a1-358860ca2335.png)

### Features

The demo showcases a successful end-to-end transaction that includes the following:

- A LocalNet composed of 4 hard-coded Validators
- A LocalNet that is started from genesis
- Orchestration that is driven by Docker & docker-compose
- A CLI that can be used to submit send transactions that are gossiped throughout the network
- A basic & functional version of HotPOKT for Consensus
- A basic & functional version of RainTree for P2P
- A persistence layer that leverages PostgreSQL, BadgerDB and Celestia's SMT for state commitment and state storage

## Shell #1: Setup LocalNet

```bash
m̶a̶k̶e̶ ̶d̶o̶c̶k̶e̶r̶_̶w̶i̶p̶e̶ # [Optional] Clear everything (takes a long time)
make # show all the commands
make install_cli_deps # install the CLI dependencies
make protogen_local # generate the protobuf files
make generate_rpc_openapi # generate the OpenAPI spec
make docker_wipe_nodes # clear all the 4 validator nodes
make db_drop # clear the existing database
make compose_and_watch # Start 4 validator node LocalNet
```

## Shell #2: Setup Consensus debugger

```bash
make client_start && make client_connect # start the consensus debugger
```

Use `TriggerNextView` and `PrintNodeState` to increment and inspect each node's `height/round/step`.

## Shell #3: Inspect the data in the database for node1

Connect to the SQL DB of node #1:

```bash
make db_show_schemas # show that there are 4 node schemas
make db_cli_node # connect to the default node 1
```

Query the blocks, accounts and pools from the DB:

```sql
show search_path;
select height, hash from block;
select * from account;
select * from pool;
```

## Shell #4: Inspect the data in the database for node3

```bash
psqlSchema=node3 make db_cli_node # connect to node 3
```

Query the blocks, accounts and pools from the DB:

```sql
show search_path;
select height, hash from block;
select * from account;
select * from pool;
```

## Shell #5: Trigger a send transaction from the CLI

### Available Commands

Show all the commands available in the CLI by running `p1` or:

```bash
go run app/client/*.go
```

#### Accounts setup

Since our Keybase is under development, we have to manually inject the private keys of the accounts we want to use in the CLI.

For the following steps, you'll need to use the accounts of the first two validators in the hard-coded development genesis file. Therefore you have some options:

1. You can just:

```bash
echo '"4ff3292ff14213149446f8208942b35439cb4b2c5e819f41fb612e880b5614bdd6cea8706f6ee6672c1e013e667ec8c46231e0e7abcf97ba35d89fceb8edae45"' > /tmp/val1.json

echo '"25b385b367a827eaafcdb1003bd17a25f2ecc0d10d41f138846f52ae1015aa941041a9c76539791fef9bee5b4fcd5bf4a1a489e0790c44cbdfa776b901e13b50"' > /tmp/val2.json
```

2. You can use `jq` and run these commands:

```bash
cat ./build/config/config1.json | jq '.private_key' > /tmp/val1.json
cat ./build/config/config2.json | jq '.private_key' > /tmp/val2.json
```

3. You can manually copy-paste the private keys from the config files into the `/tmp/val1.json` and `/tmp/val2.json` files. Remember to keep the double quotes around the private keys ("private_key" field in the JSON).

### First Transaction

Trigger a send transaction from validator 1 to validator 2.

```bash
go run app/client/*.go --path_to_private_key_file=/tmp/val1.json Account Send 00404a570febd061274f72b50d0a37f611dfe339 00304d0101847b37fd62e7bebfbdddecdbb7133e 1000
```

1. Use shell #2 to `TriggerNextView` and confirm height increased via `PrintNodeState`
   - You may need to do this more than once in case there's a bug.
2. Use shell #3 to inspect how the balances changes
   - You should see new records with the height `1`
   - You should see that the `DAO` got some money
   - You should see that funds were moved from one account to another
3. Use shell #4 to inspect how the balances changes
   - You should see the same data as above
4. Use shell #2 to `ShowLatestBlockInStore`
   - You should see the data for the block at height `1`

### Second Transaction

Trigger a send transaction from validator 2 to validator 1.

```bash
go run app/client/*.go --path_to_private_key_file=/tmp/val2.json Account Send 00304d0101847b37fd62e7bebfbdddecdbb7133e 00404a570febd061274f72b50d0a37f611dfe339 1000
```

1. Use shell #2 to `TriggerNextView` (one or more times) and confirm height increased via `PrintNodeState`
   - You may need to do this more than once in case there's a bug.
2. Use shell #3 to inspect how the balances changes
   - You should see new records with the height `2`
   - You should see that the `DAO` got some money
   - You should see that funds were moved from one account to another
3. Use shell #4 to inspect how the balances changes
   - You should see the same data as above
4. Use shell #2 to `ShowLatestBlockInStore`
   - You should see the data for the block at height `2`

## [Optional] Shell #6: See Swagger UI

```bash
make swagger-ui
```

<!-- GITHUB_WIKI: guides/demos/iteration_3_end_to_end_tx_poc -->
