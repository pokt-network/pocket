# Keybase <!-- omit in toc -->

This document is intended to outline the current Keybase implementation used by the V1 client, and is primarily focused on its design and implementation as well as testing.

- [Backend Database](#backend-database)
- [Keybase Interface](#keybase-interface)
  - [V0\<-\>V1 Interoperability](#v0-v1-interoperability)
  - [Keybase Code Structure](#keybase-code-structure)
- [Makefile Testing Helper](#makefile-testing-helper)
- [KeyPair Encryption \& Armouring](#keypair-encryption--armouring)
- [TODO: Future Work](#todo-future-work)

_TODO(#150): The current keybase has not been integrated with any CLI endpoints, and as such is only accessible through the [keybase interface](#keybase-interface)_

## Backend Database

The Keybase package uses a filesystem key-value database, `BadgerDB`, as its backend to persistently store keys locally on the client machine. The DB stores the local keys encoded as `[]byte` using `encoding/gob`.

The `KeyPair` defined in [crypto package](../../../shared/crypto) is the data structure that's stored in the DB. Specifically:

- **Key**: The `[]byte` returned by the `GetAddressBytes()` function is used as the key in the key-value store.
- **Value**: The `gob` encoded struct of the entire `KeyPair`, containing both the `PublicKey` and `PrivKeyArmour` (JSON encoded, encrypted private key string), is the value.

The Keybase DB layer exposes several functions, defined by the [Keybase interface](#keybase-interface), to fulfill CRUD operations on the DB itself and oeprate with the Keypairs.

## Keybase Interface

The [Keybase interface](./keybase.go) exposes the CRUD operations to operate on keys, and supports the following operations:

- Create password protected private keys
- Export/Import string/json keypairs
- Retrieve public/private keys or keypairs
- List all keys stored
- Check keys exist in the keybase
- Update passphrase on a private key
- Message signing and verification

### V0<->V1 Interoperability

The `Keybase` interface supports full interoperability of key export & import between Pocket [V0](https://github.com/pokt-network/pocket-core)<->[V1](https://github.com/pokt-network/pocket).

Any private key created in the V0 protocol can be imported into V1 via one of the following two ways:

1. **JSON keyfile**: This method will take the JSON encoded, encrypted private key, and will import it into the V1 keybase. The `passphrase` supplied must be the same as the one use to encrypt the key in the first place or the key won't be importable.

2. **Private Key Hex String**: This method will directly import the private key from the hex string provided and encrypt it with the passphrase provided. This enables the passphrase to be different from the original as the provided plaintext is already decrypted.

Although key pairs are stored in the local DB using the serialized (`[]byte`) representation of the public key, the associated address can be used for accessing the record in the DB for simplicity.

Keys can be created without a password by specifying an empty (`""`) passphrase. The private key will still be encrypted at rest but will use the empty string as the passphrase for decryption.

### Keybase Code Structure

```bash
app
└── client
    └── keybase
          ├── README.md
          ├── keybase.go
          ├── keybase_test.go
          └── keystore.go
```

The interface is found in [keybase.go](./keybase.go) whereas its implementation can be found in [keystore.go](./keystore.go)

## Makefile Testing Helper

The unit tests for the keybase are defined in [keybase_test.go](./keybase_test.go) and can therefore be executed alongside other application specific tests by running `make test_app`.

## KeyPair Encryption & Armouring

The [documentation in the crypto library](../../../shared/crypto/README.md) covers all of the details related to the `KeyPair` interface, as well as `PrivateKey` encryption, armouring and unarmouring.

The primitives and functions defined there are heavily used throughout this package.

## TODO: Future Work

- [ ] Improve error handling and error messages for importing keys with invalid strings/invalid JSON
- [ ] Research and implement threshold signatures and threshold keys
- [ ] Look into a fully feature signature implementation beyond trivial `[]byte` messages
- [ ] Integrate the keybase with the CLI (#150)
