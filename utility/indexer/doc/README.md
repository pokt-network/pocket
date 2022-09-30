## Transaction Indexer


`txIndexer` implementation uses a `KVStore` (interface) to index the transactions

The transaction is indexed in the following formats:
- HASHKEY:      "h/SHA3(TxResultProtoBytes)"  VAL: TxResultProtoBytes     store value by hash (the key here is equivalent to the VALs below)
- HEIGHTKEY:    "b/height/index"              VAL: HASHKEY                store hashKey by height
- SENDERKEY:    "s/senderAddr"                VAL: HASHKEY                store hashKey by sender
- RECIPIENTKEY: "r/recipientAddr"             VAL: HASHKEY                store hashKey by recipient (if not empty)

FOOTNOTE: the height/index store is using [ELEN](https://github.com/jordanorelli/lexnum/blob/master/elen.pdf)
This is to ensure the results are stored sorted (assuming the `KVStore`` uses a byte-wise lexicographical sorting)