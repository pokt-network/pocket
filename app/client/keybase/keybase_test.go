package keybase

import (
	"encoding/hex"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	testPassphrase    = "testingtesting123"
	testNewPassphrase = "321gnitsetgnitset"
)

var (
	testKey, _ = createTestKey()
)

func TestKeybase_CreateNewKey(t *testing.T) {
	db, err := initDB()
	defer db.Stop()
	require.NoError(t, err)

	err = db.Create(testPassphrase)
	require.NoError(t, err)

	addresses, keypairs, err := db.GetAll()
	require.NoError(t, err)
	require.Equal(t, len(addresses), 1)
	require.Equal(t, len(keypairs), 1)

	addr := addresses[0]
	kp := keypairs[0]
	require.Equal(t, len(addr), crypto.AddressLen)
	require.Equal(t, addr, kp.GetAddressBytes())
}

func TestKeybase_CreateNewKeyFromBytes(t *testing.T) {
	db, err := initDB()
	defer db.Stop()
	require.NoError(t, err)

	err = db.CreateFromBytes(testKey.Bytes(), testPassphrase)
	require.NoError(t, err)

	addresses, keypairs, err := db.GetAll()
	require.NoError(t, err)
	require.Equal(t, len(addresses), 1)
	require.Equal(t, len(keypairs), 1)

	addr := addresses[0]
	kp := keypairs[0]
	require.Equal(t, len(addr), crypto.AddressLen)
	require.Equal(t, addr, kp.GetAddressBytes())
	require.Equal(t, kp.GetAddressString(), testKey.Address().String())

	privKey, err := kp.Unarmour(testPassphrase)
	require.NoError(t, err)
	require.Equal(t, privKey.String(), testKey.String())
}

func TestKeybase_CreateNewKeyFromString(t *testing.T) {
	db, err := initDB()
	defer db.Stop()
	require.NoError(t, err)

	err = db.CreateFromString(testKey.String(), testPassphrase)
	require.NoError(t, err)

	addresses, keypairs, err := db.GetAll()
	require.NoError(t, err)
	require.Equal(t, len(addresses), 1)
	require.Equal(t, len(keypairs), 1)

	addr := addresses[0]
	kp := keypairs[0]
	require.Equal(t, len(addr), crypto.AddressLen)
	require.Equal(t, addr, kp.GetAddressBytes())
	require.Equal(t, kp.GetAddressString(), testKey.Address().String())

	privKey, err := kp.Unarmour(testPassphrase)
	require.NoError(t, err)
	require.Equal(t, privKey.String(), testKey.String())
}

// TODO: Improve this test/create functions to check string validity
func TestKeybase_CreateNewKeyFromStringInvalidString(t *testing.T) {
	db, err := initDB()
	defer db.Stop()
	require.NoError(t, err)

	falseAddr := testKey.String() + "aa"
	falseBz, err := hex.DecodeString(falseAddr)
	require.NoError(t, err)

	err = db.CreateFromString(falseAddr, testPassphrase)
	require.EqualError(t, err, crypto.ErrInvalidPrivateKeyLen(len(falseBz)).Error())
}

func TestKeybase_GetKey(t *testing.T) {
	db, err := initDB()
	defer db.Stop()
	require.NoError(t, err)

	err = db.CreateFromString(testKey.String(), testPassphrase)
	require.NoError(t, err)

	kp, err := db.Get(testKey.Address().String())
	require.NoError(t, err)
	require.Equal(t, testKey.Address().Bytes(), kp.GetAddressBytes())
	require.Equal(t, kp.GetAddressString(), testKey.Address().String())

	privKey, err := kp.Unarmour(testPassphrase)
	require.NoError(t, err)

	equal := privKey.Equals(testKey)
	require.Equal(t, equal, true)
	require.Equal(t, privKey.String(), testKey.String())
}

func TestKeybase_GetKeyDoesntExist(t *testing.T) {
	db, err := initDB()
	defer db.Stop()
	require.NoError(t, err)

	kp, err := db.Get(testKey.Address().String())
	require.EqualError(t, err, ErrorAddrNotFound(testKey.Address().String()).Error())
	require.Equal(t, kp, KeyPair{})
}

func TestKeybase_CheckKeyExists(t *testing.T) {
	db, err := initDB()
	defer db.Stop()
	require.NoError(t, err)

	err = db.CreateFromString(testKey.String(), testPassphrase)
	require.NoError(t, err)

	exists, err := db.Exists(testKey.Address().String())
	require.NoError(t, err)
	require.Equal(t, exists, true)
}

func TestKeybase_CheckKeyExistsDoesntExist(t *testing.T) {
	db, err := initDB()
	defer db.Stop()
	require.NoError(t, err)

	exists, err := db.Exists(testKey.Address().String())
	require.EqualError(t, err, ErrorAddrNotFound(testKey.Address().String()).Error())
	require.Equal(t, exists, false)
}

