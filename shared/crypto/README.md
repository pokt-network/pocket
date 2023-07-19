# Pocket Crypto <!-- omit in toc -->

- [KeyPair Interface](#keypair-interface)
  - [KeyPair Code Structure](#keypair-code-structure)
- [Encryption and Armouring](#encryption-and-armouring)
- [SLIP-0010 HD Child Key Generation](#slip-0010-hd-child-key-generation)

_DOCUMENT: Note that this README is a WIP and does not exhaustively document all the current types in this package_

## KeyPair Interface

The [KeyPair interface](./keypair.go) exposes methods related to operating on `PublicKey` types and `PrivKeyArmour` strings, such as:

- Retrieve the PublicKey or armoured PrivateKey JSON string
- Get PublicKey address `[]byte` or hex `string`
- Unarmour the PrivateKey JSON string
- Export the PrivateKey hex string or JSON as an armoured string
- Marshal or unmarshal the KeyPair to/from a `[]byte`

The [KeyPair](./keypair.go) interface is implemented by the `encKeyPair` struct which stores:

1. `PublicKey` of the KeyPair
2. `PrivateKey` armoured JSON string

The PrivateKey armoured JSON string is created after the [encryption step](#encryption-and-armouring) has encrypted the PrivateKey and marshalled it into a JSON string.

### KeyPair Code Structure

The KeyPair code is separated into two files: [keypair.go](./keypair.go) and [armour.go](./armour.go)

```bash
shared
└── crypto
    ├── armour.go
    └── keypair.go
```

## Encryption and Armouring

The passphrase provided or `""` (default) is used for encrypting and armouring new or imported keys.

Keys are encrypted using the `secretbox` library based on the NaCl (libsodium) primitives. Secretbox uses the `XSalsa20` stream cipher and `Poly1305` message authentication suite to encrypt and authenticate the key.

The following flowchart shows this process:

```mermaid
flowchart LR
    subgraph C[core lib]
        A["rand([16]byte)"]
    end
    subgraph S[scrypt lib]
        B["key(salt, pass, ...)"]
    end
    subgraph SecretBox
        direction TB
        D["Read(rand.Reader)"]
        F["Seal(nonce, plaintext, key)"]
        D--Nonce-->F
    end
    subgraph Armour
        direction LR
        G["base64Encode(encryptedPrivateKey)"]
        H["hexEncode(Salt)"]
        G --> armoured
        H --> armoured
    end
    C--Salt-->S
    S--Key-->SecretBox
    SecretBox--encryptedPrivateKey-->Armour
    C--Salt-->Armour
    kdf --> Armour
    hint --> Armour
    Armour--Marshal-->Return(encryptedArmouredPrivateKey)
```

The process above is reversed when unarmouring and decrypting a key in the keybase:

```mermaid
flowchart LR
    subgraph U[Unarmour]
        armoured
        B["hexDecode(salt)"]
        C["base64Decode(cipherText)"]
        D["verify"]
        armoured--salt-->B
        armoured--cipherText-->C
        armoured--kdf-->D

    end
    subgraph S[scrypt lib]
        E["key(salt, pass, ...)"]
    end
    subgraph SecretBox
        direction TB
        F["encryptedBytes[:nonceSize]"]
        H["Open(encryptedBytes[nonceSize:], nonce)"]
        F--Nonce-->H
    end
    encryptedArmouredPrivateKey --Unmarshal--> U
    B--Salt-->S
    C--encryptedBytes-->SecretBox
    S--Key-->SecretBox
    SecretBox-->PrivateKey
```

## SLIP-0010 HD Child Key Generation

[SLIP-0010](https://github.com/satoshilabs/slips/blob/master/slip-0010.md) key generation from a master key or seed is supported through the file [slip.go](./slip.go)

The keys are generated using the BIP-44 path `m/44'/635'/%d'` where `%d` is the index of the child key - this allows for the deterministic generation of up to `2147483647` hardened ed25519 child keys per master key.
Master key derivation is done as follows:

```mermaid
flowchart LR
    subgraph HMAC
        direction TB
        A["hmac = hmacNew(sha512, seedModifier)"]
        B["hmac.Write(seed)"]
        C["convertToBytes(hmac)"]
        A-->B
        B--hmac-->C
    end
    subgraph MASTER-KEY
        direction LR
        D["SecretKey: hmacBytes[:32]"]
        E["ChainCode: hmacBytes[32:]"]
        D --> KEY
        E --> KEY
    end
    seed-->HMAC
    HMAC--hmacBytes-->MASTER-KEY
```

Child keys are derived from their parents as follows:

```mermaid
flowchart LR
    subgraph HCHILD["HMAC-CHILD"]
        direction TB
        C["append(0x0, parent.SecretKey, bigEndian(index))"]
        A["hmacNew(sha512, parent.Chaincode)"]
        B["hmac.Write(data)"]
        D["convertToBytes(hmac)"]
        A--hmac-->B
        C--data-->B
        B--hmac-->D
    end
    subgraph CKEY[CHILD-KEY]
        direction LR
        F["SecretKey: hmacBytes[:32]"]
        G["ChainCode: hmacBytes[32:]"]
        F --> KEY
        G --> KEY
    end
    Index-->HCHILD
    Parent-->HCHILD
    HCHILD--hmacBytes-->CKEY
```

<!-- GITHUB_WIKI: shared/crypto/readme -->
