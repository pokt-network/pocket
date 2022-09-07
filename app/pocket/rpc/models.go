package rpc

// RawTXRequest defines model for RawTXRequest.
type RawTXRequest struct {
	Address     string `json:"address"`
	RawHexBytes string `json:"raw_hex_bytes"`
}

// PostV1ClientBroadcastTxSyncJSONBody defines parameters for PostV1ClientBroadcastTxSync.
type PostV1ClientBroadcastTxSyncJSONBody = RawTXRequest

// PostV1ClientBroadcastTxSyncJSONRequestBody defines body for PostV1ClientBroadcastTxSync for application/json ContentType.
type PostV1ClientBroadcastTxSyncJSONRequestBody = PostV1ClientBroadcastTxSyncJSONBody
