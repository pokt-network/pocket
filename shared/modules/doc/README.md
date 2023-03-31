# Modules <!-- omit in toc -->

This document outlines how we structured the code by splitting it into modules, what a module is and how to create one.

<!-- IMPROVE: Add a PR example when a good example arises. -->

## Contents <!-- omit in toc -->

- [Definitions](#definitions)
	- [Shared module interface](#shared-module-interface)
	- [Module](#module)
	- [Module mock](#module-mock)
	- [Base module](#base-module)
- [Code Organization](#code-organization)
- [Modules in detail](#modules-in-detail)
	- [Module creation](#module-creation)
	- [Interacting \& Registering with the `bus`](#interacting--registering-with-the-bus)
		- [Modules Registry](#modules-registry)
			- [Modules Registry Example](#modules-registry-example)
	- [Start the module](#start-the-module)
	- [Add a logger to the module](#add-a-logger-to-the-module)
	- [Get the module `bus`](#get-the-module-bus)
	- [Stop the module](#stop-the-module)

## Definitions

### Shared module interface

A shared module interface is an interface that defines the methods that we modelled as common to multiple modules. For example: the ability to start/stop a module is pretty common and for that we defined the `InterruptableModule` interface.
There are some interfaces that are common to multiple modules and we followed the [Interface segregation principle](https://en.wikipedia.org/wiki/Interface_segregation_principle) and also [Rob Pike's Go Proverb](https://youtu.be/PAAkCSZUG1c?t=317):

> The bigger the interface, the weaker the abstraction.

These interfaces that can be embedded in modules are defined in `shared/modules/module.go`. GoDoc comments will provide you with more information about each interface.

### Module

A module is an abstraction (go interface) of a self-contained unit of functionality that carries out a specific task/set of tasks with the idea of being reusable, modular, testable and having a clear and concise API. A module might implement 0..N common interfaces. You can find additional details in the [Modules in detail](#modules-in-detail) section below.

### Module mock

A mock is a stand-in, a fake or simplified implementation of a module that is used for testing purposes. It is used to simulate the behaviour of the module and to verify that the module is interacting with other modules correctly. Mocks are generated using `go:generate` directives together with the [`mockgen` tool](https://pkg.go.dev/github.com/golang/mock#readme-running-mockgen).

### Base module

A base module is a module that implements a common interface, exposing the most basic logic. Base modules are meant to be **embedded** in module structs which implement this common interface **and** don't need to override the respective interface member(s). The intention being to improve DRYness (Don't Repeat Yourself) and to reduce boilerplate code. You can find the base modules in the `shared/modules/base_modules` package.

## Code Organization

```bash
├── base_modules            # All the base modules are defined here
├── doc                     # Documentation for the modules
├── mocks                   # Mocks of the modules will be generated in this folder
├── module.go               # Common interfaces for the modules
├── [moduleName]_module.go  # These files contain module interface definitions
└── types                   # Common types for the modules
```

## Modules in detail

We structured the code so that each module has its interface (and supporting interfaces, if any) defined in the `shared/modules` package, where the file containing the module interface follows the naming convention `[moduleName]_module.go`.
You can start by looking at the interfaces of the modules we already implemented to get a better idea of what a module is and how it should be structured.
You might notice that these files include `go:generate` directives, these are used to generate the module mocks for testing purposes.

### Module creation

Module creation uses a typical constructor pattern signature `Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error)` where `options ...modules.ModuleOption` is an optional variadic argument that allows for the passing of options to the module.
This is useful to configure the module at creation time and it's usually used during prototyping and in "sub-modules" that don't have a specific configuration file and where adding it would add unnecessary complexity and overhead. If a module has a lot of `ModuleOption`s, at that point a configuration file might be advisable.

Currently, module creation is not embedded or enforced in the interface to prevent the initializer from having to use
clunky creation syntax -> `modPackage.new(module).Create(bus modules.Bus)` rather `modPackage.Create(bus modules.Bus)`

This is done to optimize for code clarity rather than creation signature enforceability but **may change in the future**.

```golang
newModule, err := newModule.Create(bus modules.Bus)

if err != nil {
	// handle error
}
```

For an example of a module that uses a `ModuleOption`, you can search for `WithCustomRPCURL` within the codebase. The code might have changed since this document was written so we are referring to the commit hash [19bf4d3f](https://github.com/pokt-network/pocket/tree/19bf4d3f6507f5d406d9fafdb69b81359bccf110).

Essentially the `ModuleOption` sets a custom RPC URL for the module at runtime.


### Interacting & Registering with the `bus`

The `bus` is the specific integration mechanism that enables the greater application.

When a module is constructed via the `Create(bus modules.Bus, options ...modules.ModuleOption)` function, it is expected to internally call `bus.RegisterModule(module)`, which registers the module with the `bus` so its sibling modules can access it synchronously via a DI-like pattern.

#### Modules Registry

tl;dr Pocket module's version of dependency injection.

We implemented a `ModulesRegistry` module [here](https://github.com/pokt-network/pocket/blob/19bf4d3f6507f5d406d9fafdb69b81359bccf110/runtime/modules_registry.go) that takes care of the module registration and retrieval.

This module is registered with the `bus` at the application level, it is accessible to all modules via the `bus` interface and it's also mockable as you would expect.

Modules register themselves with the `bus` by calling `bus.RegisterModule(module)`. This is done in the `Create` function of the module. (For example, in the [consensus module](https://github.com/pokt-network/pocket/blob/19bf4d3f6507f5d406d9fafdb69b81359bccf110/consensus/module.go#L146))


What the `bus` does is setting its reference to the module instance and delegating the registration to the `ModulesRegistry`.

```golang
func (m *bus) RegisterModule(module modules.Module) {
	module.SetBus(m)
	m.modulesRegistry.RegisterModule(module)
}
```

Under the hood, the module name is used to map to the instance of the module so that it's possible to retrieve a module by its name.

This is quite **important** because it unlock a powerful concept **Dependency Injection**.

This enables the developer to define different implementations of a module and to register the one that is needed at runtime. This is because we can only have one module registered with a unique name and also because, by convention, we keep module names defined as constants.
This is useful not only for prototyping but also for different use cases such as the `p1` CLI and the `pocket` binary where different implementations of the same module are necessary due to the fact that the `p1` CLI doesn't have a persistence module but still needs to know what's going on in the network.

##### Modules Registry Example

For example, see the `peerstore_provider` ([here](https://github.com/pokt-network/pocket/tree/19bf4d3f6507f5d406d9fafdb69b81359bccf110/p2p/providers/peerstore_provider)).

We have a `persistence` and an `rpc` implementation of the same module and they are registered at runtime depending on the use case.

Within the `P2P` module ([here](https://github.com/pokt-network/pocket/blob/19bf4d3f6507f5d406d9fafdb69b81359bccf110/p2p/module.go#L84-L88)), we check if we have a registration of the `peerstore_provider` module and if not we fallback to the default one (the `persistence` implementation).

### Start the module

Starting the module begins the service and enables operation.

Starting must come after creation and setting the bus.

```golang
err := newModule.Start()

if err != nil {
	// handle error
}
```

### Add a logger to the module

<!-- DISCUSS: I believe we should change this convention, to me it's more semantic if logging is configured during construction/initialization and not during `Start` (which some modules might not even have if they don't implement `InterruptableModule`). I believe that a better approach is summarized here: https://github.com/pokt-network/pocket/blob/8bee148f3b0e768154be4bce02e94813c9382aac/state_machine/module.go#L29-L32 -->

When defining the start function for the module, it is essential to initialise a namespace logger as well:

```golang
func (m *newModule) Start() error {
    m.logger = logger.Global.CreateLoggerForModule(u.GetModuleName())
    ...
}
```

### Get the module `bus`

The bus may be accessed by the module object at anytime using the `getter`

```golang
bus := newModule.GetBus()

# The bus enables access to interfaces exposed by other modules in the codebase
bus.GetP2PModule().<FunctionName>
bus.GetPersistenceModule().<FunctionName>
...
```

### Stop the module

Stopping the module, ends the service and disables operation.

This is the proper way to conclude the lifecycle of the module.

```golang
err := newModule.Stop()

if err != nil {
	// handle error
}
```

<!-- GITHUB_WIKI: shared/modules/readme -->
