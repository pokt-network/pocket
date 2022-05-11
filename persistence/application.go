package persistence

import (
	"encoding/hex"

	"github.com/jackc/pgx/v4"
	"github.com/pokt-network/pocket/persistence/schema"
	"github.com/pokt-network/pocket/shared/types"
)

// TODO(team): get rid of status and/or move to shared and/or create an enum
const (
	UnknownStakingStatus int = iota
	UnstakedStatus
	UnstakingStatus
	StakedStatus
)

func (p PostgresContext) GetAppExists(address []byte) (exists bool, err error) {
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return
	}
	if err = conn.QueryRow(ctx, schema.AppExistsQuery(hex.EncodeToString(address))).Scan(&exists); err != nil {
		return
	}

	return
}

func (p PostgresContext) GetApp(address []byte) (operator, publicKey, stakedTokens, maxRelays, outputAddress string, pausedHeight, unstakingHeight, endHeight int64, chains []string, err error) {
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return
	}
	if err = conn.QueryRow(ctx, schema.AppQuery(hex.EncodeToString(address))).Scan(&operator, &publicKey, &stakedTokens, &maxRelays, &outputAddress, &pausedHeight, &unstakingHeight, &endHeight); err != nil {
		return
	}

	row, err := conn.Query(ctx, schema.AppChainsQuery(hex.EncodeToString(address)))
	if err != nil {
		row.Close()
		return
	}
	defer row.Close()

	var chainID string
	var chainEndHeight int64
	for row.Next() {
		err = row.Scan(&operator, &chainID, &chainEndHeight)
		if err != nil {
			return
		}
		chains = append(chains, chainID)
	}
	return
}

// TODO(Andrew): remove paused and status from the interface
func (p PostgresContext) InsertApp(address []byte, publicKey []byte, output []byte, paused bool, status int, maxRelays string, stakedTokens string, chains []string, pausedHeight int64, unstakingHeight int64) error {
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return err
	}
	_, err = conn.Exec(ctx, schema.InsertAppQuery(hex.EncodeToString(address), hex.EncodeToString(publicKey), stakedTokens, maxRelays, hex.EncodeToString(output), pausedHeight, unstakingHeight, chains))
	return err
}

// TODO(Andrew): change `amountToAdd` to`amountToSET`
// NOTE: originally, we thought we could do arithmetic operations quite easily to just 'bump' the max relays - but since
// it's a bigint (TEXT in Postgres) I don't believe this optimization is possible. Best use new amounts for 'Update'
func (p PostgresContext) UpdateApp(address []byte, maxRelaysToAdd string, amountToAdd string, chainsToUpdate []string) error {
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
	// DISCUSS(drewsky): Should we also add a a check that it's not empty? Is an app valid without any chains?
	if chainsToUpdate != nil {
		if _, err = tx.Exec(ctx, schema.NullifyAppChainsQuery(addrString, height)); err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, schema.UpdateAppChainsQuery(addrString, chainsToUpdate, height)); err != nil {
			return err
		}
	}
	// DISCUSS(drewsky): Confirm that we can trust this input (no validation required).
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

func (p PostgresContext) DeleteApp(address []byte) error {
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

// TODO(Andrew): remove status (second parameter) - not needed
func (p PostgresContext) GetAppsReadyToUnstake(height int64, _ int) (apps []*types.UnstakingActor, err error) {
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
		apps = append(apps, &unstakingActor)
	}
	return
}

func (p PostgresContext) GetAppStatus(address []byte) (status int, err error) {
	var unstakingHeight int64
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return UnknownStakingStatus, err
	}
	height, err := p.GetHeight()
	if err != nil {
		return UnknownStakingStatus, err
	}
	if err := conn.QueryRow(ctx, schema.AppUnstakingHeightQuery(hex.EncodeToString(address), height)).Scan(&unstakingHeight); err != nil {
		return UnknownStakingStatus, err
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

// TODO(Andrew): remove status (third parameter) - no longer needed
func (p PostgresContext) SetAppUnstakingHeightAndStatus(address []byte, unstakingHeight int64, _ int) error {
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
	if _, err = tx.Exec(ctx, schema.NullifyAppQuery(hex.EncodeToString(address), height)); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, schema.UpdateAppUnstakingHeightQuery(hex.EncodeToString(address), unstakingHeight, height)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// DISCUSS(drewsky): Need to create a semantic constant for an error return value, but should it be 0 or -1?
func (p PostgresContext) GetAppPauseHeightIfExists(address []byte) (pausedHeight int64, err error) {
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return 0, err
	}
	height, err := p.GetHeight()
	if err != nil {
		return 0, err
	}
	if err := conn.QueryRow(ctx, schema.AppPauseHeightQuery(hex.EncodeToString(address), height)).Scan(&pausedHeight); err != nil {
		return 0, err
	}
	return pausedHeight, nil
}

// TODO(Andrew): remove status (third parameter) - it's not needed
// DISCUSS(drewsky): This function seems to be doing too much from a naming perspective.
func (p PostgresContext) SetAppsStatusAndUnstakingHeightPausedBefore(pausedBeforeHeight, unstakingHeight int64, _ int) error {
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
	if _, err = tx.Exec(ctx, schema.NullifyAppsPausedBeforeQuery(pausedBeforeHeight, currentHeight)); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, schema.UpdateAppsPausedBefore(pausedBeforeHeight, unstakingHeight, currentHeight)); err != nil {
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
	if err := conn.QueryRow(ctx, schema.AppOutputAddressQuery(hex.EncodeToString(operator), height)).Scan(&outputAddr); err != nil {
		return nil, err
	}
	return hex.DecodeString(outputAddr)
}
