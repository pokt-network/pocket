package rpc

import (
	"github.com/labstack/echo/v4"
)

// PostV1NodeBackup triggers a backup of the TreeStore, the BlockStore, the PostgreSQL database.
// TECHDEBT: Run each backup process in a goroutine to as elapsed time will become significant
// with the current waterfall approach when even a moderate amount of data resides in each store.
func (s *rpcServer) PostV1NodeBackup(ctx echo.Context) error {
	// TECHDEBT: Wire this up to a default config param if dir == ""
	// cfg := s.GetBus().GetRuntimeMgr().GetConfig()

	dir := ctx.Param("dir")

	s.logger.Info().Msgf("creating backup in %s", dir)

	// backup the TreeStore
	if err := s.GetBus().GetTreeStore().Backup(dir); err != nil {
		return err
	}

	// backup the BlockStore
	if err := s.GetBus().GetPersistenceModule().GetBlockStore().Backup(dir); err != nil {
		return err
	}

	s.logger.Info().Msgf("backup created in %s", dir)
	return nil
}
