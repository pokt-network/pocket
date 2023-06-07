package modules

const (
	TreeStoreModuleName = "tree_store"
)

type TreeStoreModule interface {
	Module

	TreeStore
}
