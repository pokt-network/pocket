# Transaction Indexer

`txIndexer` implementation uses a `KVStore` (interface) to index the transactions.

## Index Types

| Key          | Index                        | Value              | Description                                                        |
| ------------ | ---------------------------- | ------------------ | ------------------------------------------------------------------ |
| HASHKEY      | `h/SHA3(TxResultProtoBytes)` | TxResultProtoBytes | store value by hash (the key here is equivalent to the VALs below) |
| HEIGHTKEY    | `b/height/index`             | HASHKEY            | store hashKey by height                                            |
| SENDERKEY    | `s/senderAddr`               | HASHKEY            | store hashKey by sender                                            |
| RECIPIENTKEY | `r/recipientAddr`            | HASHKEY            | store hashKey by recipient (if not empty)                          

## ELEN Index

The height/index store uses [ELEN](https://github.com/jordanorelli/lexnum/blob/master/elen.pdf). This is to ensure the results are stored sorted (assuming the `KVStore` uses a byte-wise lexicographical sorting).