package persistence

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/jackc/pgx/v4"

	"github.com/pokt-network/pocket/persistence/schema"
	"github.com/pokt-network/pocket/shared/types"
)

// IMPROVE(team): Move this into a proto enum
const (
	UndefinedStakingStatus = 3
	UnstakedStatus         = 0
	UnstakingStatus        = 1
	StakedStatus           = 2
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

func (p *PostgresContext) GetExists(actorSchema schema.ProtocolActorSchema, address []byte, height int64) (exists bool, err error) {
	ctx, txn, err := p.DB.GetCtxAndTxn()
	if err != nil {
		return
	}

	if err = txn.QueryRow(ctx, actorSchema.GetExistsQuery(hex.EncodeToString(address), height)).Scan(&exists); err != nil {
		return
	}

	return
}

func (p *PostgresContext) GetActor(actorSchema schema.ProtocolActorSchema, address []byte, height int64) (actor schema.BaseActor, err error) {
	ctx, txn, err := p.DB.GetCtxAndTxn()
	if err != nil {
		return
	}
	actor, height, err = p.GetActorFromRow(txn.QueryRow(ctx, actorSchema.GetQuery(hex.EncodeToString(address), height)))
	if err != nil {
		return
	}

	return p.GetChainsForActor(ctx, txn, actorSchema, actor, height)
}

func (p *PostgresContext) GetActorFromRow(row pgx.Row) (actor schema.BaseActor, height int64, err error) {
	err = row.Scan(
		&actor.Address, &actor.PublicKey, &actor.StakedTokens, &actor.ActorSpecificParam,
		&actor.OutputAddress, &actor.PausedHeight, &actor.UnstakingHeight,
		&height)
	return
}

func (p *PostgresContext) GetChainsForActor(
	ctx context.Context,
	txn pgx.Tx,
	actorSchema schema.ProtocolActorSchema,
	actor schema.BaseActor,
	height int64) (a schema.BaseActor, err error) {
	if actorSchema.GetChainsTableName() == "" {
		return
	}
	rows, err := txn.Query(ctx, actorSchema.GetChainsQuery(actor.Address, height))
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
			return actor, fmt.Errorf("unexpected address %s, expected %s when reading chains", chainAddr, actor.Address)
		}
		actor.Chains = append(actor.Chains, chainID)
	}
	return actor, nil
}

func (p *PostgresContext) InsertActor(actorSchema schema.ProtocolActorSchema, actor schema.BaseActor) error {
	ctx, txn, err := p.DB.GetCtxAndTxn()
	if err != nil {
		return err
	}

	height, err := p.GetHeight()
	if err != nil {
		return err
	}

	_, err = txn.Exec(ctx, actorSchema.InsertQuery(
		actor.Address, actor.PublicKey, actor.StakedTokens, actor.ActorSpecificParam,
		actor.OutputAddress, actor.PausedHeight, actor.UnstakingHeight, actor.Chains,
		height))
	return err
}

func (p *PostgresContext) UpdateActor(actorSchema schema.ProtocolActorSchema, actor schema.BaseActor) error {
	ctx, txn, err := p.DB.GetCtxAndTxn()
	if err != nil {
		return err
	}

	height, err := p.GetHeight()
	if err != nil {
		return err
	}

	if _, err = txn.Exec(ctx, actorSchema.UpdateQuery(actor.Address, actor.StakedTokens, actor.ActorSpecificParam, height)); err != nil {
		return err
	}

	chainsTableName := actorSchema.GetChainsTableName()
	if chainsTableName != "" && actor.Chains != nil {
		if _, err = txn.Exec(ctx, schema.NullifyChains(actor.Address, height, chainsTableName)); err != nil {
			return err
		}
		if _, err = txn.Exec(ctx, actorSchema.UpdateChainsQuery(actor.Address, actor.Chains, height)); err != nil {
			return err
		}
	}

	return nil
}