func TestKeybase_GetAllKeys(t *testing.T) {
	db, err := initDB()
	defer db.Stop()
	require.NoError(t, err)

	pks := make(map[string]crypto.PrivateKey, 0)
	for i := 0; i < 5; i++ {
		pk, err := createTestKey()
		require.NoError(t, err)
		err = db.CreateFromString(pk.String(), testPassphrase)
		require.NoError(t, err)
		pks[pk.Address().String()] = pk
	}

	addresses, keypairs, err := db.GetAll()
	require.NoError(t, err)
	require.Equal(t, len(addresses), 5)
	require.Equal(t, len(keypairs), 5)

	for i := 0; i < 5; i++ {
		privKey, err := keypairs[i].Unarmour(testPassphrase)
		require.NoError(t, err)

		require.Equal(t, addresses[i], keypairs[i].GetAddressBytes())
		require.Equal(t, addresses[i], privKey.Address().Bytes())

		equal := privKey.Equals(pks[privKey.Address().String()])
		require.Equal(t, equal, true)
	}
}

func TestKeybase_GetPrivKey(t *testing.T) {
	db, err := initDB()
	defer db.Stop()
	require.NoError(t, err)

	err = db.CreateFromString(testKey.String(), testPassphrase)
	require.NoError(t, err)

	privKey, err := db.GetPrivKey(testKey.Address().String(), testPassphrase)
	require.NoError(t, err)
	require.Equal(t, testKey.Address().Bytes(), privKey.Address().Bytes())
	require.Equal(t, privKey.Address().String(), testKey.Address().String())

	equal := privKey.Equals(testKey)
	require.Equal(t, equal, true)
	require.Equal(t, privKey.String(), testKey.String())
}

func TestKeybase_GetPrivKeyWrongPassphrase(t *testing.T) {
	db, err := initDB()
	defer db.Stop()
	require.NoError(t, err)

	err = db.CreateFromString(testKey.String(), testPassphrase)
	require.NoError(t, err)

	privKey, err := db.GetPrivKey(testKey.Address().String(), testNewPassphrase)
	require.Equal(t, err, ErrorWrongPassphrase)
	require.Nil(t, privKey)
}

func TestBadgerKeybase_UpdatePassphrase(t *testing.T) {
	db, err := initDB()
	defer db.Stop()
	require.NoError(t, err)

	err = db.CreateFromString(testKey.String(), testPassphrase)
	require.NoError(t, err)

	_, err = db.GetPrivKey(testKey.Address().String(), testPassphrase)
	require.NoError(t, err)

	err = db.UpdatePassphrase(testKey.Address().String(), testPassphrase, testNewPassphrase)
	require.NoError(t, err)

	privKey, err := db.GetPrivKey(testKey.Address().String(), testNewPassphrase)
	require.NoError(t, err)
	require.Equal(t, testKey.Address().Bytes(), privKey.Address().Bytes())
	require.Equal(t, privKey.Address().String(), testKey.Address().String())

	equal := privKey.Equals(testKey)
	require.Equal(t, equal, true)
	require.Equal(t, privKey.String(), testKey.String())
}

func TestBadgerKeybase_UpdatePassphraseWrongPassphrase(t *testing.T) {
	db, err := initDB()
	defer db.Stop()
	require.NoError(t, err)

	err = db.CreateFromString(testKey.String(), testPassphrase)
	require.NoError(t, err)

	_, err = db.GetPrivKey(testKey.Address().String(), testPassphrase)
	require.NoError(t, err)

	err = db.UpdatePassphrase(testKey.Address().String(), testNewPassphrase, testNewPassphrase)
	require.ErrorIs(t, err, ErrorWrongPassphrase)
}

func TestBadgerKeybase_DeleteKey(t *testing.T) {
	db, err := initDB()
	defer db.Stop()
	require.NoError(t, err)

	err = db.CreateFromString(testKey.String(), testPassphrase)
	require.NoError(t, err)

	_, err = db.GetPrivKey(testKey.Address().String(), testPassphrase)
	require.NoError(t, err)

	err = db.Delete(testKey.Address().String(), testPassphrase)
	require.NoError(t, err)

	kp, err := db.Get(testKey.Address().String())
	require.EqualError(t, err, ErrorAddrNotFound(testKey.Address().String()).Error())
	require.Equal(t, kp, KeyPair{})
}

func TestBadgerKeybase_DeleteKeyWrongPassphrase(t *testing.T) {
	db, err := initDB()
	defer db.Stop()
	require.NoError(t, err)

	err = db.CreateFromString(testKey.String(), testPassphrase)
	require.NoError(t, err)

	_, err = db.GetPrivKey(testKey.Address().String(), testPassphrase)
	require.NoError(t, err)

	err = db.Delete(testKey.Address().String(), testNewPassphrase)
	require.ErrorIs(t, err, ErrorWrongPassphrase)
}

func initDB() (Keybase, error) {
	db, err := NewKeybaseInMemory("")
	if err != nil {
		return nil, err
	}
	return db, nil
}

func createTestKey() (crypto.PrivateKey, error) {
	return crypto.GeneratePrivateKey()
}
