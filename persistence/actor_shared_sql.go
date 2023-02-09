package persistence

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/pokt-network/pocket/persistence/types"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	moduleTypes "github.com/pokt-network/pocket/shared/modules/types"
)

// IMPROVE(team): Move this into a proto enum. We are not using `iota` for the time being
// for the purpose of being explicit: https://github.com/pokt-network/pocket/pull/140#discussion_r939731342
// TODO: Consolidate with proto enum in the utility module
const (
	UndefinedStakingStatus = int32(0)
	UnstakingStatus        = int32(1)
	StakedStatus           = int32(2)
	UnstakedStatus         = int32(3)
)

func UnstakingHeightToStatus(unstakingHeight int64) int32 {
	switch unstakingHeight {
	case -1:
		return StakedStatus
	case 0:
		return UnstakedStatus
	default:
		return UnstakingStatus
	}
}

func (p *PostgresContext) GetExists(actorSchema types.ProtocolActorSchema, address []byte, height int64) (exists bool, err error) {
	ctx, tx := p.getCtxAndTx()

	if err = tx.QueryRow(ctx, actorSchema.GetExistsQuery(hex.EncodeToString(address), height)).Scan(&exists); err != nil {
		return
	}

	return
}

func (p *PostgresContext) GetActorsUpdated(actorSchema types.ProtocolActorSchema, height int64) (actors []*coreTypes.Actor, err error) {
	ctx, tx := p.getCtxAndTx()

	rows, err := tx.Query(ctx, actorSchema.GetUpdatedAtHeightQuery(height))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	addrs := make([][]byte, 0)
	for rows.Next() {
		var addr string
		if err := rows.Scan(&addr); err != nil {
			return nil, err
		}
		addrBz, err := hex.DecodeString(addr)
		if err != nil {
			return nil, err
		}
		addrs = append(addrs, addrBz)
	}
	rows.Close()

	actors = make([]*coreTypes.Actor, len(addrs))
	for i, addr := range addrs {
		actor, err := p.getActor(actorSchema, addr, height)
		if err != nil {
			return nil, err
		}
		actors[i] = actor
	}

	return
}

func (p *PostgresContext) getActor(actorSchema types.ProtocolActorSchema, address []byte, height int64) (actor *coreTypes.Actor, err error) {
	ctx, tx := p.getCtxAndTx()
	actor, height, err = p.getActorFromRow(actorSchema.GetActorType(), tx.QueryRow(ctx, actorSchema.GetQuery(hex.EncodeToString(address), height)))
	if err != nil {
		return
	}
	return p.getChainsForActor(ctx, tx, actorSchema, actor, height)
}

func (p *PostgresContext) getActorFromRow(actorType coreTypes.ActorType, row pgx.Row) (actor *coreTypes.Actor, height int64, err error) {
	actor = &coreTypes.Actor{
		ActorType: actorType,
	}
	err = row.Scan(
		&actor.Address,
		&actor.PublicKey,
		&actor.StakedAmount,
		&actor.GenericParam,
		&actor.Output,
		&actor.PausedHeight,
		&actor.UnstakingHeight,
		&height)
	return
}

func (p *PostgresContext) getChainsForActor(
	ctx context.Context,
	tx pgx.Tx,
	actorSchema types.ProtocolActorSchema,
	actor *coreTypes.Actor,
	height int64,
) (a *coreTypes.Actor, err error) {
	if actorSchema.GetChainsTableName() == "" {
		return actor, nil
	}
	rows, err := tx.Query(ctx, actorSchema.GetChainsQuery(actor.Address, height))
	if err != nil {
		return actor, err
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
			return actor, fmt.Errorf("unexpected address %s, expected %s when reading chains", chainAddr, actor.Address)
		}
		actor.Chains = append(actor.Chains, chainID)
	}
	return actor, nil
}

func (p *PostgresContext) InsertActor(actorSchema types.ProtocolActorSchema, actor *coreTypes.Actor) error {
	ctx, tx := p.getCtxAndTx()

	height, err := p.GetHeight()
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, actorSchema.InsertQuery(
		actor.Address, actor.PublicKey, actor.StakedAmount, actor.GenericParam,
		actor.Output, actor.PausedHeight, actor.UnstakingHeight, actor.Chains,
		height))
	return err
}

