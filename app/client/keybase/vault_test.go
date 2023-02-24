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

	// pool.MaxWait = 20 // ain't nobody got time for that

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
		t.Fatalf("vk is nil")
	}

	// Test Create
	err := vk.Create("passphrase", "hint")
	if err != nil {
		t.Fatalf("error creating keypair: %s", err)
	}

	// Generate a new key pair
	privKey, err := crypto.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("error generating key pair: %s", err)
	}

	// Test ImportFromString
	err = vk.ImportFromString(privKey.String(), "passphrase", "hint")
	if err != nil {
		t.Fatalf("error importing keypair: %s", err)
	}

	// Test GetPrivKey
	origPrivKey := privKey.String()
	privKey, err = vk.GetPrivKey(privKey.Address().String(), "passphrase")
	if err != nil {
		t.Fatalf("error getting private key: %s", err)
	}
	assert.Equal(t, origPrivKey, privKey.String())

	// Test ImportFromJSON
	keyPair, err := crypto.CreateNewKey("passphrase2", "hint")
	if err != nil {
		t.Fatalf("error creating key pair: %s", err)
	}
	keyPairBytes, err := keyPair.ExportJSON("passphrase2")
	if err != nil {
		t.Fatalf("error marshaling key pair: %s", err)
	}

	err = vk.ImportFromJSON(string(keyPairBytes), "passphrase2")
	if err != nil {
		t.Fatalf("error importing keypair: %s", err)
	}

	// Test Get
	keyPair2, err := vk.Get(keyPair.GetAddressString())
	if err != nil {
		t.Fatalf("error getting keypair: %s", err)
	}
	assert.Equal(t, keyPair, keyPair2)

	// Test GetPubKey
	pubKey, err := vk.GetPubKey(keyPair.GetAddressString())
	if err != nil {
		t.Fatalf("error getting public key: %s", err)
	}
	assert.Equal(t, keyPair.GetPublicKey(), pubKey)

	// Test GetAll
	addresses, keyPairs, err := vk.GetAll()
	if err != nil {
		t.Fatalf("error getting all keypairs: %s", err)
	}
	assert.Equal(t, 3, len(addresses))
	assert.Equal(t, 3, len(keyPairs))

	// Test Exists
	_, err = vk.Exists(keyPair.GetAddressString())
	if err != nil {
		t.Fatalf("error checking key exists: %s", err)
	}

	// Test ExportPrivString
	privKeyStr, err := vk.ExportPrivString(privKey.Address().String(), "passphrase")
	if err != nil {
		t.Fatalf("error exporting private key: %s", err)
	}
	assert.Equal(t, privKey.String(), privKeyStr)

	// Test ExportPrivJSON
	_, err = vk.ExportPrivJSON(keyPair.GetAddressString(), "passphrase2")
	if err != nil {
		t.Fatalf("error exporting private key: %s", err)
	}

	// Test UpdatePassphrase
	err = vk.UpdatePassphrase(keyPair.GetAddressString(), "passphrase2", "new-passphrase", "hint")
	if err != nil {
		t.Fatalf("error updating passphrase: %s", err)
	}

	// // Test Sign
	msg := []byte("hello world")
	sig, err := vk.Sign(keyPair.GetAddressString(), "new-passphrase", msg)
	if err != nil {
		t.Fatalf("error signing message: %s", err)
	}

	// Test Verify
	verified, err := vk.Verify(keyPair.GetAddressString(), msg, sig)
	if err != nil {
		t.Fatalf("error verifying signature: %s", err)
	}

	if !verified {
		t.Fatalf("signature verification failed")
	}

	err = vk.Delete(keyPair.GetAddressString(), "new-passphrase")
	if err != nil {
		t.Fatalf("error deleting keypair: %s", err)
	}
}
