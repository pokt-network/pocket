# Keybase Vault

The Keybase Vault stores key pairs, `(public_key, armoured_private_key)`, using the [Hashicorp Vault](https://www.vaultproject.io/) KV secrets engine. The Keybase Vault is a wrapper around the Hashicorp Vault API.

Keybase Vault requires a few pieces of information to be able to connect to it:

- Address (e.g. `VAULT_ADDR`=http://127.0.0.1:8200/)
- Token (e.g. `VAULT_TOKEN`=hvs.25YM7qJDN8S2EpFEA4SL0ciD)
- Mount Path (e.g. /secret)

## Simple Vault Demo

Here are some quick commands to familiarize oneself with the keybase vault. Deep dive here: https://developer.hashicorp.com/vault/tutorials/getting-started/getting-started-intro

Or for the impatient and unsafe way, run the following commands:

```sh
brew tap hashicorp/tap
brew install hashicorp/tap/vault
# Start a vault server with kv secrets engine enabled at path "secret/"
vault server -dev -dev-root-token-id="dev-only-token"
```

Vault comes with a web UI. Open a browser and navigate to http://127.0.0.1:8200 and login with the token `dev-only-token`. Take a look at the secrets engine at `secret/`.

Then, open a new terminal window and run the following commands:

## Using environment variables

```sh
export VAULT_ADDR='http://127.0.0.1:8200'
export VAULT_TOKEN='dev-only-token'

p1 Keys Create --keybase vault

# note the address
# {"level":"info","address":"addr_goes_here","time":"2023-02-24T23:12:06-04:00","message":"New Key Created"}

p1 Keys Get --keybase vault addr_goes_here

# {"level":"info","address":"addr_goes_here","public_key":"public_key_goes_here","time":"2023-02-24T23:14:01-04:00","message":"Found key"}

p1 Keys Export addr_goes_here --keybase vault

## note the private key
# {"level":"info","private_key":"{\"kdf\":\"scrypt\",\"salt\":\"salty_goes_ere\",\"secparam\":\"12\",\"hint\":\"\",\"ciphertext\":\"ciphertext_goes_here\"}","time":"2023-02-24T23:12:53-04:00","message":"Key exported"}

p1 Keys Import --keybase vault --import_format json "{\"kdf\":\"scrypt\",\"salt\":\"salty_goes_ere\",\"secparam\":\"12\",\"hint\":\"\",\"ciphertext\":\"ciphertext_goes_here\"}"

p1 Keys List --keybase vault

# {"level":"info","addresses":["addr_goes_here"],"time":"2023-02-24T23:14:44-04:00","message":"Get all keys"}

p1 Keys Sign --keybase vault addr_goes_here abcd

# note the signature
# {"level":"info","signature":"signature_goes_here","address":"addr_goes_here","time":"2023-02-24T23:15:18-04:00","message":"Message signed"}

p1 Keys Verify --keybase vault addr_goes_here abcd signature_goes_here

# {"level":"info","address":"addr_goes_here","valid":true,"time":"2023-02-24T23:16:05-04:00","message":"Signature checked"}

p1 Keys Update --keybase vault addr_goes_here

p1 Keys DeriveChild addr_goes_here 0 --keybase vault
# {"level":"info","address":"new_addr_goes_here","parent":"addr_goes_here","index":0,"stored":true,"time":"2023-02-28T09:26:11-04:00","message":"Child key derived"}

p1 Keys Delete --keybase vault addr_goes_here

## TODO: add SignTx and VerifyTx

p1 Keys SignTx --keybase vault
p1 Keys VerifyTx --keybase vault
```

## Example using command line flags

```sh

export VAULT_ADDR='http://127.0.0.1:8200'
export VAULT_TOKEN='dev-only-token'

p1 Keys Create --keybase vault \
    --vault-token dev-only-token \
    --vault-addr http://127.0.0.1:8200/

# note the address
# {"level":"info","address":"addr_goes_here","time":"2023-02-24T23:12:06-04:00","message":"New Key Created"}

p1 Keys Get --keybase vault \
    --vault-token dev-only-token \
    --vault-addr http://127.0.0.1:8200/  addr_goes_here

# {"level":"info","address":"addr_goes_here","public_key":"d82cd23f4809491c04ab456dd9714e647093bcc6cb649a8510f4d54c194f80ea","time":"2023-02-24T23:14:01-04:00","message":"Found key"}

p1 Keys Export addr_goes_here --keybase vault \
    --vault-token dev-only-token \
    --vault-addr http://127.0.0.1:8200/

## note the private key
# {"level":"info","private_key":"{\"kdf\":\"scrypt\",\"salt\":\"salty_goes_ere\",\"secparam\":\"12\",\"hint\":\"\",\"ciphertext\":\"ciphertext_goes_here\"}","time":"2023-02-24T23:12:53-04:00","message":"Key exported"}

p1 Keys Import --keybase vault \
    --vault-token dev-only-token \
    --vault-addr http://127.0.0.1:8200/ \
    --import_format json "{\"kdf\":\"scrypt\",\"salt\":\"salty_goes_ere\",\"secparam\":\"12\",\"hint\":\"\",\"ciphertext\":\"ciphertext_goes_here\"}"

p1 Keys List --keybase vault \
    --vault-token dev-only-token \
    --vault-addr http://127.0.0.1:8200/

# {"level":"info","addresses":["addr_goes_here"],"time":"2023-02-24T23:14:44-04:00","message":"Get all keys"}

p1 Keys Sign --keybase vault \
    --vault-token dev-only-token \
    --vault-addr http://127.0.0.1:8200/ \
    addr_goes_here abcd

# note the signature
# {"level":"info","signature":"signature_goes_here","address":"addr_goes_here","time":"2023-02-24T23:15:18-04:00","message":"Message signed"}

p1 Keys Verify --keybase vault \
    --vault-token dev-only-token \
    --vault-addr http://127.0.0.1:8200/ \
    addr_goes_here abcd signature_goes_here

# {"level":"info","address":"addr_goes_here","valid":true,"time":"2023-02-24T23:16:05-04:00","message":"Signature checked"}

p1 Keys Update --keybase vault \
    --vault-token dev-only-token \
    --vault-addr http://127.0.0.1:8200/ \
    addr_goes_here


p1 Keys Keys DeriveChild addr_goes_here 0 --keybase vault \
    --vault-token dev-only-token \
    --vault-addr http://127.0.0.1:8200/ \
    addr_goes_here

# {"level":"info","address":"new_addr_goes_here","parent":"addr_goes_here","index":0,"stored":true,"time":"2023-02-28T09:26:11-04:00","message":"Child key derived"}

p1 Keys Delete --keybase vault \
    --vault-token dev-only-token \
    --vault-addr http://127.0.0.1:8200/ \
    addr_goes_here

## TODO: add SignTx and VerifyTx

p1 Keys SignTx --keybase vault \
    --vault-token dev-only-token \
    --vault-addr http://127.0.0.1:8200/
p1 Keys VerifyTx --keybase vault \
    --vault-token dev-only-token \
    --vault-addr http://127.0.0.1:8200/

```
