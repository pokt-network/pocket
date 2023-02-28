// Keybase using HashiCorp vault
package keybase

import (
	"context"
	"errors"
	"fmt"

	"github.com/dgraph-io/badger/v3"
	"github.com/hashicorp/vault/api"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/crypto/slip"
)

// vaultKeybase implements the Keybase interface using HashiCorp Vault.
type vaultKeybase struct {
	client *api.Client
	mount  string
}

// vaultKeybaseConfig contains the configuration parameters for the VaultKeybase.
type vaultKeybaseConfig struct {
	Address string
	Token   string
	Mount   string
}

// NewVaultKeybase returns a new instance of vaultKeybase.
func NewVaultKeybase(config vaultKeybaseConfig) (*vaultKeybase, error) {
	apiConfig := api.DefaultConfig()

	// Set default values for the configuration parameters
	if config.Address == "" {
		config.Address = apiConfig.Address
	}

	// Set the default mount path for the secret engine
	if config.Mount == "" {
		config.Mount = "secret"
	}

	// Create a new vault API client
	client, err := api.NewClient(&api.Config{
		Address: config.Address,
	})
	if err != nil {
		return nil, err
	}

	// Set the root token for the client if provided
	if config.Token != "" {
		client.SetToken(config.Token)
	}

	// Create a new VaultKeybase instance
	vk := &vaultKeybase{
		client: client,
		mount:  config.Mount,
	}

	return vk, nil
}

// GetBadgerDB TODO: Drop this once we have a proper keybase abstraction
func (vk *vaultKeybase) GetBadgerDB() (*badger.DB, error) {
	return nil, errors.New("not implemented")
}

// Stop the vault client by clearing the token
func (vk *vaultKeybase) Stop() error {
	vk.client.ClearToken()
	return nil
}

// Create new keypair entry in vault
func (vk *vaultKeybase) Create(passphrase, hint string) (crypto.KeyPair, error) {
	keyPair, err := crypto.CreateNewKey(passphrase, hint)
	if err != nil {
		return nil, err
	}
	err = writeVaultKeyPair(vk, keyPair.GetAddressString(), keyPair, hint)
	if err != nil {
		return nil, err
	}
	return keyPair, nil
}

// ImportFromString a new keypair from the private key hex string provided into vault
func (vk *vaultKeybase) ImportFromString(privStr, passphrase, hint string) (crypto.KeyPair, error) {
	keyPair, err := crypto.CreateNewKeyFromString(privStr, passphrase, hint)
	if err != nil {
		return nil, err
	}
	err = writeVaultKeyPair(vk, keyPair.GetAddressString(), keyPair, hint)
	if err != nil {
		return nil, err
	}
	return keyPair, nil
}

// ImportFromJSON Import a new keypair from the JSON string of the encrypted private key into vault
func (vk *vaultKeybase) ImportFromJSON(jsonStr, passphrase string) (crypto.KeyPair, error) {
	keyPair, err := crypto.ImportKeyFromJSON(jsonStr, passphrase)
	if err != nil {
		return nil, err
	}
	err = writeVaultKeyPair(vk, keyPair.GetAddressString(), keyPair, "")
	if err != nil {
		return nil, err
	}
	return keyPair, nil
}

// Get a keypair from vault
func (vk *vaultKeybase) Get(address string) (crypto.KeyPair, error) {
	data, err := vk.client.KVv2(vk.mount).Get(context.TODO(), address)
	if err != nil {
		return nil, err
	}

	if data == nil {
		return nil, errors.New("key not found")
	}

	keyPairStr, ok := data.Data["key_pair"].(string)
	if !ok {
		return nil, errors.New("invalid key pair")
	}

	keyPairStruct := crypto.GetKeypair()
	err = keyPairStruct.UnmarshalJSON([]byte(keyPairStr))
	if err != nil {
		return nil, err
	}

	return keyPairStruct, nil
}

// GetPubKey Get a public key from vault
func (vk *vaultKeybase) GetPubKey(address string) (crypto.PublicKey, error) {
	keyPair, err := vk.Get(address)
	if err != nil {
		return nil, err
	}

	return keyPair.GetPublicKey(), nil
}

// GetPrivKey Get a private key from vault
func (vk *vaultKeybase) GetPrivKey(address, passphrase string) (crypto.PrivateKey, error) {
	keyPair, err := vk.Get(address)
	if err != nil {
		return nil, err
	}

	privKey, err := keyPair.Unarmour(passphrase)
	if err != nil {
		return nil, err
	}

	return privKey, nil
}

// GetAll get all keys at this path, NOTE: these may not all be keypairs so practice good hygiene
func (vk *vaultKeybase) GetAll() ([]string, []crypto.KeyPair, error) {
	data, err := vk.client.Logical().List(fmt.Sprintf("%s/metadata", vk.mount))

	if err != nil {
		return nil, nil, err
	}

	if data == nil {
		return nil, nil, nil
	}

	var addresses []string
	var keyPairs []crypto.KeyPair

	for _, key := range data.Data["keys"].([]any) {
		addresses = append(addresses, key.(string))
		keyPair, err := vk.Get(key.(string))
		if err != nil {
			return nil, nil, err
		}
		keyPairs = append(keyPairs, keyPair)
	}

	return addresses, keyPairs, nil
}

