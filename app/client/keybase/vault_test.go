package keybase

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	vk *vaultKeybase
)

func setupTestVaultKeybase(address string) (*vaultKeybase, error) {
	return NewVaultKeybase(vaultKeybaseConfig{
		Address: address,
		Token:   "dev-only-token",
		Mount:   "secret",
	})
}

func TestMain(m *testing.M) {
	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not construct pool: %s", err)
	}

	// uses pool to try to connect to Docker
	err = pool.Client.Ping()
	if err != nil {
		log.Fatalf("Could not connect to Docker: %s", err)
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "vault",
		Tag:        "latest",
		Env: []string{
			"VAULT_DEV_ROOT_TOKEN_ID=dev-only-token",
			"VAULT_DEV_LISTEN_ADDRESS=0.0.0.0:8200",
		},
		ExposedPorts: []string{"8200/tcp"},
		PortBindings: map[docker.Port][]docker.PortBinding{
			"8200/tcp": {
				{HostIP: "127.0.0.1", HostPort: "8200"},
			},
		},
		CapAdd: []string{"IPC_LOCK"},
	}, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})

	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	err = resource.Expire(120) // Tell docker to hard kill the container in 2 minutes
	if err != nil {
		log.Fatalf("Could not set expiration on resource: %s", err)
	}

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if err = pool.Retry(func() error {

		// get the port from the container
		endpoint := resource.GetHostPort("8200/tcp")

		// test connection to vault
		vk, err = setupTestVaultKeybase(fmt.Sprintf("http://%s", endpoint))

		if err != nil {
			return err
		}

		_, err = vk.client.Sys().Health()

		return err
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

func TestVaultKeybase(t *testing.T) {

	if vk == nil {
		require.NoErrorf(t, fmt.Errorf("vault keybase is nil"), "vault keybase is nil")
	}

	// Test Create
	_, err := vk.Create("passphrase", "hint")
	require.NoErrorf(t, err, "error creating keypair: %s", err)

	// Generate a new key pair
	privKey, err := crypto.GeneratePrivateKey()
	require.NoErrorf(t, err, "error generating key pair: %s", err)

	// Test ImportFromString
	_, err = vk.ImportFromString(privKey.String(), "passphrase", "hint")
	require.NoErrorf(t, err, "error importing keypair: %s")

	// Test GetPrivKey
	origPrivKey := privKey.String()
	privKey, err = vk.GetPrivKey(privKey.Address().String(), "passphrase")
	require.NoErrorf(t, err, "error getting private key: %s")
	assert.Equal(t, origPrivKey, privKey.String())

	// Test ImportFromJSON
	keyPair, err := crypto.CreateNewKey("passphrase2", "hint")
	require.NoErrorf(t, err, "error creating key pair: %s")
	keyPairBytes, err := keyPair.ExportJSON("passphrase2")
	require.NoErrorf(t, err, "error marshaling key pair: %s")
	importedKp, err := vk.ImportFromJSON(string(keyPairBytes), "passphrase2")
	require.NoErrorf(t, err, "error importing keypair: %s")
	assert.Equal(t, keyPair, importedKp)

	// Test Get
	keyPair2, err := vk.Get(keyPair.GetAddressString())
	require.NoErrorf(t, err, "error getting keypair: %s")
	assert.Equal(t, keyPair, keyPair2)

	// Test GetPubKey
	pubKey, err := vk.GetPubKey(keyPair.GetAddressString())
	require.NoErrorf(t, err, "error getting public key: %s")
	assert.Equal(t, keyPair.GetPublicKey(), pubKey)

	// Test GetAll
	addresses, keyPairs, err := vk.GetAll()
	require.NoErrorf(t, err, "error getting all keypairs: %s")
	assert.Equal(t, 3, len(addresses))
	assert.Equal(t, 3, len(keyPairs))

	// Test ExportPrivString
	privKeyStr, err := vk.ExportPrivString(privKey.Address().String(), "passphrase")
	require.NoErrorf(t, err, "error exporting private key: %s")
	assert.Equal(t, privKey.String(), privKeyStr)

	// Test ExportPrivJSON
	_, err = vk.ExportPrivJSON(keyPair.GetAddressString(), "passphrase2")
	require.NoErrorf(t, err, "error exporting private key: %s")

	// Test UpdatePassphrase
	err = vk.UpdatePassphrase(keyPair.GetAddressString(), "passphrase2", "new-passphrase", "hint")
	require.NoErrorf(t, err, "error updating passphrase: %s")

	// // Test Sign
	msg := []byte("hello world")
	sig, err := vk.Sign(keyPair.GetAddressString(), "new-passphrase", msg)
	require.NoErrorf(t, err, "error signing message: %s")

	// Test Verify
	verified, err := vk.Verify(keyPair.GetAddressString(), msg, sig)
	require.NoErrorf(t, err, "error verifying signature: %s")

	if !verified {
		require.NoErrorf(t, fmt.Errorf("signature not verified"), "signature not verified")
	}

	err = vk.Delete(keyPair.GetAddressString(), "new-passphrase")
	require.NoErrorf(t, err, "error deleting keypair: %s")
}
