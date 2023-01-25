package crypto

// KeyPair struct stores the public key and the passphrase encrypted private key
type KeyPair struct {
	PublicKey     PublicKey
	PrivKeyArmour string
}

// Generate a new KeyPair struct given the public key and armoured private key
func NewKeyPair(pub PublicKey, priv string) KeyPair {
	return KeyPair{
		PublicKey:     pub,
		PrivKeyArmour: priv,
	}
}

// Return the byte slice address of the public key
func (kp KeyPair) GetAddressBytes() []byte {
	return kp.PublicKey.Address().Bytes()
}

// Return the string address of the public key
func (kp KeyPair) GetAddressString() string {
	return kp.PublicKey.Address().String()
}

// Unarmour the private key with the passphrase provided
func (kp KeyPair) Unarmour(passphrase string) (PrivateKey, error) {
	return unarmourDecryptPrivKey(kp.PrivKeyArmour, passphrase)
}

// Export Private Key String
func (kp KeyPair) ExportString(passphrase string) (string, error) {
	privKey, err := unarmourDecryptPrivKey(kp.PrivKeyArmour, passphrase)
	if err != nil {
		return "", err
	}
	return privKey.String(), nil
}

// Export Private Key as armoured JSON string with fields to decrypt
func (kp KeyPair) ExportJSON(passphrase string) (string, error) {
	_, err := unarmourDecryptPrivKey(kp.PrivKeyArmour, passphrase)
	if err != nil {
		return "", err
	}
	return kp.PrivKeyArmour, nil
}

// Generate new private ED25519 key and encrypt and armour it as a string
// Returns a KeyPair struct of the Public Key and Armoured String
func CreateNewKey(passphrase string) (KeyPair, error) {
	privKey, err := GeneratePrivateKey()
	if err != nil {
		return KeyPair{}, err
	}

	privArmour, err := encryptArmourPrivKey(privKey, passphrase)
	if err != nil || privArmour == "" {
		return KeyPair{}, err
	}

	pubKey := privKey.PublicKey()
	keyPair := NewKeyPair(pubKey, privArmour)

	return keyPair, nil
}

// Generate new KeyPair from the hex string provided, encrypt and armour it as a string
func CreateNewKeyFromString(privStrHex, passphrase string) (KeyPair, error) {
	privKey, err := NewPrivateKey(privStrHex)
	if err != nil {
		return KeyPair{}, err
	}

	privArmour, err := encryptArmourPrivKey(privKey, passphrase)
	if err != nil || privArmour == "" {
		return KeyPair{}, err
	}

	pubKey := privKey.PublicKey()
	keyPair := NewKeyPair(pubKey, privArmour)

	return keyPair, nil
}

// Create new KeyPair from the JSON encoded privStr
func ImportKeyFromJSON(jsonStr, passphrase string) (KeyPair, error) {
	// Get Private Key from armouredStr
	privKey, err := unarmourDecryptPrivKey(jsonStr, passphrase)
	if err != nil {
		return KeyPair{}, err
	}
	pubKey := privKey.PublicKey()
	keyPair := NewKeyPair(pubKey, jsonStr)

	return keyPair, nil
}