// DeriveChildFromSeed deterministically generates and return the child at the given index from the seed provided
// By default this stores the key in the keybase and returns the KeyPair interface and any error
func (vk *vaultKeybase) DeriveChildFromSeed(seed []byte, childIndex uint32, childPassphrase, childHint string, shouldStore bool) (crypto.KeyPair, error) {
	path := fmt.Sprintf(slip.PoktAccountPathFormat, childIndex)
	childKey, err := slip.DeriveChild(path, seed)
	if err != nil {
		return nil, err
	}

	if !shouldStore {
		return childKey, nil
	}

	keyPair := childKey
	// Re-encrypt child key with passphrase and hint
	if childPassphrase != "" && childHint != "" {
		// Get the private key hex string from the child key
		privKeyHex, err := childKey.ExportString("") // No passphrase by default
		if err != nil {
			return nil, err
		}

		keyPair, err = crypto.CreateNewKeyFromString(privKeyHex, childPassphrase, childHint)
		if err != nil {
			return nil, err
		}
	}

	// Use key address as key in DB
	addrKey := keyPair.GetAddressString()

	writeVaultKeyPair(vk, addrKey, keyPair, childHint)

	return childKey, nil
}

// DeriveChildFromKey deterministically generates and return the child at the given index from the parent key provided
// By default this stores the key in the keybase and returns the KeyPair interface and any error
func (vk *vaultKeybase) DeriveChildFromKey(masterAddrHex, passphrase string, childIndex uint32, childPassphrase, childHint string, shouldStore bool) (crypto.KeyPair, error) {
	privKey, err := vk.GetPrivKey(masterAddrHex, passphrase)
	if err != nil {
		return nil, err
	}
	seed := privKey.Seed()
	return vk.DeriveChildFromSeed(seed, childIndex, childPassphrase, childHint, shouldStore)
}

// ExportPrivString exports a private key as a hex string
func (vk *vaultKeybase) ExportPrivString(address, passphrase string) (string, error) {
	privKey, err := vk.Get(address)
	if err != nil {
		return "", err
	}

	privKeyHex, err := privKey.ExportString(passphrase)
	if err != nil {
		return "", err
	}

	return privKeyHex, nil
}

// ExportPrivJSON exports a private key as a JSON string
func (vk *vaultKeybase) ExportPrivJSON(address, passphrase string) (string, error) {
	privKey, err := vk.Get(address)
	if err != nil {
		return "", err
	}

	privKeyJSON, err := privKey.ExportJSON(passphrase)
	if err != nil {
		return "", err
	}

	return privKeyJSON, nil
}

// UpdatePassphrase updates the passphrase of a key
func (vk *vaultKeybase) UpdatePassphrase(address, oldPassphrase, newPassphrase, hint string) error {
	privKey, err := vk.GetPrivKey(address, oldPassphrase)
	if err != nil {
		return err
	}
	privStr := privKey.String()

	keyPair, err := crypto.CreateNewKeyFromString(privStr, newPassphrase, hint)
	if err != nil {
		return err
	}

	return writeVaultKeyPair(vk, keyPair.GetAddressString(), keyPair, hint)
}

// Sign a message using a private key
func (vk *vaultKeybase) Sign(address, passphrase string, msg []byte) ([]byte, error) {
	privKey, err := vk.GetPrivKey(address, passphrase)
	if err != nil {
		return nil, err
	}

	sig, err := privKey.Sign(msg)
	if err != nil {
		return nil, err
	}

	return sig, nil
}

// Verify a message signature using a public key
func (vk *vaultKeybase) Verify(address string, msg, sig []byte) (bool, error) {
	pubKey, err := vk.GetPubKey(address)
	if err != nil {
		return false, err
	}

	return pubKey.Verify(msg, sig), nil
}

// Delete a keypair from vault
func (vk *vaultKeybase) Delete(address, passphrase string) error {

	_, err := vk.GetPrivKey(address, passphrase)
	if err != nil {
		return err
	}

	versionsMeta, err := vk.client.KVv2(vk.mount).GetVersionsAsList(context.TODO(), address)
	if err != nil {
		return err
	}

	versions := make([]int, 0, len(versionsMeta))
	for _, version := range versionsMeta {
		versions = append(versions, version.Version)
	}

	return vk.client.KVv2(vk.mount).Destroy(context.TODO(), address, versions)
}

// writeVaultKeyPair writes a keypair to vault
func writeVaultKeyPair(vk *vaultKeybase, address string, kp crypto.KeyPair, hint string) error {
	dataBz, err := kp.MarshalJSON()
	if err != nil {
		return err
	}

	_, err = vk.client.KVv2(vk.mount).Put(context.TODO(), address, map[string]interface{}{
		"key_pair": string(dataBz),
		"hint":     hint,
	})

	return err
}