func (p *PostgresContext) UpdateActor(actorSchema types.ProtocolActorSchema, actor *coreTypes.Actor) error {
	ctx, tx := p.getCtxAndTx()

	height, err := p.GetHeight()
	if err != nil {
		return err
	}

	if _, err = tx.Exec(ctx, actorSchema.UpdateQuery(actor.Address, actor.StakedAmount, actor.GenericParam, height)); err != nil {
		return err
	}

	chainsTableName := actorSchema.GetChainsTableName()
	if chainsTableName != "" && actor.Chains != nil {
		if _, err = tx.Exec(ctx, types.NullifyChains(actor.Address, height, chainsTableName)); err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, actorSchema.UpdateChainsQuery(actor.Address, actor.Chains, height)); err != nil {
			return err
		}
	}

	return nil
}

func (p *PostgresContext) GetActorsReadyToUnstake(actorSchema types.ProtocolActorSchema, height int64) (actors []*moduleTypes.UnstakingActor, err error) {
	ctx, tx := p.getCtxAndTx()

	rows, err := tx.Query(ctx, actorSchema.GetReadyToUnstakeQuery(height))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		actor := &moduleTypes.UnstakingActor{}
		if err = rows.Scan(&actor.Address, &actor.StakeAmount, &actor.OutputAddress); err != nil {
			return
		}
		actors = append(actors, actor)
	}
	return
}

func (p *PostgresContext) GetActorStatus(actorSchema types.ProtocolActorSchema, address []byte, height int64) (int32, error) {
	var unstakingHeight int64
	ctx, tx := p.getCtxAndTx()

	if err := tx.QueryRow(ctx, actorSchema.GetUnstakingHeightQuery(hex.EncodeToString(address), height)).Scan(&unstakingHeight); err != nil {
		return UndefinedStakingStatus, err
	}

	switch {
	case unstakingHeight == -1:
		return StakedStatus, nil
	case unstakingHeight > height:
		return UnstakingStatus, nil
	default:
		return UnstakedStatus, nil
	}
}

func (p *PostgresContext) SetActorUnstakingHeightAndStatus(actorSchema types.ProtocolActorSchema, address []byte, unstakingHeight int64) error {
	ctx, tx := p.getCtxAndTx()

	height, err := p.GetHeight()
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, actorSchema.UpdateUnstakingHeightQuery(hex.EncodeToString(address), unstakingHeight, height))
	return err
}

func (p *PostgresContext) GetActorPauseHeightIfExists(actorSchema types.ProtocolActorSchema, address []byte, height int64) (pausedHeight int64, err error) {
	ctx, tx := p.getCtxAndTx()

	if err := tx.QueryRow(ctx, actorSchema.GetPausedHeightQuery(hex.EncodeToString(address), height)).Scan(&pausedHeight); err != nil {
		return types.DefaultBigInt, err
	}

	return pausedHeight, nil
}

func (p *PostgresContext) SetActorStatusAndUnstakingHeightIfPausedBefore(actorSchema types.ProtocolActorSchema, pausedBeforeHeight, unstakingHeight int64) error {
	ctx, tx := p.getCtxAndTx()

	currentHeight, err := p.GetHeight()
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, actorSchema.UpdateUnstakedHeightIfPausedBeforeQuery(pausedBeforeHeight, unstakingHeight, currentHeight))
	return err
}

func (p *PostgresContext) SetActorPauseHeight(actorSchema types.ProtocolActorSchema, address []byte, pauseHeight int64) error {
	ctx, tx := p.getCtxAndTx()

	currentHeight, err := p.GetHeight()
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, actorSchema.UpdatePausedHeightQuery(hex.EncodeToString(address), pauseHeight, currentHeight))
	return err
}

func (p *PostgresContext) setActorStakeAmount(actorSchema types.ProtocolActorSchema, address []byte, stakeAmount string) error {
	ctx, tx := p.getCtxAndTx()

	currentHeight, err := p.GetHeight()
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, actorSchema.SetStakeAmountQuery(hex.EncodeToString(address), stakeAmount, currentHeight))
	return err
}

func (p *PostgresContext) GetActorOutputAddress(actorSchema types.ProtocolActorSchema, operatorAddr []byte, height int64) ([]byte, error) {
	ctx, tx := p.getCtxAndTx()

	var outputAddr string
	if err := tx.QueryRow(ctx, actorSchema.GetOutputAddressQuery(hex.EncodeToString(operatorAddr), height)).Scan(&outputAddr); err != nil {
		return nil, err
	}

	return hex.DecodeString(outputAddr)
}

func (p *PostgresContext) getActorStakeAmount(actorSchema types.ProtocolActorSchema, address []byte, height int64) (string, error) {
	ctx, tx := p.getCtxAndTx()

	var stakeAmount string
	if err := tx.QueryRow(ctx, actorSchema.GetStakeAmountQuery(hex.EncodeToString(address), height)).Scan(&stakeAmount); err != nil {
		return "", err
	}
	return stakeAmount, nil
}
