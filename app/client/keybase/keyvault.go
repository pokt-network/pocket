package keybase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hashicorp/vault/api"
	"github.com/pokt-network/pocket/shared/crypto"
)

// vaultKeybase implements the Keybase interface using HashiCorp Vault
type vaultKeybase struct {
	client *api.Client
	mount  string
}

// NewVaultKeybase returns a new instance of vaultKeybase
func NewVaultKeybase(client *api.Client, mount string) *vaultKeybase {
	return &vaultKeybase{
		client: client,
		mount:  mount,
	}
}

// Create new keypair entry in Vault
func (vk *vaultKeybase) Create(passphrase, hint string) error {
	keyPair, err := crypto.CreateNewKey(passphrase, hint)
	if err != nil {
		return err
	}

	keyPairBytes, err := json.Marshal(keyPair)
	if err != nil {
		return err
	}

	data := map[string]interface{}{
		"key_pair": string(keyPairBytes),
		"hint":     hint,
	}

	_, err = vk.client.KVv2(vk.mount).Put(context.TODO(), keyPair.GetAddressString(), data)
	if err != nil {
		return err
	}

	return nil
}

// Import a new keypair from the private key hex string provided into Vault
func (vk *vaultKeybase) ImportFromString(privStr, passphrase, hint string) error {
	keyPair, err := crypto.CreateNewKeyFromString(privStr, passphrase, hint)
	if err != nil {
		return err
	}

	keyPairBytes, err := json.Marshal(keyPair)
	if err != nil {
		return err
	}

	data := map[string]interface{}{
		"key_pair": string(keyPairBytes),
		"hint":     hint,
	}

	_, err = vk.client.KVv2(vk.mount).Put(context.TODO(), keyPair.GetAddressString(), data)
	if err != nil {
		return err
	}

	return nil
}

// Import a new keypair from the JSON string of the encrypted private key into Vault
func (vk *vaultKeybase) ImportFromJSON(jsonStr, passphrase string) error {
	keyPair, err := crypto.ImportKeyFromJSON(jsonStr, passphrase)
	if err != nil {
		return err
	}

	keyPairBytes, err := json.Marshal(keyPair)
	if err != nil {
		return err
	}

	data := map[string]interface{}{
		"key_pair": string(keyPairBytes),
	}

	_, err = vk.client.KVv2(vk.mount).Put(context.TODO(), keyPair.GetAddressString(), data)
	if err != nil {
		return err
	}

	return nil
}

// Get a keypair from Vault
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

	var keyPairStruct crypto.EncKeyPair
	err = json.Unmarshal([]byte(keyPairStr), &keyPairStruct)
	if err != nil {
		return nil, err
	}

	return &keyPairStruct, nil
}

// Get a public key from Vault
func (vk *vaultKeybase) GetPubKey(address string) (crypto.PublicKey, error) {
	keyPair, err := vk.Get(address)
	if err != nil {
		return nil, err
	}

	return keyPair.GetPublicKey(), nil
}

// Get a private key from Vault
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

// Get all keypairs from Vault
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

	for _, key := range data.Data["keys"].([]interface{}) {
		addresses = append(addresses, key.(string))
		keyPair, err := vk.Get(key.(string))
		if err != nil {
			return nil, nil, err
		}
		keyPairs = append(keyPairs, keyPair)
	}

	return addresses, keyPairs, nil
}

// Check if a key exists in Vault
func (vk *vaultKeybase) Exists(address string) (bool, error) {
	_, err := vk.Get(address)
	if err != nil {
		if errors.Is(err, errors.New("key not found")) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// Export a private key as a hex string
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

// Export a private key as a JSON string
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

// Update the passphrase of a key
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

	addrKey := keyPair.GetAddressString()
	if addrKey != address {
		return errors.New("address mismatch")
	}

	keyPairBz, err := keyPair.Marshal()
	if err != nil {
		return err
	}

	return vk.ImportFromJSON(string(keyPairBz), newPassphrase)
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

// Delete a keypair from Vault
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

	err = vk.client.KVv2(vk.mount).Destroy(context.TODO(), address, versions)
	if err != nil {
		return err
	}

	return nil
}
