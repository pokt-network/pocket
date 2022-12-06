# Iteration 3 Demo <!-- omit in toc -->

**Table of Contents**

- [Goals](#goals)
- [Shell #1: Setup LocalNet](#shell-1-setup-localnet)
- [Shell #2: Setup Consensus debugger](#shell-2-setup-consensus-debugger)
- [Shell #3: Inspect the data in the database for node1](#shell-3-inspect-the-data-in-the-database-for-node1)
- [Shell #4: Inspect the data in the database for node3](#shell-4-inspect-the-data-in-the-database-for-node3)
- [Shell #5: Trigger a send transaction from the CLI](#shell-5-trigger-a-send-transaction-from-the-cli)
  - [Available Commands](#available-commands)
  - [First Transaction](#first-transaction)
  - [Second Transaction](#second-transaction)
- [\[Optional\] Shell #6: See Swagger UI](#optional-shell-6-see-swagger-ui)

## Goals

The first video of this demo can be accessed [here](https://drive.google.com/file/d/1IOrzq-XJP04BJjyqPPpPu873aSfwrnur/view?usp=sharing).

The goal of iteration 3 was to have a success end-to-end transaction that:

- Uses docker-compose on LocalNet
- Is composed of 4 hard-coded validator nodes
- Starts the LocalNet from genesis
- Uses the CLI to send a transaction
- Uses a basic version of HotPOKT for consensus
- Uses a basic version of RainTree for brodcast

<img width="842" alt="Screenshot 2022-12-05 at 9 02 28 PM" src="https://user-images.githubusercontent.com/1892194/205820691-26e801e4-ff79-4132-a7a1-358860ca2335.png">

## Shell #1: Setup LocalNet

```bash
m̶a̶k̶e̶ ̶d̶o̶c̶k̶e̶r̶_̶w̶i̶p̶e̶ # [Optional] Clear everything (takes a long time)
make # show all the commands
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

### First Transaction

Trigger a send transaction from validator 1 to validator 2.

```bash
go run app/client/*.go --path_to_private_key_file=/Users/olshansky/workspace/pocket/pocket/build/pkeys/val1.json Account Send 6f66574e1f50f0ef72dff748c3f11b9e0e89d32a 67eb3f0a50ae459fecf666be0e93176e92441317 1000
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
go run app/client/*.go --path_to_private_key_file=/Users/olshansky/workspace/pocket/pocket/build/pkeys/val2.json Account Send 67eb3f0a50ae459fecf666be0e93176e92441317 6f66574e1f50f0ef72dff748c3f11b9e0e89d32a 1000
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
