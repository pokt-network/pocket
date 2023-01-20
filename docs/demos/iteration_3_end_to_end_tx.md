# Iteration 3 Demo <!-- omit in toc -->

**Table of Contents**

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

<img width="842" alt="Screenshot 2022-12-05 at 9 02 28 PM" src="https://user-images.githubusercontent.com/1892194/205820691-26e801e4-ff79-4132-a7a1-358860ca2335.png">

### Features

The demo showcases a successful end-to-end transaction that includes the following:

- A LocalNet composed of 4 hard-coded Validators
- A LocalNet that is started from genesis
- Orchestration that is driven by Docker & docker-compose
- A CLI that can be use to submit send transactions that are gossiped throughout the network
- A basic & functional version of HotPOKT for Consensus
- A basic & functional version of RainTree for P2P
- A persistence layer than leverages PostgreSQL, BadgerDB and Celestia's SMT for state commitment and state storage

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

Show all the commands available in the CLI:

```bash
go run app/client/*.go
```

#### Accounts setup

Since our Keybase is under development, we have to manually inject the private keys of the accounts we want to use in the CLI.

For the following steps, you'll need to use the accounts of the first two validators in the hard-coded development genesis file. Therefore you have some options:

1. You can just:

```bash
echo '"d936becdefdc21490dd1d1789c1887676ef2b17139526a68c79cfa66075fad2df511f0037512e802a584a1ef714790013f3db8d79e5f62cc2cae6902e1d7410b"' > /tmp/val1.json

echo '"02e31dc1242d2350d1dd463672f6b6c82881db0870705ea004cb5f8afeb52996130584fbf284bf68010b643a868b89dbbee68dc72d4e8f5e6c9bb9b48df67cd4"' > /tmp/val2.json
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
go run app/client/*.go --path_to_private_key_file=/tmp/val1.json Account Send 004aaa1375c252eecc1e8737b2062b967532b07f 003aabccc8c3938c772e783ffdd9b9fabdc118cd 1000
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
go run app/client/*.go --path_to_private_key_file=/tmp/val2.json Account Send 003aabccc8c3938c772e783ffdd9b9fabdc118cd 004aaa1375c252eecc1e8737b2062b967532b07f 1000
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
