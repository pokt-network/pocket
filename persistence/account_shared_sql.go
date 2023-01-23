package persistence

import (
	"fmt"
	"math/big"

	"github.com/jackc/pgx/v5"
	"github.com/pokt-network/pocket/persistence/types"
	"github.com/pokt-network/pocket/shared/converters"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
)

const (
	defaultAccountAmountStr = "0"
)

func (p *PostgresContext) getAccountAmount(accountSchema types.ProtocolAccountSchema, identifier string, height int64) (amount string, err error) {
	ctx, tx, err := p.getCtxAndTx()
	if err != nil {
		return
	}
	amount = defaultAccountAmountStr
	if err = tx.QueryRow(ctx, accountSchema.GetAccountAmountQuery(identifier, height)).Scan(&amount); err != pgx.ErrNoRows {
		return
	}

	return amount, nil
}

func (p *PostgresContext) operationAccountAmount(
	accountSchema types.ProtocolAccountSchema,
	identifier, amount string,
	op func(*big.Int, *big.Int) error,
) error {
	ctx, tx, err := p.getCtxAndTx()
	if err != nil {
		return err
	}
	height, err := p.GetHeight()
	if err != nil {
		return err
	}
	originalAmount, err := p.getAccountAmount(accountSchema, identifier, height)
	if err != nil {
		return err
	}
	originalAmountBig, err := converters.StringToBigInt(originalAmount)
	if err != nil {
		return err
	}
	amountBig, err := converters.StringToBigInt(amount)
	if err != nil {
		return err
	}
	if err := op(originalAmountBig, amountBig); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, accountSchema.InsertAccountQuery(identifier, converters.BigIntToString(originalAmountBig), height)); err != nil {
		return err
	}
	return nil
}

func (p *PostgresContext) getAccountsUpdated(accountType types.ProtocolAccountSchema, height int64) (accounts []*coreTypes.Account, err error) {
	query := accountType.GetAccountsUpdatedAtHeightQuery(height)

	ctx, tx, err := p.getCtxAndTx()
	if err != nil {
		return
	}

	rows, err := tx.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		acc := new(coreTypes.Account)
		if err = rows.Scan(&acc.Address, &acc.Amount); err != nil {
			return nil, err
		}
		accounts = append(accounts, acc)
	}

	return
}

func (p *PostgresContext) insertAccount(accountType types.ProtocolAccountSchema, identifier, amount string) error {
	ctx, tx, err := p.getCtxAndTx()
	if err != nil {
		return err
	}
	height, err := p.GetHeight()
	if err != nil {
		return err
	}
	// DISCUSS(team): Do we want to panic if `amount < 0` here?
	if _, err = tx.Exec(ctx, accountType.InsertAccountQuery(identifier, amount, height)); err != nil {
		return err
	}
	return nil
}

func (p *PostgresContext) getParamsOrFlagsAtHeightQuery(tableName string, height int64, descending bool) string {
	fields := "name,value"
	if tableName == types.FlagsTableName {
		fields += ",enabled"
	}
	sort := "ASC"
	if descending {
		sort = "DESC"
	}
	// Build correct query to get all Params/Flags at certain height ordered by their name values
	query := fmt.Sprintf(`SELECT %s FROM %s WHERE height=%d ORDER BY name %s`, fields, tableName, height, sort)
	return query
}
