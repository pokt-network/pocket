# Keybase <!-- omit in toc -->

This document is intended to outline the current Keybase implementation used by the V1 client, and is primarily focused on its design and implementation as well as testing.

- [Backend Options](#backend-options)
- [Keybase Interface](#keybase-interface)
  - [V0\<-\>V1 Interoperability](#v0-v1-interoperability)
  - [Keybase Code Structure](#keybase-code-structure)
- [Configuration Methods](#configuration-methods)
- [Makefile Testing Helper](#makefile-testing-helper)
- [KeyPair Encryption \& Armouring](#keypair-encryption--armouring)
- [SLIP-0010 Child Key Generation](#slip-0010-child-key-generation)
- [TODO: Future Work](#todo-future-work)

## Backend Options

The Keybase package supports multiple backend options to store keys:

- **BadgerDB**: A filesystem key-value database used to persistently store keys locally on the client machine. The DB stores the local keys encoded as `[]byte` using `encoding/gob`. It is the default backend for the Keybase.
- **Hashicorp Vault**: An external secrets management system that can be used to store keys securely. The Vault backend requires additional configuration to connect and authenticate with a Vault server.

The backend option can be selected using the CLI or by configuring environment variables. Check the [Configuration Methods](#configuration-methods) section for more details. The key pairs are stored in the vault as an encoded JSON string using `encoding/json` at the vault mount path + the public address.

## Keybase Interface

The [Keybase interface](./keybase.go) exposes the CRUD operations to operate on keys, and supports the following operations:

- Create password protected private keys
- Export/Import string/json keypairs
- Retrieve public/private keys or keypairs
- List all keys stored
- Check keys exist in the keybase
- Update passphrase on a private key
- Message signing and verification

The `KeyPair` defined in [crypto package](../../../shared/crypto) is the data structure that's stored in the DB. Specifically:

- **Key**: The `[]byte` returned by the `GetAddressBytes()` function is used as the key in the key-value store.
- **Value**: The `gob` encoded struct of the entire `KeyPair`, containing both the `PublicKey` and `PrivKeyArmour` (JSON encoded, encrypted private key string), is the value.

### V0<->V1 Interoperability

The `Keybase` interface supports full interoperability of key export & import between Pocket [V0](https://github.com/pokt-network/pocket-core)<->[V1](https://github.com/pokt-network/pocket).

Any private key created in the V0 protocol can be imported into V1 via one of the following two ways:

1. **JSON keyfile**: This method will take the JSON encoded, encrypted private key, and will import it into the V1 keybase. The `passphrase` supplied must be the same as the one use to encrypt the key in the first place or the key won't be importable.

2. **Private Key Hex String**: This method will directly import the private key from the hex string provided and encrypt it with the passphrase provided. This enables the passphrase to be different from the original as the provided plaintext is already decrypted.

Although key pairs are stored in the local DB using the serialized (`[]byte`) representation of the public key, the associated address can be used for accessing the record in the DB for simplicity.

Keys can be created without a password by specifying an empty (`""`) passphrase. The private key will still be encrypted at rest but will use the empty string as the passphrase for decryption.

### Keybase Code Structure

```
app/client/keybase/
├── debug
│   └── keystore.go
├── doc
│   └── vault.md
├── hashicorp
│   ├── vault.go
│   └── vault_test.go
├── keybase.go
├── keybase_test.go
├── keystore.go
└── README.md
```

The interface is found in [keybase.go](./keybase.go) whereas its implementations can be found in:

- [keystore.go](./keystore.go) A keybase implementation that uses a filesystem badger database as its backend
- [vault.go](./hashicorp/vault.go) A keybase implementation that uses a Hashicorp vault as its backend

## Configuration Methods

To configure the Keybase to use a specific backend, you can use one of the following methods:

1. **Configuration File**: Create a configuration file with the following JSON structure:

   ```jsonc
   {
     "keybase": {
       "type": "file" // or "vault"
     }
   }
   ```

   Save this file and pass its path to the client using the `--config` flag. If a configuration is present in the file, you can override it by passing an environment variable prefixed with `POCKET_` such as `POCKET_KEYBASE_TYPE=vault`.

2. **Command-Line Flags**: When starting the client, pass the appropriate flags to specify the desired backend and its settings. For example, use `--keybase-type=vault`, `--vault-addr=http://127.0.0.1:8200`, `--vault-token=dev-only-token`, and `--vault-mount=secret`. Command-line flags have the highest precedence.

For the Hashicorp Vault backend, you need to configure the Vault connection and authentication. Please see a detailed explanation in the [Vault documentation](./doc/vault.md).

## Makefile Testing Helper

The unit tests the keybase are defined in:

- [keybase_test.go](./keybase_test.go)
- [vault_test.go](./vault_test.go)

They can be executed application specific tests by running `make test_app`.

## KeyPair Encryption & Armouring

The [documentation in the crypto library](../../../shared/crypto/README.md) covers all of the details related to the `KeyPair` interface, as well as `PrivateKey` encryption, armouring and unarmouring.

The primitives and functions defined there are heavily used throughout this package.

## Child Key Generation

The [documentation in the crypto library](../../../shared/crypto/README.md) covers the specifics of the [SLIPS-0010](https://github.com/satoshilabs/slips/blob/master/slip-0010.md) implementation related to child key generation from a single master key

## TODO: Future Work

- [ ] Improve error handling and error messages for importing keys with invalid strings/invalid JSON
- [ ] Research and implement threshold signatures and threshold keys
- [ ] Look into a fully feature signature implementation beyond trivial `[]byte` messages

<!-- GITHUB_WIKI: app/client/keybase -->
