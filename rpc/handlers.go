package rpc

import (
	"encoding/hex"
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

	if err := s.GetBus().GetUtilityModule().HandleTransaction(txBz); err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

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

// TECHDEBT: This will need to be changed when the HandleRelay function is actually implemented
// because it copies data structures from v0. For example, AATs are no longer necessary in v1.
func (s *rpcServer) PostV1ClientRelay(ctx echo.Context) error {
	var utility = s.GetBus().GetUtilityModule()

	if utility.GetServicerModule() == nil {
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
	aat := &coreTypes.AAT{
		Version:              body.Meta.Token.Version,
		ApplicationPublicKey: body.Meta.Token.AppPubKey,
		ClientPublicKey:      body.Meta.Token.ClientPubKey,
		ApplicationSignature: body.Meta.Token.AppSignature,
	}
	relayMeta := &coreTypes.RelayMeta{
		BlockHeight:       body.Meta.BlockHeight,
		ServicerPublicKey: body.Meta.ServicerPubKey,
		RelayChain:        chain,
		GeoZone:           geozone,
		Token:             aat,
		Signature:         body.Meta.Signature,
	}

	payload := &coreTypes.RelayPayload{
		Data:     body.Payload.Data,
		Method:   body.Payload.Method,
		HttpPath: body.Payload.Path,
	}

	headers := make(map[string]string)
	for _, header := range body.Payload.Headers {
		headers[header.Name] = header.Value
	}
	payload.Headers = headers

	relayRequest := &coreTypes.Relay{
		Payload: payload,
		Meta:    relayMeta,
	}

	relayResponse, err := utility.HandleRelay(relayRequest)
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
