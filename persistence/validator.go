package persistence

import (
	"encoding/hex"
	"github.com/jackc/pgx/v4"
	"github.com/pokt-network/pocket/persistence/schema"
	"github.com/pokt-network/pocket/shared/types"
)

func (p PostgresContext) GetValidatorExists(address []byte) (exists bool, err error) {
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return
	}
	row, err := conn.Query(ctx, schema.ValidatorExistsQuery(hex.EncodeToString(address)))
	if err != nil {
		return
	}
	for row.Next() {
		err = row.Scan(&exists)
		if err != nil {
			return false, err
		}
	}
	row.Close()
	return
}

func (p PostgresContext) GetValidator(address []byte) (operator, publicKey, stakedTokens, serviceURL, outputAddress string, pauseHeight, unstakingHeight, height int64, err error) {
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return
	}
	row, err := conn.Query(ctx, schema.ValidatorQuery(hex.EncodeToString(address)))
	if err != nil {
		return
	}
	for row.Next() {
		err = row.Scan(&operator, &publicKey, &stakedTokens, &serviceURL, &outputAddress, &pauseHeight, &unstakingHeight, &height)
		if err != nil {
			return
		}
	}
	row.Close()
	return
}

// TODO (Andrew) remove paused and status from the interface
func (p PostgresContext) InsertValidator(address []byte, publicKey []byte, output []byte, paused bool, status int, serviceURL string, stakedTokens string, pausedHeight int64, unstakingHeight int64) error {
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return err
	}
	height, err := p.GetHeight()
	if err != nil {
		return err
	}
	_, err = conn.Exec(ctx, schema.InsertValidatorQuery(
		hex.EncodeToString(address),
		hex.EncodeToString(publicKey),
		stakedTokens,
		serviceURL,
		hex.EncodeToString(output),
		pausedHeight,
		unstakingHeight,
		height,
	))
	return err
}

// TODO (Andrew) change amount to add, to the amount to be SET
func (p PostgresContext) UpdateValidator(address []byte, serviceURL string, amountToAdd string) error {
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
	addrString := hex.EncodeToString(address)
	if serviceURL != "" || amountToAdd != "" {
		if _, err = tx.Exec(ctx, schema.NullifyValidatorQuery(addrString, height)); err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, schema.UpdateValidatorQuery(addrString, amountToAdd, serviceURL, height)); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// NOTE: Leaving as transaction as I anticipate we'll need more ops in the future
func (p PostgresContext) DeleteValidator(address []byte) error {
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
	addrString := hex.EncodeToString(address)
	if _, err = tx.Exec(ctx, schema.NullifyValidatorQuery(addrString, height)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// TODO (Andrew) remove status - not needed
func (p PostgresContext) GetValidatorsReadyToUnstake(height int64, status int) (Validators []*types.UnstakingActor, err error) {
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return nil, err
	}
	rows, err := conn.Query(ctx, schema.ValidatorReadyToUnstakeQuery(height))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		unstakingActor := types.UnstakingActor{}
		addr, output := "", ""
		err = rows.Scan(&addr, &unstakingActor.StakeAmount, &output)
		if err != nil {
			return nil, err
		}
		unstakingActor.Address, err = hex.DecodeString(addr)
		if err != nil {
			return nil, err
		}
		unstakingActor.OutputAddress, err = hex.DecodeString(output)
		if err != nil {
			return nil, err
		}
		Validators = append(Validators, &unstakingActor)
	}
	return
}

func (p PostgresContext) GetValidatorStatus(address []byte) (status int, err error) {
	var unstakingHeight int64
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return 0, err
	}
	height, err := p.GetHeight()
	if err != nil {
		return 0, err
	}
	row, err := conn.Query(ctx, schema.ValidatorUnstakingHeightQuery(hex.EncodeToString(address), height))
	if err != nil {
		return 0, err
	}
	defer row.Close()
	for row.Next() {
		if err = row.Scan(&unstakingHeight); err != nil {
			return 0, err
		}
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

// TODO (Andrew) remove status - no longer needed
func (p PostgresContext) SetValidatorUnstakingHeightAndStatus(address []byte, unstakingHeight int64, status int) error {
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
	_, err = tx.Exec(ctx, schema.NullifyValidatorQuery(hex.EncodeToString(address), height))
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, schema.UpdateValidatorUnstakingHeightQuery(hex.EncodeToString(address), unstakingHeight, height))
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (p PostgresContext) GetValidatorPauseHeightIfExists(address []byte) (int64, error) {
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return 0, err
	}
	height, err := p.GetHeight()
	if err != nil {
		return 0, err
	}
	var pauseHeight int64
	row, err := conn.Query(ctx, schema.ValidatorPauseHeightQuery(hex.EncodeToString(address), height))
	if err != nil {
		return 0, err
	}
	defer row.Close()
	for row.Next() {
		err = row.Scan(&pauseHeight)
		if err != nil {
			return 0, err
		}
	}
	return pauseHeight, nil
}

// TODO (Andrew) remove status - it's not needed
func (p PostgresContext) SetValidatorsStatusAndUnstakingHeightPausedBefore(pausedBeforeHeight, unstakingHeight int64, status int) error {
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
	_, err = tx.Exec(ctx, schema.NullifyValidatorsPausedBeforeQuery(pausedBeforeHeight, currentHeight))
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, schema.UpdateValidatorsPausedBefore(pausedBeforeHeight, unstakingHeight, currentHeight))
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (p PostgresContext) SetValidatorPauseHeight(address []byte, height int64) error {
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
	if _, err = tx.Exec(ctx, schema.NullifyValidatorQuery(hex.EncodeToString(address), currentHeight)); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, schema.UpdateValidatorPausedHeightQuery(hex.EncodeToString(address), height, currentHeight)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (p PostgresContext) SetValidatorStakedTokens(address []byte, tokens string) error { // TODO deprecate and use update validator
	operator, _, _, serviceURL, _, _, _, _, err := p.GetValidator(address)
	if err != nil {
		return err
	}
	addr, err := hex.DecodeString(operator)
	if err != nil {
		return err
	}
	return p.UpdateValidator(addr, serviceURL, tokens)
}

func (p PostgresContext) GetValidatorStakedTokens(address []byte) (tokens string, err error) { // TODO deprecate and use update validator
	_, _, tokens, _, _, _, _, _, err = p.GetValidator(address)
	return
}

func (p PostgresContext) GetValidatorOutputAddress(operator []byte) (output []byte, err error) {
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return nil, err
	}
	var outputAddr string
	row, err := conn.Query(ctx, schema.ValidatorOutputAddressQuery(hex.EncodeToString(operator)))
	if err != nil {
		return nil, err
	}
	defer row.Close()
	for row.Next() {
		err = row.Scan(&outputAddr)
		if err != nil {
			return nil, err
		}
	}
	return hex.DecodeString(outputAddr)
}

func (p PostgresContext) SetValidatorPauseHeightAndMissedBlocks(address []byte, pauseHeight int64, missedBlocks int) error {
	// TODO implement missed blocks
	return nil
}

func (p PostgresContext) SetValidatorMissedBlocks(address []byte, missedBlocks int) error {
	// TODO implement missed blocks
	return nil
}

func (p PostgresContext) GetValidatorMissedBlocks(address []byte) (int, error) {
	// TODO implement missed blocks
	return 0, nil
}
