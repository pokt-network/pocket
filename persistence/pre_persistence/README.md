# Pre_Persistence Module

# Origin Document
Add a pre-persistence implementation to mock needed storage ops. 
This mock should both unblock module developers and be utilized to demonstrate the storage needs of each module. 
This is meant to inform the development of the v1 persistence module while enabling integration of core modules.

Creator: @andrewnguyen22

Co-Owners: @iajrz

Deliverables:
- [ ] Pre-Persistence Prototype implementation
- [ ] How to build guide
- [ ] How to use guide
- [ ] How to test guide

## How to build

Pre_Persistence Module does not come with its own cmd executables

Rather, it is purposed to be a dependency of other modules when an in-memory
persistence database is needed.

## How to use

Pre_Persistence implements the /Pre_Persistence/PrePersitence and subsequent PersistenceContext interfaces 
`github.com/pokt-network/pocket/shared/modules/persistence_module.go`

To use, simply initialize a Pre_Persistence instance using the constructor function:

```
// Params: in memory goleveldb; mempool for storing transactions; global configuration object
func NewPrePersistenceModule(commitDB *memdb.DB, mempool types.Mempool, cfg *config.Config) *PrePersistenceModule {
```

and use as the persistence module in the desired integration / test. 

## How to test
```
go test ./persistence/pre_persistence/....
```