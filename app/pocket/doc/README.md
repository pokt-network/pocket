# Node binary

The node binary is essentially the program that executes the node logic along with its supporting functions like for example the RPC server.

The spirit of the documentation is to continuously update and inform the reader of the general scope of the node binary as breaking, rapid development occurs.

## Flags

Currently, in order to run the node, it's necessary to provide at least two flags:

- `config`: Relative or absolute path to the config file
- `genesis`: Relative or absolute path to the genesis file.

### Example

```bash
pocket -config ./config.json -genesis ./genesis.json
```

## Configuration

The configuration file provides a structured way for configuring various aspects of the node and how it should behave functionally.

Things like "should the RPC server be enabled?", "what port should it be listening to?" are all defined in the config file.

<!-- TODO(team) add link to config after we merge #178 -->

For a detailed overview of all the possible settings, please look here.

## Genesis

The genesis file contains basically the initial state (aka genesis) of the blockchain, specifically in terms of Consensus and Utility (feel free to dive into the specific modules for more information).

<!-- TODO(team) add link to genesis after we merge #178 -->

For a detailed overview of all the possible settings, please look here.
