package rpc

import (
	"github.com/labstack/echo/v4"
)

func (s *rpcServer) PostV1NodeBackup(ctx echo.Context) error {
	store := s.GetBus().GetPersistenceModule().GetBus().GetTreeStore()
	rw, err := s.GetBus().GetPersistenceModule().NewRWContext(0)
	if err != nil {
		return err
	}
	if err := rw.SetSavePoint(); err != nil {
		return err
	}
	if err := store.Backup(ctx.Param("dir")); err != nil {
		return err
	}
	rw.Release()
	s.logger.Info().Msgf("backup created in %s", ctx.Param("dir"))
	return nil
}
