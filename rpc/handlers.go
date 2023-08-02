package rpc

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pokt-network/pocket/app"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
)

// CONSIDER: Remove all the V1 prefixes from the RPC module

func (s *rpcServer) GetV1Health(ctx echo.Context) error {
	return ctx.NoContent(http.StatusOK)
}

func (s *rpcServer) GetV1Version(ctx echo.Context) error {
	return ctx.String(http.StatusOK, app.AppVersion)
}

func (s *rpcServer) PostV1ClientBroadcastTxSync(ctx echo.Context) error {
	txParams := new(RawTXRequest)
	if err := ctx.Bind(txParams); err != nil {
		return ctx.String(http.StatusBadRequest, "bad request")
	}

	txBz, err := hex.DecodeString(txParams.RawHexBytes)
	if err != nil {
		return ctx.String(http.StatusBadRequest, "cannot decode tx bytes")
	}

	// Validate the transaction and add it to the mempool
	if err := s.GetBus().GetUtilityModule().HandleTransaction(txBz); err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	// Broadcast the transaction to the rest of the network if it passed the basic validation above
	if err := s.broadcastMessage(txBz); err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	return nil
}

func (s *rpcServer) PostV1ClientGetSession(ctx echo.Context) error {
	var body SessionRequest
	if err := ctx.Bind(&body); err != nil {
		return ctx.String(http.StatusBadRequest, "bad request")
	}

	session, err := s.GetBus().GetUtilityModule().GetSession(body.AppAddress, body.SessionHeight, body.Chain, body.Geozone)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	application := session.GetApplication()
	rpcApp := protocolActorToRPCProtocolActor(application)

	rpcServicers := make([]ProtocolActor, 0)
	for _, servicer := range session.GetServicers() {
		actor := protocolActorToRPCProtocolActor(servicer)
		rpcServicers = append(rpcServicers, actor)
	}

	rpcFishermen := make([]ProtocolActor, 0)
	for _, fisher := range session.GetFishermen() {
		actor := protocolActorToRPCProtocolActor(fisher)
		rpcFishermen = append(rpcFishermen, actor)
	}

	return ctx.JSON(http.StatusOK, Session{
		SessionId:        session.GetId(),
		SessionNumber:    session.GetSessionNumber(),
		SessionHeight:    session.GetSessionHeight(),
		NumSessionBlocks: session.GetNumSessionBlocks(),
		Chain:            string(session.GetRelayChain()),
		Geozone:          string(session.GetGeoZone()),
		Application:      rpcApp,
		Servicers:        rpcServicers,
		Fishermen:        rpcFishermen,
	})
}

func (s *rpcServer) PostV1ClientRelay(ctx echo.Context) error {
	servicer, err := s.GetBus().GetUtilityModule().GetServicerModule()
	if err != nil {
		return ctx.String(http.StatusInternalServerError, "node is not a servicer")
	}

	var body RelayRequest
	if err := ctx.Bind(&body); err != nil {
		return ctx.String(http.StatusBadRequest, "bad request")
	}

	// Parse body into the protobuf messages
	chain := &coreTypes.Identifiable{
		Id:   body.Meta.Chain.Id,
		Name: body.Meta.Chain.Name,
	}
	geozone := &coreTypes.Identifiable{
		Id:   body.Meta.Geozone.Id,
		Name: body.Meta.Geozone.Name,
	}
	relayMeta := &coreTypes.RelayMeta{
		BlockHeight:       body.Meta.BlockHeight,
		ServicerPublicKey: body.Meta.ServicerPubKey,
		RelayChain:        chain,
		GeoZone:           geozone,
		Signature:         body.Meta.Signature,
	}

	var relayRequest *coreTypes.Relay
	switch p := body.Payload.(type) {
	case JSONRPCPayload:
		relayRequest = buildJsonRPCRelayPayload(&p)
	case RESTPayload:
		relayRequest = buildRestRelayPayload(&p)
	default:
		return ctx.String(http.StatusBadRequest, "unsupported relay type")
	}
	relayRequest.Meta = relayMeta

	relayResponse, err := servicer.HandleRelay(relayRequest)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, RelayResponse{
		Payload:           relayResponse.Payload,
		ServicerSignature: relayResponse.ServicerSignature,
	})
}

// TECHDEBT: This will need to be changed when the HandleChallenge function is actually implemented
// because it copies data structures from v0
func (s *rpcServer) PostV1ClientChallenge(ctx echo.Context) error {
	var body ChallengeRequest
	if err := ctx.Bind(&body); err != nil {
		return ctx.String(http.StatusBadRequest, "bad request")
	}

	// Parse body into the protobuf messages
	majorityResponses := make([]*coreTypes.RelayResponse, 0)
	for _, resp := range body.MajorityResponses {
		relayResponse := &coreTypes.RelayResponse{
			Payload:           resp.Payload,
			ServicerSignature: resp.ServicerSignature,
		}
		majorityResponses = append(majorityResponses, relayResponse)
	}

	minorityResponse := &coreTypes.RelayResponse{
		Payload:           body.MinorityResponse.Payload,
		ServicerSignature: body.MinorityResponse.ServicerSignature,
	}

	challenge := &coreTypes.Challenge{
		SessionId:         body.SessionId,
		Address:           body.Address,
		ServicerPublicKey: body.ServicerPubKey,
		MinorityResponse:  minorityResponse,
		MajorityResponses: majorityResponses,
	}

	challengeResponse, err := s.GetBus().GetUtilityModule().HandleChallenge(challenge)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, "bad request")
	}

	return ctx.JSON(http.StatusOK, ChallengeResponse{
		Response: challengeResponse.Response,
	})
}

