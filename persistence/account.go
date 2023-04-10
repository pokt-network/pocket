package persistence

import (
	"encoding/hex"
	"math/big"

	"github.com/pokt-network/pocket/persistence/types"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
)

// --- Account Functions ---

func (p *PostgresContext) GetAccountAmount(address []byte, height int64) (amount string, err error) {
	return p.getAccountAmount(types.Account, hex.EncodeToString(address), height)
}

func (p *PostgresContext) AddAccountAmount(address []byte, amount string) error {
	return p.operationAccountAmount(types.Account, hex.EncodeToString(address), amount, func(orig, delta *big.Int) error {
		orig.Add(orig, delta)
		return nil
	})
}

func (p *PostgresContext) SubtractAccountAmount(address []byte, amount string) error {
	return p.operationAccountAmount(types.Account, hex.EncodeToString(address), amount, func(orig, delta *big.Int) error {
		orig.Sub(orig, delta)
		return nil
	})
}

func (p *PostgresContext) SetAccountAmount(address []byte, amount string) error {
	return p.operationAccountAmount(types.Account, hex.EncodeToString(address), amount, func(orig, amount *big.Int) error {
		orig.Set(amount)
		return nil
	})
}

func (p *PostgresContext) GetAccountsUpdated(height int64) (accounts []*coreTypes.Account, err error) {
	return p.getAccountsUpdated(types.Account, height)
}

// --- Pool Functions ---

func (p *PostgresContext) InsertPool(address []byte, amount string) error {
	return p.insertAccount(types.Pool, hex.EncodeToString(address), amount)
}

func (p *PostgresContext) GetPoolAmount(address []byte, height int64) (amount string, err error) {
	return p.getAccountAmount(types.Pool, hex.EncodeToString(address), height)
}

func (p *PostgresContext) AddPoolAmount(address []byte, amount string) error {
	return p.operationAccountAmount(types.Pool, hex.EncodeToString(address), amount, func(orig, delta *big.Int) error {
		orig.Add(orig, delta)
		return nil
	})
}

func (p *PostgresContext) SubtractPoolAmount(address []byte, amount string) error {
	return p.operationAccountAmount(types.Pool, hex.EncodeToString(address), amount, func(orig, delta *big.Int) error {
		orig.Sub(orig, delta)
		return nil
	})
}

func (p *PostgresContext) SetPoolAmount(address []byte, amount string) error {
	return p.operationAccountAmount(types.Pool, hex.EncodeToString(address), amount, func(orig, amount *big.Int) error {
		orig.Set(amount)
		return nil
	})
}

func (p *PostgresContext) GetPoolsUpdated(height int64) ([]*coreTypes.Account, error) {
	return p.getAccountsUpdated(types.Pool, height)
}
