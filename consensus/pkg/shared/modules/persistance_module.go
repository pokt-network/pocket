package modules

type PersistanceModule interface {
	PocketModule
	GetLatestBlockHeight() (uint64, error)
	GetBlockHash(height uint64) ([]byte, error)
}
