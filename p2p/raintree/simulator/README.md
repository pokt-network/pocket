# RainTree Simulator

The Python scripts in this package are used to simulate RainTree (in Python) in order to visualize, validate and understand the main Golang implementation.

It uses a Breadth First Search approach to mimic the real implementation of RainTree (implemented in Go) in this library, and can be considered to be a "secondary client" to verify the real P2P implementation

## Code Structure

```bash
p2p/raintree/simulator
├── README.md # This file
├── evaluator.py # WIP - The entrypoint used to collect statistics across many simulations and plot it
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
- [ ] Multi-simulation aggregation

## Test Generation

```bash
rainTreeTestOutputFilename=/tmp/answer.go numRainTreeNodes=12 make p2p_test_generator
```

## Evaluation

TODO(olshansky)
