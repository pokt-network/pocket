# Pre_Persistence Module

# Origin Document

<<<<<<< HEAD
Add a pre_persistence implementation to mock needed storage ops.
=======
Add a pre-persistence implementation to mock needed storage ops.
>>>>>>> milestone/v1-prototype
This mock should both unblock module developers and be utilized to demonstrate the storage needs of each module.
This is meant to inform the development of the v1 persistence module while enabling integration of core modules.

Creator: @andrewnguyen22

Co-Owners: @iajrz

Deliverables:

<<<<<<< HEAD
- Pre-Persistence Prototype
- How to build guide
- How to use guide
- How to test guide

## How to build

Pre_Persistence Module does not come with its own cmd executables
=======
- [ ] Pre-Persistence Prototype implementation
- [ ] How to build guide
- [ ] How to use guide
- [ ] How to test guide

## How to build

Pre_Persistence Module does not come with its own cmd executables.
>>>>>>> milestone/v1-prototype

Rather, it is purposed to be a dependency of other modules when an in-memory
persistence database is needed.

## How to use

<<<<<<< HEAD
Pre_Persistence implements the PersistenceModule and subsequent PersistenceContext interfaces
`github.com/pokt-network/pocket/shared/modules/persistence_module.go`

To use, simply initialize a Pre_Persistence instance using the constructor function:

and use as the persistence module in the desired integration / test.
=======
Pre_Persistence implements the `PersistenceModule` and subsequent `PersistenceContext` interfaces
[`pocket/shared/modules/persistence_module.go`](https://github.com/pokt-network/pocket/shared/modules/utility_module.go)

To use, simply initialize a Pre_Persistence instance using the factory function like so:

```go
prePersistenceMod, err := prePersistence.Create(config)
```

Under the hood, the PrePersistence module is initialize like so:

```
// Params: in memory goleveldb; mempool for storing transactions; global configuration object
func NewPrePersistenceModule(commitDB *memdb.DB, mempool types.Mempool, cfg *config.Config) *PrePersistenceModule {
```

You can then use it the module in the desired integration / test.
>>>>>>> milestone/v1-prototype

## How to test

```
<<<<<<< HEAD
cd persistence/pre_persistence
go test ./...
=======
make test_pre_persistence
>>>>>>> milestone/v1-prototype
```
