# Pocket Crypto <!-- omit in toc -->

- [KeyPair Interface](#keypair-interface)
  - [KeyPair Code Structure](#keypair-code-structure)
- [Encryption and Armouring](#encryption-and-armouring)

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

The following flowchart shows this process:

```mermaid
flowchart LR
    subgraph C[core lib]
        A["rand([16]byte)"]
    end
    subgraph S[scrypt lib]
        B["key(salt, pass, ...)"]
    end
    subgraph AES-GCM
        direction TB
        D["Cipher(key)"]
        E["GCM(block)"]
        F["Seal(plaintext, nonce)"]
        D--Block-->E
        E--Nonce-->F
    end
    subgraph Armour
        direction LR
        G["base64Encode(encryptedPrivateKey)"]
        H["hexEncode(Salt)"]
        G --> armoured
        H --> armoured
    end
    C--Salt-->S
    S--Key-->AES-GCM
    AES-GCM--encryptedPrivateKey-->Armour
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
    subgraph AES-GCM
        direction TB
        F["Cipher(key)"]
        G["GCM(block)"]
        H["Open(encryptedBytes, nonce)"]
        F--Block-->G
        G--Nonce-->H
    end
    encryptedArmouredPrivateKey --Unmarshal--> U
    B--Salt-->S
    C--encryptedBytes-->AES-GCM
    S--Key-->AES-GCM
    AES-GCM-->PrivateKey
```
