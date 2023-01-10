# Genesis Module

This document is meant to be the development level documentation for the genesis state details related to the design of the codebase and information related to development.

!!! IMPORTANT !!!

This directory was created for the purposes of integration between the four core modules and is
not intended to store all the core shared types in the long-term.

Speak to @andrewnguyen22 or @Olshansk for more details.

!!! IMPORTANT !!!

## Implementation

In order to maintain code agnostic from the inception of the implementation, protobuf3 is utilized for all structures in this package.

It is important to note, that while Pocket V1 strives to not share objects between modules, the genesis module will inevitably overlap between other modules.

Another architecture worth considering and perhaps is more optimal as the project nears mainnet is allowing each module to create and maintain their own genesis object and config files

### Code Organization

```bash
genesis
├── docs
│   ├── CHANGELOG.md    # Genesis module changelog
│   ├── README.md       # Genesis module README
├── proto
│   ├── account.proto   # account structure
│   ├── actor.proto     # actor structure
│   ├── config.proto    # configuration structure
│   ├── gov.proto       # params structure
│   ├── state.proto     # genesis state structure
├── test_artifacts      # the central point of all testing code (WIP)
│   ├── generator.go    # generate the genesis and config.json for tests and build
│   ├── gov.go          # default testing parameters

```

TODO(#235): Update once runtime configs are implemented
### Module Typical Usage Example

#### Create the module

Module creation uses a typical constructor pattern signature `Create(configPath, genesisPath string) (module.Interface, error)`

Currently, module creation is not embedded or enforced in the interface to prevent the initializer from having to use 
clunky creation syntax -> `modPackage.new(module).Create(configPath, genesisPath)` rather `modPackage.Create(configPath, genesisPath)`

This is done to optimize for code clarity rather than creation signature enforceability but **may change in the future**.

```golang
newModule, err := newModule.Create(configFilePath, genesisFilePath)

if err != nil {
	// handle error
}
```

#### Set the module `bus`

The `bus` is the specific integration mechanism that enables the greater application.

Setting the `bus` allows the module to interact with its sibling modules

```golang
newModule.SetBus(bus)
```

##### Start the module

Starting the module begins the service and enables operation.

Starting must come after creation and setting the bus.

```golang
err := newModule.Start()

if err != nil {
	// handle error
}
```

#### Get the module `bus`

The bus may be accessed by the module object at anytime using the `getter`

```golang
bus := newModule.GetBus()

# The bus enables access to interfaces exposed by other modules in the codebase
bus.GetP2PModule().<FunctionName>
bus.GetPersistenceModule().<FunctionName>
...
```

#### Stop the module

Stopping the module, ends the service and disables operation.

This is the proper way to conclude the lifecycle of the module.

```golang
err := newModule.Stop()

if err != nil {
	// handle error
}
```

<!-- GITHUB_WIKI: shared/modules/readme -->