func (p *PostgresContext) GetActorsReadyToUnstake(actorSchema schema.ProtocolActorSchema, height int64) (actors []*types.UnstakingActor, err error) {
	ctx, txn, err := p.DB.GetCtxAndTxn()
	if err != nil {
		return nil, err
	}

	rows, err := txn.Query(ctx, actorSchema.GetReadyToUnstakeQuery(height))
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

func (p *PostgresContext) GetActorStatus(actorSchema schema.ProtocolActorSchema, address []byte, height int64) (int, error) {
	var unstakingHeight int64
	ctx, txn, err := p.DB.GetCtxAndTxn()
	if err != nil {
		return UndefinedStakingStatus, err
	}

	if err := txn.QueryRow(ctx, actorSchema.GetUnstakingHeightQuery(hex.EncodeToString(address), height)).Scan(&unstakingHeight); err != nil {
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

func (p *PostgresContext) SetActorUnstakingHeightAndStatus(actorSchema schema.ProtocolActorSchema, address []byte, unstakingHeight int64) error {
	ctx, txn, err := p.DB.GetCtxAndTxn()
	if err != nil {
		return err
	}

	height, err := p.GetHeight()
	if err != nil {
		return err
	}

	_, err = txn.Exec(ctx, actorSchema.UpdateUnstakingHeightQuery(hex.EncodeToString(address), unstakingHeight, height))
	return err
}

func (p *PostgresContext) GetActorPauseHeightIfExists(actorSchema schema.ProtocolActorSchema, address []byte, height int64) (pausedHeight int64, err error) {
	ctx, txn, err := p.DB.GetCtxAndTxn()
	if err != nil {
		return schema.DefaultBigInt, err
	}

	if err := txn.QueryRow(ctx, actorSchema.GetPausedHeightQuery(hex.EncodeToString(address), height)).Scan(&pausedHeight); err != nil {
		return schema.DefaultBigInt, err
	}

	return pausedHeight, nil
}

func (p PostgresContext) SetActorStatusAndUnstakingHeightIfPausedBefore(actorSchema schema.ProtocolActorSchema, pausedBeforeHeight, unstakingHeight int64) error {
	ctx, txn, err := p.DB.GetCtxAndTxn()
	if err != nil {
		return err
	}

	currentHeight, err := p.GetHeight()
	if err != nil {
		return err
	}

	_, err = txn.Exec(ctx, actorSchema.UpdateUnstakedHeightIfPausedBeforeQuery(pausedBeforeHeight, unstakingHeight, currentHeight))
	return err
}

func (p PostgresContext) SetActorPauseHeight(actorSchema schema.ProtocolActorSchema, address []byte, pauseHeight int64) error {
	ctx, txn, err := p.DB.GetCtxAndTxn()
	if err != nil {
		return err
	}

	currentHeight, err := p.GetHeight()
	if err != nil {
		return err
	}

	_, err = txn.Exec(ctx, actorSchema.UpdatePausedHeightQuery(hex.EncodeToString(address), pauseHeight, currentHeight))
	return err
}

func (p PostgresContext) SetActorStakeAmount(actorSchema schema.ProtocolActorSchema, address []byte, stakeAmount string) error {
	ctx, txn, err := p.DB.GetCtxAndTxn()
	if err != nil {
		return err
	}

	currentHeight, err := p.GetHeight()
	if err != nil {
		return err
	}
	_, err = txn.Exec(ctx, actorSchema.SetStakeAmountQuery(hex.EncodeToString(address), stakeAmount, currentHeight))
	return err
}

func (p PostgresContext) GetActorOutputAddress(actorSchema schema.ProtocolActorSchema, operatorAddr []byte, height int64) ([]byte, error) {
	ctx, txn, err := p.DB.GetCtxAndTxn()
	if err != nil {
		return nil, err
	}

	var outputAddr string
	if err := txn.QueryRow(ctx, actorSchema.GetOutputAddressQuery(hex.EncodeToString(operatorAddr), height)).Scan(&outputAddr); err != nil {
		return nil, err
	}

	return hex.DecodeString(outputAddr)
}

func (p PostgresContext) GetActorStakeAmount(actorSchema schema.ProtocolActorSchema, address []byte, height int64) (string, error) {
	ctx, txn, err := p.DB.GetCtxAndTxn()
	if err != nil {
		return types.EmptyString, err
	}

	var stakeAmount string
	if err := txn.QueryRow(ctx, actorSchema.GetStakeAmountQuery(hex.EncodeToString(address), height)).Scan(&stakeAmount); err != nil {
		return types.EmptyString, err
	}
	return stakeAmount, nil
}
