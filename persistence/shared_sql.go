package persistence

import (
	"encoding/hex"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/pokt-network/pocket/persistence/schema"
	"github.com/pokt-network/pocket/shared/types"
)

// IMPROVE(team): Move this into a proto enum
const (
	UndefinedStakingStatus = iota
	UnstakedStatus
	UnstakingStatus
	StakedStatus
)

func (p *PostgresContext) GetExists(actorSchema schema.ProtocolActor, address []byte, height int64) (exists bool, err error) {
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return
	}

	if err = conn.QueryRow(ctx, actorSchema.GetExistsQuery(hex.EncodeToString(address), height)).Scan(&exists); err != nil {
		return
	}

	return
}

func (p *PostgresContext) GetActor(actorSchema schema.ProtocolActor, address []byte, height int64) (actor schema.GenericActor, err error) {
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return
	}

	if err = conn.QueryRow(ctx, actorSchema.GetQuery(hex.EncodeToString(address), height)).Scan(
		&actor.Address, &actor.PublicKey, &actor.StakedTokens, &actor.GenericParam,
		&actor.OutputAddress, &actor.PausedHeight, &actor.UnstakingHeight,
		&height,
	); err != nil {
		return
	}

	if actorSchema.GetChainsTableName() == "" {
		return
	}

	rows, err := conn.Query(ctx, actorSchema.GetChainsQuery(hex.EncodeToString(address), height))
	if err != nil {
		return
	}
	defer rows.Close()

	var chainAddr string
	var chainID string
	var chainEndHeight int64 // unused
	for rows.Next() {
		err = rows.Scan(&chainAddr, &chainID, &chainEndHeight)
		if err != nil {
			return
		}
		if chainAddr != actor.Address {
			return actor, fmt.Errorf("unexpected address %s, expected %s when reading chains", chainAddr, address)
		}
		actor.Chains = append(actor.Chains, chainID)
	}

	return
}

func (p *PostgresContext) InsertActor(actorSchema schema.ProtocolActor, actor schema.GenericActor) error {
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return err
	}

	height, err := p.GetHeight()
	if err != nil {
		return err
	}

	_, err = conn.Exec(ctx, actorSchema.InsertQuery(
		actor.Address, actor.PublicKey, actor.StakedTokens, actor.GenericParam,
		actor.OutputAddress, actor.PausedHeight, actor.UnstakingHeight, actor.Chains,
		height))
	return err
}

func (p *PostgresContext) UpdateActor(actorSchema schema.ProtocolActor, actor schema.GenericActor) error {
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

	if _, err = tx.Exec(ctx, actorSchema.UpdateQuery(actor.Address, actor.StakedTokens, actor.GenericParam, height)); err != nil {
		return err
	}

	chainsTableName := actorSchema.GetChainsTableName()
	if chainsTableName != "" && actor.Chains != nil {
		if _, err = tx.Exec(ctx, schema.NullifyChains(actor.Address, height, chainsTableName)); err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, actorSchema.UpdateChainsQuery(actor.Address, actor.Chains, height)); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (p *PostgresContext) ActorReadyToUnstakeWithChains(actorSchema schema.ProtocolActor, height int64) (actors []*types.UnstakingActor, err error) {
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return nil, err
	}

	rows, err := conn.Query(ctx, actorSchema.GetReadyToUnstakeQuery(height))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		// IMPROVE(team): Can we refactor so we pass the unstaking actor fields directly?
		unstakingActor := types.UnstakingActor{}
		var addr, output string
		if err = rows.Scan(&addr, &unstakingActor.StakeAmount, &output); err != nil {
			return
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

func (p *PostgresContext) GetActorStatus(actorSchema schema.ProtocolActor, address []byte, height int64) (int, error) {
	var unstakingHeight int64
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return -1, err
	}

	if err := conn.QueryRow(ctx, actorSchema.GetUnstakingHeightQuery(hex.EncodeToString(address), height)).Scan(&unstakingHeight); err != nil {
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

func (p *PostgresContext) SetActorUnstakingHeightAndStatus(actorSchema schema.ProtocolActor, address []byte, unstakingHeight int64) error {
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return err
	}

	height, err := p.GetHeight()
	if err != nil {
		return err
	}

	_, err = conn.Exec(ctx, actorSchema.UpdateUnstakingHeightQuery(hex.EncodeToString(address), unstakingHeight, height))
	return err
}

func (p *PostgresContext) GetActorPauseHeightIfExists(actorSchema schema.ProtocolActor, address []byte, height int64) (pausedHeight int64, err error) {
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return 0, err
	}

	if err := conn.QueryRow(ctx, actorSchema.GetPausedHeightQuery(hex.EncodeToString(address), height)).Scan(&pausedHeight); err != nil {
		return 0, err
	}

	return pausedHeight, nil
}

func (p PostgresContext) SetActorStatusAndUnstakingHeightPausedBefore(actorSchema schema.ProtocolActor, pausedBeforeHeight, unstakingHeight int64) error {
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return err
	}

	currentHeight, err := p.GetHeight()
	if err != nil {
		return err
	}

	_, err = conn.Exec(ctx, actorSchema.UpdatePausedBefore(pausedBeforeHeight, unstakingHeight, currentHeight))
	return err
}

func (p PostgresContext) SetActorPauseHeight(actorSchema schema.ProtocolActor, address []byte, pauseHeight int64) error {
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return err
	}

	currentHeight, err := p.GetHeight()
	if err != nil {
		return err
	}

	_, err = conn.Exec(ctx, actorSchema.UpdatePausedHeightQuery(hex.EncodeToString(address), pauseHeight, currentHeight))
	return err
}

func (p PostgresContext) GetActorOutputAddress(actorSchema schema.ProtocolActor, operatorAddr []byte, height int64) ([]byte, error) {
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return nil, err
	}

	var outputAddr string
	if err := conn.QueryRow(ctx, actorSchema.GetOutputAddressQuery(hex.EncodeToString(operatorAddr), height)).Scan(&outputAddr); err != nil {
		return nil, err
	}

	return hex.DecodeString(outputAddr)
}
