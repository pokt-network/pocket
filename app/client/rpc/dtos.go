package rpc

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type SendRawTxParams struct {
	Addr        string `json:"address"`
	RawHexBytes string `json:"raw_hex_bytes"`
}
