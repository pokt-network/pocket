package persistence

import (
	"encoding/hex"

	"github.com/jackc/pgx/v4"
	"github.com/pokt-network/pocket/persistence/schema"
	"github.com/pokt-network/pocket/shared/types"
)

func (p PostgresContext) GetServiceNodeExists(address []byte) (exists bool, err error) {
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return
	}
	if err = conn.QueryRow(ctx, schema.ServiceNodeExistsQuery(hex.EncodeToString(address))).Scan(&exists); err != nil {
		return
	}
	return
}

func (p PostgresContext) GetServiceNode(address []byte) (operator, publicKey, stakedTokens, serviceURL, outputAddress string, pauseHeight, unstakingHeight, height int64, chains []string, err error) {
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return
	}
	row, err := conn.Query(ctx, schema.ServiceNodeQuery(hex.EncodeToString(address)))
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
	row, err = conn.Query(ctx, schema.ServiceNodeChainsQuery(hex.EncodeToString(address)))
	if err != nil {
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
func (p PostgresContext) InsertServiceNode(address []byte, publicKey []byte, output []byte, paused bool, status int, serviceURL string, stakedTokens string, chains []string, pausedHeight int64, unstakingHeight int64) error {
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return err
	}
	height, err := p.GetHeight()
	if err != nil {
		return err
	}
	_, err = conn.Exec(ctx, schema.InsertServiceNodeQuery(
		hex.EncodeToString(address),
		hex.EncodeToString(publicKey),
		stakedTokens,
		serviceURL,
		hex.EncodeToString(output),
		pausedHeight,
		unstakingHeight,
		height,
		chains,
	))
	return err
}

// TODO (Andrew) change amount to add, to the amount to be SET
func (p PostgresContext) UpdateServiceNode(address []byte, serviceURL string, amountToAdd string, chainsToUpdate []string) error {
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
		if _, err = tx.Exec(ctx, schema.NullifyServiceNodeChainsQuery(addrString, height)); err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, schema.UpdateServiceNodeChainsQuery(addrString, chainsToUpdate, height)); err != nil {
			return err
		}
	}
	if serviceURL != "" || amountToAdd != "" {
		if _, err = tx.Exec(ctx, schema.NullifyServiceNodeQuery(addrString, height)); err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, schema.UpdateServiceNodeQuery(addrString, amountToAdd, serviceURL, height)); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// NOTE: Leaving as transaction as I anticipate we'll need more ops in the future
func (p PostgresContext) DeleteServiceNode(address []byte) error {
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
	if _, err = tx.Exec(ctx, schema.NullifyServiceNodeQuery(addrString, height)); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, schema.NullifyServiceNodeChainsQuery(addrString, height)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (p PostgresContext) GetServiceNodeCount(chain string, height int64) (int, error) {
	var count int
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return 0, err
	}
	rows, err := conn.Query(ctx, schema.ServiceNodeCountQuery(chain, height))
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&count)
		if err != nil {
			return 0, err
		}
	}
	return count, nil
}

// TODO (Andrew) remove status - not needed
func (p PostgresContext) GetServiceNodesReadyToUnstake(height int64, status int) (ServiceNodes []*types.UnstakingActor, err error) {
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return nil, err
	}
	rows, err := conn.Query(ctx, schema.ServiceNodeReadyToUnstakeQuery(height))
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
		ServiceNodes = append(ServiceNodes, &unstakingActor)
	}
	return
}

func (p PostgresContext) GetServiceNodeStatus(address []byte) (status int, err error) {
	var unstakingHeight int64
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return 0, err
	}
	height, err := p.GetHeight()
	if err != nil {
		return 0, err
	}
	row, err := conn.Query(ctx, schema.ServiceNodeUnstakingHeightQuery(hex.EncodeToString(address), height))
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
func (p PostgresContext) SetServiceNodeUnstakingHeightAndStatus(address []byte, unstakingHeight int64, status int) error {
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
	_, err = tx.Exec(ctx, schema.NullifyServiceNodeQuery(hex.EncodeToString(address), height))
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, schema.UpdateServiceNodeUnstakingHeightQuery(hex.EncodeToString(address), unstakingHeight, height))
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (p PostgresContext) GetServiceNodePauseHeightIfExists(address []byte) (int64, error) {
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return 0, err
	}
	height, err := p.GetHeight()
	if err != nil {
		return 0, err
	}
	var pauseHeight int64
	row, err := conn.Query(ctx, schema.ServiceNodePauseHeightQuery(hex.EncodeToString(address), height))
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
func (p PostgresContext) SetServiceNodesStatusAndUnstakingHeightPausedBefore(pausedBeforeHeight, unstakingHeight int64, status int) error {
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
	_, err = tx.Exec(ctx, schema.NullifyServiceNodesPausedBeforeQuery(pausedBeforeHeight, currentHeight))
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, schema.UpdateServiceNodesPausedBefore(pausedBeforeHeight, unstakingHeight, currentHeight))
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (p PostgresContext) SetServiceNodePauseHeight(address []byte, height int64) error {
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
	if _, err = tx.Exec(ctx, schema.NullifyServiceNodeQuery(hex.EncodeToString(address), currentHeight)); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, schema.UpdateServiceNodePausedHeightQuery(hex.EncodeToString(address), height, currentHeight)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (p PostgresContext) GetServiceNodeOutputAddress(operator []byte) (output []byte, err error) {
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return nil, err
	}
	var outputAddr string
	row, err := conn.Query(ctx, schema.ServiceNodeOutputAddressQuery(hex.EncodeToString(operator)))
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