func (s *rpcServer) GetV1ConsensusState(ctx echo.Context) error {
	consensus := s.GetBus().GetConsensusModule()
	return ctx.JSON(200, ConsensusState{
		Height: int64(consensus.CurrentHeight()),
		Round:  int64(consensus.CurrentRound()),
		Step:   int64(consensus.CurrentStep()),
	})
}

func (s *rpcServer) GetV1P2pStakedActorsAddressBook(ctx echo.Context, params GetV1P2pStakedActorsAddressBookParams) error {
	var height int64
	var actors []Actor

	if params.Height != nil {
		height = *params.Height
	} else {
		height = int64(s.GetBus().GetConsensusModule().CurrentHeight())
	}

	readCtx, err := s.GetBus().GetPersistenceModule().NewReadContext(height)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	defer readCtx.Release()

	protocolActorGetter := getProtocolActorGetter(readCtx, params)

	protocolActors, err := protocolActorGetter(height)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	for _, protocolActor := range protocolActors {
		actors = append(actors, Actor{
			Address:    protocolActor.Address,
			Type:       protocolActorToRPCActorTypeEnum(protocolActor.ActorType),
			PublicKey:  protocolActor.PublicKey,
			ServiceUrl: protocolActor.ServiceUrl,
		})
	}

	response := P2PStakedActorsResponse{
		Actors: actors,
		Height: height,
	}

	return ctx.JSON(http.StatusOK, response)
}

// TECHDEBT: handle other relay payload types, e.g. JSON, GRPC, etc.
func buildJsonRPCRelayPayload(src *JSONRPCPayload) *coreTypes.Relay {
	payload := &coreTypes.Relay_JsonRpcPayload{
		JsonRpcPayload: &coreTypes.JSONRPCPayload{
			JsonRpc: src.Jsonrpc,
			Method:  src.Method,
		},
	}

	if src.Id != nil {
		payload.JsonRpcPayload.Id = src.Id.Id
	}

	if src.Parameters != nil {
		payload.JsonRpcPayload.Parameters = *src.Parameters
	}

	if src.Headers != nil {
		headers := make(map[string]string)
		for _, header := range *src.Headers {
			headers[header.Name] = header.Value
		}
		payload.JsonRpcPayload.Headers = headers
	}

	return &coreTypes.Relay{
		RelayPayload: payload,
	}
}

// DISCUSS: Path and Method requirements of relays in REST format.
func buildRestRelayPayload(src *RESTPayload) *coreTypes.Relay {
	return &coreTypes.Relay{
		RelayPayload: &coreTypes.Relay_RestPayload{
			RestPayload: &coreTypes.RESTPayload{
				Contents: *src,
			},
		},
	}
}

// UnmarshalJSON is the custom unmarshaller for RelayRequest type. This is needed because the payload could be JSONRPC or REST.
func (r *RelayRequest) UnmarshalJSON(data []byte) error {
	type relayWithJsonRpcPayload struct {
		Meta    RelayRequestMeta `json:"meta"`
		Payload JSONRPCPayload   `json:"payload"`
	}
	var jsonRpcRelay relayWithJsonRpcPayload
	if err := json.Unmarshal(data, &jsonRpcRelay); err == nil && jsonRpcRelay.Payload.Validate() == nil {
		r.Meta = jsonRpcRelay.Meta
		r.Payload = jsonRpcRelay.Payload
		return nil
	}

	type relayWithRestPayload struct {
		Meta    RelayRequestMeta `json:"meta"`
		Payload RESTPayload      `json:"payload"`
	}
	var restRelay relayWithRestPayload
	if err := json.Unmarshal(data, &restRelay); err == nil {
		r.Meta = restRelay.Meta
		r.Payload = restRelay.Payload
		return nil
	}

	return fmt.Errorf("invalid relay: %s", string(data))
}

// Validate returns an error if the payload struct is not valid JSONRPC
func (p *JSONRPCPayload) Validate() error {
	if p.Method == "" {
		return fmt.Errorf("%w: missing method field", errInvalidJsonRpc)
	}

	if p.Jsonrpc != "2.0" {
		return fmt.Errorf("%w: invalid JSONRPC field value: %q", errInvalidJsonRpc, p.Jsonrpc)
	}

	return nil
}

// UnmarshalJSON is the custom unmarshaller for JsonRpcId type. It is needed because JSONRPC spec allows the "id" field to be nil, an integer, or a string.
//
//	See the following link for more details:
//	https://www.jsonrpc.org/specification#request_object
func (i *JsonRpcId) UnmarshalJSON(data []byte) error {
	var v int64
	if err := json.Unmarshal(data, &v); err == nil {
		i.Id = []byte(fmt.Sprintf("%d", v))
		return nil
	}

	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		i.Id = []byte(s)
		return nil
	}

	return fmt.Errorf("invalid JSONRPC ID value: %v", data)
}

// MarshalJSON is the custom marshaller for JsonRpcId type. It is needed to flatten the struct as a single value, i.e. mask the "Id" field.
func (i *JsonRpcId) MarshalJSON() ([]byte, error) {
	return i.Id, nil
}
