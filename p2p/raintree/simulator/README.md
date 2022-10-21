# RainTree Simulator

The Python scripts in this package are used to simulate RainTree (in Python) in order to visualize, validate and understand the main Golang implementation.

It uses a Breadth First Search approach to mimic the real implementation of RainTree (implemented in Go) in this library, and can be considered to be a "secondary client" to verify the real P2P implementation

## Code Structure

```bash
p2p/raintree/simulator
├── README.md # This file
├── simulator.py # Utility functions used to simulate RainTree
└── test_generator.py # The entrypoint used by `make p2p_test_generator` to generate RainTree unit tests
```

## Feature Completeness

- [x] Basic RainTree implementation
- [x] Unit Test generation
- [ ] Fuzz testing
- [ ] Redundancy Layer
- [ ] Cleanup Layer
- [ ] Dead / partially visible nodes
- [ ] Multi-simulation evaluation + plotting

## Test Generation

You can specify 2 parameters to the `p2p_test_generator` make target:

- `rainTreeTestOutputFilename` # the file where the unit test should be written to
- `numRainTreeNodes`: the number of nodes to run in the RainTree simulation

Example:

```bash
rainTreeTestOutputFilename=/tmp/answer.go numRainTreeNodes=12 make p2p_test_generator
```

You can then copy pasta the output from `/tmp/answer.go` to `module_raintree_test.go` to add a new unit test.

_NOTE: You must add comments to the tree visualization component manually._
