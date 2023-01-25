package crypto

// The KeyPair interface exposes functions relating to public and encrypted private key pairs
type KeyPair interface {
	// Accessors
	GetPublicKey() PublicKey
	GetPrivArmour() string

	// Public Key Address
	GetAddressBytes() []byte
	GetAddressString() string // hex string

	// Unarmour
	Unarmour(passphrase string) (PrivateKey, error)

	// Export
	ExportString(passphrase string) (string, error)
	ExportJSON(passphrase string) (string, error)
}

var _ KeyPair = &EncKeyPair{}

// EncKeyPair struct stores the public key and the passphrase encrypted private key
type EncKeyPair struct {
	PublicKey     PublicKey
	PrivKeyArmour string
}

// Generate a new KeyPair struct given the public key and armoured private key
func NewKeyPair(pub PublicKey, priv string) KeyPair {
	return EncKeyPair{
		PublicKey:     pub,
		PrivKeyArmour: priv,
	}
}

// Return the public key
func (kp EncKeyPair) GetPublicKey() PublicKey {
	return kp.PublicKey
}

// Return private key armoured string
func (kp EncKeyPair) GetPrivArmour() string {
	return kp.PrivKeyArmour
}

// Return the byte slice address of the public key
func (kp EncKeyPair) GetAddressBytes() []byte {
	return kp.PublicKey.Address().Bytes()
}

// Return the string address of the public key
func (kp EncKeyPair) GetAddressString() string {
	return kp.PublicKey.Address().String()
}

// Unarmour the private key with the passphrase provided
func (kp EncKeyPair) Unarmour(passphrase string) (PrivateKey, error) {
	return unarmourDecryptPrivKey(kp.PrivKeyArmour, passphrase)
}

// Export Private Key String
func (kp EncKeyPair) ExportString(passphrase string) (string, error) {
	privKey, err := unarmourDecryptPrivKey(kp.PrivKeyArmour, passphrase)
	if err != nil {
		return "", err
	}
	return privKey.String(), nil
}

// Export Private Key as armoured JSON string with fields to decrypt
func (kp EncKeyPair) ExportJSON(passphrase string) (string, error) {
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
		return nil, err
	}

	privArmour, err := encryptArmourPrivKey(privKey, passphrase)
	if err != nil || privArmour == "" {
		return nil, err
	}

	pubKey := privKey.PublicKey()
	kp := NewKeyPair(pubKey, privArmour)

	return kp, nil
}

// Generate new KeyPair from the hex string provided, encrypt and armour it as a string
func CreateNewKeyFromString(privStrHex, passphrase string) (KeyPair, error) {
	privKey, err := NewPrivateKey(privStrHex)
	if err != nil {
		return nil, err
	}

	privArmour, err := encryptArmourPrivKey(privKey, passphrase)
	if err != nil || privArmour == "" {
		return nil, err
	}

	pubKey := privKey.PublicKey()
	kp := NewKeyPair(pubKey, privArmour)

	return kp, nil
}

// Create new KeyPair from the JSON encoded privStr
func ImportKeyFromJSON(jsonStr, passphrase string) (KeyPair, error) {
	// Get Private Key from armouredStr
	privKey, err := unarmourDecryptPrivKey(jsonStr, passphrase)
	if err != nil {
		return nil, err
	}
	pubKey := privKey.PublicKey()
	kp := NewKeyPair(pubKey, jsonStr)

	return kp, nil
}
