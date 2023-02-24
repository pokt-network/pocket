package keybase

import (
	"context"
	"fmt"
	"testing"

	vault "github.com/hashicorp/vault/api"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestVaultKeybase(address string) (*vault.Client, *vaultKeybase, error) {
	config := vault.DefaultConfig()

	config.Address = address

	// Create a new Vault API client
	client, err := vault.NewClient(config)
	if err != nil {
		return nil, nil, err
	}

	// Set the root token for the client
	client.SetToken("dev-only-token")

	// Create a new VaultKeybase instance
	vk := NewVaultKeybase(client, "secret")

	return client, vk, nil
}

func TestVaultKeybase(t *testing.T) {

	ctx := context.Background()

	// Create a new Vault container
	req := testcontainers.ContainerRequest{
		Image:        "vault:latest",
		ExposedPorts: []string{"8200/tcp"},
		Env: map[string]string{
			"VAULT_DEV_ROOT_TOKEN_ID":  "dev-only-token",
			"VAULT_DEV_LISTEN_ADDRESS": "0.0.0.0:8200",
		},
		CapAdd:     []string{"IPC_LOCK"},
		WaitingFor: wait.ForLog("core: successful mount"),
	}
	vaultC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Error(err)
	}
	defer func() {
		if err := vaultC.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err.Error())
		}
	}()

	mappedPort, err := vaultC.MappedPort(ctx, "8200")
	if err != nil {
		t.Fatalf("failed to get mapped port: %s", err.Error())
	}

	hostIP, err := vaultC.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get host IP: %s", err.Error())
	}

	uri := fmt.Sprintf("http://%s:%s", hostIP, mappedPort.Port())
	_, vk, err := setupTestVaultKeybase(uri)
	if err != nil {
		t.Fatalf("error setting up test: %s", err)
	}

	// Test Create
	err = vk.Create("passphrase", "hint")
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
	assert.NotEmpty(t, addresses)
	assert.NotEmpty(t, keyPairs)
	// assert.Equal(t, 2, len(addresses))
	// assert.Equal(t, 2, len(keyPairs))

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

	// // Test UpdatePassphrase
	// err = vk.UpdatePassphrase("key1", "passphrase", "new-passphrase", "hint")
	// if err != nil {
	// 	t.Fatalf("error updating passphrase: %s", err)
	// }

	// // Test Sign
	// msg := []byte("hello world")
	// _, err = vk.Sign("key1", "new-passphrase", msg)
	// if err != nil {
	// 	t.Fatalf("error signing message: %s", err)
	// }

	// // Test Verify
	// sig, err := vk.Sign("key1", "new-passphrase", msg)
	// if err != nil {
	// 	t.Fatalf("error signing message: %s", err)
	// }

	// verified, err := vk.Verify("key1", msg, sig)
	// if err != nil {
	// 	t.Fatalf("error verifying signature: %s", err)
	// }

	// if !verified {
	// 	t.Fatalf("signature verification failed")
	// }

	// err = vk.Delete("key1", "new-passphrase")
	// if err != nil {
	// 	t.Fatalf("error deleting keypair: %s", err)
	// }
}
