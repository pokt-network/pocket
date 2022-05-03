package persistence

import (
	"encoding/hex"
	"github.com/jackc/pgx/v4"
	"github.com/pokt-network/pocket/persistence/schema"
	"github.com/pokt-network/pocket/shared/types"
)

const (
	StakedStatus    = 2 // TODO (Team) get rid of status and/or move to shared
	UnstakingStatus = 1
	UnstakedStatus  = 0
)

func (p PostgresContext) GetAppExists(address []byte) (exists bool, err error) {
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return
	}
	row, err := conn.Query(ctx, schema.AppExistsQuery(hex.EncodeToString(address)))
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

func (p PostgresContext) GetApp(address []byte) (operator, publicKey, stakedTokens, maxRelays, outputAddress string, pauseHeight, unstakingHeight, endHeight int64, chains []string, err error) {
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return
	}
	row, err := conn.Query(ctx, schema.AppQuery(hex.EncodeToString(address)))
	if err != nil {
		return
	}
	for row.Next() {
		err = row.Scan(&operator, &publicKey, &stakedTokens, &maxRelays, &outputAddress, &pauseHeight, &unstakingHeight, &endHeight)
		if err != nil {
			row.Close()
			return
		}
	}
	row.Close()
	row, err = conn.Query(ctx, schema.AppChainsQuery(hex.EncodeToString(address)))
	if err != nil {
		row.Close()
		return
	}
	var chainID string
	var chainEndHeight int64
	defer row.Close()
	for row.Next() {
		err = row.Scan(&operator, &chainID, &chainEndHeight)
		if err != nil {
			return
		}
		chains = append(chains, chainID)
	}
	return
}

// TODO (Andrew) remove paused and status from the interface
func (p PostgresContext) InsertApplication(address []byte, publicKey []byte, output []byte, paused bool, status int, maxRelays string, stakedTokens string, chains []string, pausedHeight int64, unstakingHeight int64) error {
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return err
	}
	_, err = conn.Exec(ctx, schema.InsertAppQuery(hex.EncodeToString(address), hex.EncodeToString(publicKey), stakedTokens, maxRelays, hex.EncodeToString(output), pausedHeight, unstakingHeight, chains))
	return err
}

// TODO (Andrew) change amount to add, to the amount to be SET
// NOTE: originally, we thought we could do arithmetic operations quite easily to just 'bump' the max relays - but since
// it's a bigint (TEXT in Postgres) I don't beleive this optimization is possible. Best use new amounts for 'Update'
func (p PostgresContext) UpdateApplication(address []byte, maxRelaysToAdd string, amountToAdd string, chainsToUpdate []string) error {
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
	if chainsToUpdate != nil {
		if _, err = tx.Exec(ctx, schema.NullifyAppChainsQuery(addrString, height)); err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, schema.UpdateAppChainsQuery(addrString, chainsToUpdate, height)); err != nil {
			return err
		}
	}
	if maxRelaysToAdd != "" || amountToAdd != "" {
		if _, err = tx.Exec(ctx, schema.NullifyAppQuery(addrString, height)); err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, schema.UpdateAppQuery(addrString, amountToAdd, maxRelaysToAdd, height)); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// NOTE: Leaving as transaction as I anticipate we'll need more ops in the future
func (p PostgresContext) DeleteApplication(address []byte) error {
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
	if _, err = tx.Exec(ctx, schema.NullifyAppQuery(addrString, height)); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, schema.NullifyAppChainsQuery(addrString, height)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// TODO (Andrew) remove status - not needed
func (p PostgresContext) GetAppsReadyToUnstake(height int64, status int) (apps []*types.UnstakingActor, err error) {
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return nil, err
	}
	rows, err := conn.Query(ctx, schema.AppReadyToUnstakeQuery(height))
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
		apps = append(apps, &unstakingActor)
	}
	return
}

func (p PostgresContext) GetAppStatus(address []byte) (status int, err error) {
	var unstakingHeight int64
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return 0, err
	}
	height, err := p.GetHeight()
	if err != nil {
		return 0, err
	}
	row, err := conn.Query(ctx, schema.AppUnstakingHeightQuery(hex.EncodeToString(address), height))
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
func (p PostgresContext) SetAppUnstakingHeightAndStatus(address []byte, unstakingHeight int64, status int) error {
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
	_, err = tx.Exec(ctx, schema.NullifyAppQuery(hex.EncodeToString(address), height))
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, schema.UpdateAppUnstakingHeightQuery(hex.EncodeToString(address), unstakingHeight, height))
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (p PostgresContext) GetAppPauseHeightIfExists(address []byte) (int64, error) {
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return 0, err
	}
	height, err := p.GetHeight()
	if err != nil {
		return 0, err
	}
	var pauseHeight int64
	row, err := conn.Query(ctx, schema.AppPauseHeightQuery(hex.EncodeToString(address), height))
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
func (p PostgresContext) SetAppsStatusAndUnstakingHeightPausedBefore(pausedBeforeHeight, unstakingHeight int64, status int) error {
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
	_, err = tx.Exec(ctx, schema.NullifyAppsPausedBeforeQuery(pausedBeforeHeight, currentHeight))
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, schema.UpdateAppsPausedBefore(pausedBeforeHeight, unstakingHeight, currentHeight))
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (p PostgresContext) SetAppPauseHeight(address []byte, height int64) error {
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
	if _, err = tx.Exec(ctx, schema.NullifyAppQuery(hex.EncodeToString(address), currentHeight)); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, schema.UpdateAppPausedHeightQuery(hex.EncodeToString(address), height, currentHeight)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (p PostgresContext) GetAppOutputAddress(operator []byte) (output []byte, err error) {
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return nil, err
	}
	height, err := p.GetHeight()
	if err != nil {
		return nil, err
	}
	var outputAddr string
	row, err := conn.Query(ctx, schema.AppOutputAddressQuery(hex.EncodeToString(operator), height))
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
