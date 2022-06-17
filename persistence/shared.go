package persistence

import (
	"encoding/hex"
	"github.com/jackc/pgx/v4"
	"github.com/pokt-network/pocket/persistence/schema"
	"github.com/pokt-network/pocket/shared/types"
)

// TODO(team): get rid of status and/or move to shared and/or create an enum
const (
	UnstakedStatus int = iota
	UnstakingStatus
	StakedStatus
)

func (p *PostgresContext) GetExists(address []byte, height int64, query func(string, int64) string) (exists bool, err error) {
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return
	}
	if err = conn.QueryRow(ctx, query(hex.EncodeToString(address), height)).Scan(&exists); err != nil {
		return
	}
	return
}

func (p *PostgresContext) GetActor(address []byte, height int64, query func(string, int64) string, chainsQuery func(string, int64) string) (actor schema.GenericActor, err error) {
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return
	}
	if err = conn.QueryRow(ctx, query(hex.EncodeToString(address), height)).Scan(
		&actor.Address, &actor.PublicKey, &actor.StakedTokens, &actor.GenericParam,
		&actor.OutputAddress, &actor.PausedHeight, &actor.UnstakingHeight, &height); err != nil {
		return
	}
	if chainsQuery == nil {
		return
	}
	row, err := conn.Query(ctx, chainsQuery(hex.EncodeToString(address), height))
	if err != nil {
		row.Close()
		return
	}
	defer row.Close()

	// DISCUSS(team): It's a little bit weird that the process of reading multiple items is done at the
	// logic layer, and the process of writing multiple items is done at the SQL level.
	var chainID string
	var chainEndHeight int64
	for row.Next() {
		err = row.Scan(&actor.Address, &chainID, &chainEndHeight)
		if err != nil {
			return
		}
		actor.Chains = append(actor.Chains, chainID)
	}
	return
}

func (p *PostgresContext) InsertActor(actor schema.GenericActor,
	query func(string, string, string, string, string, int64, int64, []string, int64) string) error {
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return err
	}
	height, err := p.GetHeight()
	if err != nil {
		return err
	}
	_, err = conn.Exec(ctx, query(actor.Address, actor.PublicKey, actor.StakedTokens, actor.GenericParam,
		actor.OutputAddress, actor.PausedHeight, actor.UnstakingHeight, actor.Chains, height))
	return err
}

func (p *PostgresContext) UpdateActor(actor schema.GenericActor, updateQuery func(string, string, string, int64) string,
	updateChainsQuery func(string, []string, int64) string, chainsTableName string) error {
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return err
	}
	height, err := p.GetHeight()
	if err != nil {
		return err
	}
	tx, err := conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, updateQuery(actor.Address, actor.StakedTokens, actor.GenericParam, height)); err != nil {
		return err
	}
	if actor.Chains != nil {
		if updateChainsQuery != nil {
			if _, err = tx.Exec(ctx, schema.NullifyChains(actor.Address, height, chainsTableName)); err != nil {
				return err
			}
			if _, err = tx.Exec(ctx, updateChainsQuery(actor.Address, actor.Chains, height)); err != nil {
				return err
			}
		}
	}
	return tx.Commit(ctx)
}

func (p *PostgresContext) ActorReadyToUnstakeWithChains(height int64, query func(int64) string) (actors []*types.UnstakingActor, err error) {
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return nil, err
	}
	rows, err := conn.Query(ctx, query(height))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		unstakingActor := types.UnstakingActor{}
		var addr, output string
		if err = rows.Scan(&addr, &unstakingActor.StakeAmount, &output); err != nil {
			return nil, err
		}
		if unstakingActor.Address, err = hex.DecodeString(addr); err != nil {
			return nil, err
		}
		if unstakingActor.OutputAddress, err = hex.DecodeString(output); err != nil {
			return nil, err
		}
		actors = append(actors, &unstakingActor)
	}
	return
}

func (p *PostgresContext) GetActorStatus(address []byte, height int64, query func(string, int64) string) (int, error) {
	var unstakingHeight int64
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return -1, err
	}
	if err := conn.QueryRow(ctx, query(hex.EncodeToString(address), height)).Scan(&unstakingHeight); err != nil {
		return -1, err
	}
	switch {
	case unstakingHeight == schema.DefaultUnstakingHeight:
		return StakedStatus, nil
	case unstakingHeight > height:
		return UnstakingStatus, nil
	default:
		return UnstakedStatus, nil
	}
}

func (p *PostgresContext) SetActorUnstakingHeightAndStatus(address []byte, unstakingHeight int64, query func(string, int64, int64) string) error {
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return err
	}
	height, err := p.GetHeight()
	if err != nil {
		return err
	}
	tx, err := conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, query(hex.EncodeToString(address), unstakingHeight, height)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (p *PostgresContext) GetActorPauseHeightIfExists(address []byte, height int64, query func(string, int64) string) (pausedHeight int64, err error) {
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return 0, err
	}
	if err := conn.QueryRow(ctx, query(hex.EncodeToString(address), height)).Scan(&pausedHeight); err != nil {
		return 0, err
	}
	return pausedHeight, nil
}

func (p PostgresContext) SetActorStatusAndUnstakingHeightPausedBefore(pausedBeforeHeight, unstakingHeight int64, query func(int64, int64, int64) string) error {
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return err
	}
	currentHeight, err := p.GetHeight()
	if err != nil {
		return err
	}
	tx, err := conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, query(pausedBeforeHeight, unstakingHeight, currentHeight)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (p PostgresContext) SetActorPauseHeight(address []byte, pauseHeight int64, query func(string, int64, int64) string) error {
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return err
	}
	currentHeight, err := p.GetHeight()
	if err != nil {
		return err
	}
	tx, err := conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, query(hex.EncodeToString(address), pauseHeight, currentHeight)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (p PostgresContext) GetActorOutputAddress(operator []byte, height int64, query func(string, int64) string) (output []byte, err error) {
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return nil, err
	}
	var outputAddr string
	if err := conn.QueryRow(ctx, query(hex.EncodeToString(operator), height)).Scan(&outputAddr); err != nil {
		return nil, err
	}
	return hex.DecodeString(outputAddr)
}
