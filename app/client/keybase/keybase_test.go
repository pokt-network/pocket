package keybase

import (
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	testPassphrase = "testingtesting123"
)

var (
	testPrivKeyAddr  = "74fa6b8f3a4ec6959a2f86b63d0774af952cdb91"
	testPrivKeyBytes = []byte{
		188, 138, 150, 24, 38, 193, 136, 7, 4, 20, 162, 74, 51, 102,
		213, 188, 192, 27, 60, 71, 20, 14, 104, 116, 80, 84, 6, 134,
		197, 240, 54, 227, 83, 112, 165, 101, 75, 106, 249, 65, 126,
		242, 179, 71, 87, 172, 95, 232, 200, 31, 67, 124, 203, 84,
		178, 160, 14, 79, 38, 79, 6, 71, 43, 236,
	}
)

// TODO: Implement unarmouring privKey str and test this is correct
func TestKeybase_CreateNewKey(t *testing.T) {
	db, err := initDB()
	defer db.Stop()
	require.NoError(t, err)

	err = db.Create(testPassphrase)
	require.NoError(t, err)

	addresses, keypairs, err := db.GetAll()
	require.NoError(t, err)

	addr := addresses[0]
	kp := keypairs[0]
	require.Equal(t, len(addr), crypto.AddressLen)
	require.Equal(t, addr, kp.PublicKey.Address().Bytes())
}

// TODO: Implement unarmouring privKey str and test this is correct
func TestKeybase_CreateNewKeyFromBytes(t *testing.T) {
	db, err := initDB()
	defer db.Stop()
	require.NoError(t, err)

	err = db.CreateFromBytes(testPrivKeyBytes, testPassphrase)
	require.NoError(t, err)

	addresses, keypairs, err := db.GetAll()
	require.NoError(t, err)

	addr := addresses[0]
	kp := keypairs[0]
	require.Equal(t, len(addr), crypto.AddressLen)
	require.Equal(t, addr, kp.PublicKey.Address().Bytes())
	require.Equal(t, kp.PublicKey.Address().String(), testPrivKeyAddr)
}

func initDB() (Keybase, error) {
	db, err := NewKeybaseInMemory("")
	if err != nil {
		return nil, err
	}
	return db, nil
}
