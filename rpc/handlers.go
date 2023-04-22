package rpc

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/pokt-network/pocket/app"
	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/utils"
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

	if err = s.GetBus().GetUtilityModule().HandleTransaction(txBz); err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	if err := s.broadcastMessage(txBz); err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	return nil
}

// DISCUSSION: This may need to be changed when the GetSession function is actually implemented
func (s *rpcServer) PostV1ClientDispatch(ctx echo.Context) error {
	var body DispatchRequest
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
		SessionId:   session.GetId(),
		Height:      session.GetHeight(),
		Chain:       string(session.GetRelayChain()),
		Geozone:     string(session.GetGeoZone()),
		Application: rpcApp,
		Servicers:   rpcServicers,
		Fishermen:   rpcFishermen,
	})
}

// DISCUSSION: This may need to be changed when the SendRelay function is actually implemented
func (s *rpcServer) PostV1ClientRelay(ctx echo.Context) error {
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

	relayResponse, err := s.GetBus().GetUtilityModule().SendRelay(relayRequest)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, RelayResponse{
		Payload:           relayResponse.Payload,
		ServicerSignature: relayResponse.ServicerSignature,
	})
}

// DISCUSSION: This may need to be changed when the HandleChallenge function is actually implemented
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

// Queries

func (s *rpcServer) PostV1QueryAccount(ctx echo.Context) error {
	var body QueryAddressHeight
	if err := ctx.Bind(&body); err != nil {
		return ctx.String(http.StatusBadRequest, "bad request")
	}

	// Get latest stored block height
	height := body.Height
	if height == 0 {
		currentHeight := int64(s.GetBus().GetConsensusModule().CurrentHeight())
		if currentHeight > 0 {
			currentHeight -= 1
		}
		height = currentHeight
	}
	readCtx, err := s.GetBus().GetPersistenceModule().NewReadContext(height)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	defer readCtx.Release() //nolint:errcheck // We only need to make sure the readCtx is released

	accBz, err := hex.DecodeString(body.Address)
	if err != nil {
		return ctx.String(http.StatusBadRequest, err.Error())
	}
	amount, err := readCtx.GetAccountAmount(accBz, height)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, Account{
		Address: body.Address,
		Coins:   []Coin{{Amount: amount, Denom: "upokt"}},
	})
}

func (s *rpcServer) PostV1QueryAccounts(ctx echo.Context) error {
	var body QueryHeightPaginated
	if err := ctx.Bind(&body); err != nil {
		return ctx.String(http.StatusBadRequest, "bad request")
	}

	// Get latest stored block height
	height := body.Height
	if height == 0 {
		currentHeight := int64(s.GetBus().GetConsensusModule().CurrentHeight())
		if currentHeight > 0 {
			currentHeight -= 1
		}
		height = currentHeight
	}
	readCtx, err := s.GetBus().GetPersistenceModule().NewReadContext(height)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	defer readCtx.Release() //nolint:errcheck // We only need to make sure the readCtx is released

	allAccounts, err := readCtx.GetAllAccounts(height)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	start, end, totalPages, err := getPageIndexes(len(allAccounts), int(body.Page), int(body.PerPage))
	if err != nil && !errors.Is(err, errNoItems) {
		return ctx.String(http.StatusBadRequest, err.Error())
	}
	if totalPages == 0 || errors.Is(err, errNoItems) {
		return ctx.JSON(http.StatusOK, QueryAccountsResponse{})
	}

	accounts := make([]Account, 0)
	for _, account := range allAccounts[start : end+1] {
		accounts = append(accounts, Account{
			Address: account.Address,
			Coins:   []Coin{{Amount: account.Amount, Denom: "upokt"}},
		})
	}

	return ctx.JSON(http.StatusOK, QueryAccountsResponse{
		Result:     accounts,
		Page:       body.Page,
		TotalPages: int64(totalPages),
	})
}

func (s *rpcServer) PostV1QueryAccountTxs(ctx echo.Context) error {
	var body QueryAddressPaginated
	if err := ctx.Bind(&body); err != nil {
		return ctx.String(http.StatusBadRequest, "bad request")
	}
	sortDesc := checkSortDesc(*body.Sort)

	txIndexer := s.GetBus().GetPersistenceModule().GetTxIndexer()
	txResults, err := txIndexer.GetBySender(body.Address, sortDesc)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	start, end, totalPages, err := getPageIndexes(len(txResults), int(body.Page), int(body.PerPage))
	if err != nil && !errors.Is(err, errNoItems) {
		return ctx.String(http.StatusBadRequest, err.Error())
	}
	if totalPages == 0 || errors.Is(err, errNoItems) {
		return ctx.JSON(http.StatusOK, QueryAccountTxsResponse{})
	}

	pageTxs := make([]Transaction, 0)
	for _, txResult := range txResults[start : end+1] {
		rpcTx, err := s.txResultToRPCTransaction(txResult)
		if err != nil {
			return ctx.String(http.StatusInternalServerError, err.Error())
		}
		pageTxs = append(pageTxs, *rpcTx)
	}

	return ctx.JSON(http.StatusOK, QueryAccountTxsResponse{
		Txs:        pageTxs,
		Page:       body.Page,
		TotalTxs:   int64(len(txResults)),
		TotalPages: int64(totalPages),
	})
}

func (s *rpcServer) GetV1QueryAllChainParams(ctx echo.Context) error {
	// Get latest stored block height
	currentHeight := int64(s.GetBus().GetConsensusModule().CurrentHeight())
	if currentHeight > 0 {
		currentHeight -= 1
	}
	readCtx, err := s.GetBus().GetPersistenceModule().NewReadContext(currentHeight)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	defer readCtx.Release() //nolint:errcheck // We only need to make sure the readCtx is released

	paramSlice, err := readCtx.GetAllParams()
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	resp := make([]Parameter, 0)
	for i := 0; i < len(paramSlice); i++ {
		resp = append(resp, Parameter{
			ParameterName:  paramSlice[i][0],
			ParameterValue: paramSlice[i][1],
		})
	}
	return ctx.JSON(200, resp)
}

func (s *rpcServer) PostV1QueryApp(ctx echo.Context) error {
	var body QueryAddressHeight
	if err := ctx.Bind(&body); err != nil {
		return ctx.String(http.StatusBadRequest, "bad request")
	}

	// Get latest stored block height
	height := body.Height
	if height == 0 {
		currentHeight := int64(s.GetBus().GetConsensusModule().CurrentHeight())
		if currentHeight > 0 {
			currentHeight -= 1
		}
		height = currentHeight
	}
	readCtx, err := s.GetBus().GetPersistenceModule().NewReadContext(height)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	defer readCtx.Release() //nolint:errcheck // We only need to make sure the readCtx is released

	addrBz, err := hex.DecodeString(body.Address)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	application, err := readCtx.GetApp(addrBz, height)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	actor := protocolActorToRPCProtocolActor(application)
	return ctx.JSON(http.StatusOK, actor)
}

func (s *rpcServer) PostV1QueryApps(ctx echo.Context) error {
	var body QueryHeightPaginated
	if err := ctx.Bind(&body); err != nil {
		return ctx.String(http.StatusBadRequest, "bad request")
	}

	// Get latest stored block height
	height := body.Height
	if height == 0 {
		currentHeight := int64(s.GetBus().GetConsensusModule().CurrentHeight())
		if currentHeight > 0 {
			currentHeight -= 1
		}
		height = currentHeight
	}
	readCtx, err := s.GetBus().GetPersistenceModule().NewReadContext(height)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	defer readCtx.Release() //nolint:errcheck // We only need to make sure the readCtx is released

	allApps, err := readCtx.GetAllApps(height)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	start, end, totalPages, err := getPageIndexes(len(allApps), int(body.Page), int(body.PerPage))
	if err != nil && !errors.Is(err, errNoItems) {
		return ctx.String(http.StatusBadRequest, err.Error())
	}
	if totalPages == 0 || errors.Is(err, errNoItems) {
		return ctx.JSON(http.StatusOK, QueryAppsResponse{})
	}

	rpcApps := make([]ProtocolActor, 0)
	for _, app := range allApps[start : end+1] {
		actor := protocolActorToRPCProtocolActor(app)
		rpcApps = append(rpcApps, actor)
	}

	return ctx.JSON(http.StatusOK, QueryAppsResponse{
		Apps:       rpcApps,
		TotalApps:  int64(len(allApps)),
		Page:       body.Page,
		TotalPages: int64(totalPages),
	})
}

func (s *rpcServer) PostV1QueryBalance(ctx echo.Context) error {
	var body QueryAddressHeight
	if err := ctx.Bind(&body); err != nil {
		return ctx.String(http.StatusBadRequest, "bad request")
	}

	// Get latest stored block height
	height := body.Height
	if height == 0 {
		currentHeight := int64(s.GetBus().GetConsensusModule().CurrentHeight())
		if currentHeight > 0 {
			currentHeight -= 1
		}
		height = currentHeight
	}
	readCtx, err := s.GetBus().GetPersistenceModule().NewReadContext(height)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	defer readCtx.Release() //nolint:errcheck // We only need to make sure the readCtx is released

	accBz, err := hex.DecodeString(body.Address)
	if err != nil {
		return ctx.String(http.StatusBadRequest, err.Error())
	}
	amountStr, err := readCtx.GetAccountAmount(accBz, height)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	amount, err := strconv.ParseInt(amountStr, 10, 64)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, QueryBalanceResponse{
		Balance: amount,
	})
}

func (s *rpcServer) PostV1QueryBlock(ctx echo.Context) error {
	var body QueryHeight
	if err := ctx.Bind(&body); err != nil {
		return ctx.String(http.StatusBadRequest, "bad request")
	}

	// Get latest stored block height
	height := uint64(body.Height)
	if height == 0 {
		currentHeight := s.GetBus().GetConsensusModule().CurrentHeight()
		if currentHeight > 0 {
			currentHeight -= 1
		}
		height = currentHeight
	}

	blockStore := s.GetBus().GetPersistenceModule().GetBlockStore()
	blockBz, err := blockStore.Get(utils.HeightToBytes(height))
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	block := new(coreTypes.Block)
	if err := codec.GetCodec().Unmarshal(blockBz, block); err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	rpcBlock, err := s.blockToRPCBlock(block)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, rpcBlock)
}

func (s *rpcServer) PostV1QueryBlockTxs(ctx echo.Context) error {
	var body QueryHeightPaginated
	if err := ctx.Bind(&body); err != nil {
		return ctx.String(http.StatusBadRequest, "bad request")
	}
	sortDesc := checkSortDesc(*body.Sort)

	// Get latest stored block height
	height := uint64(body.Height)
	if height == 0 {
		currentHeight := s.GetBus().GetConsensusModule().CurrentHeight()
		if currentHeight > 0 {
			currentHeight -= 1
		}
		height = currentHeight
	}

	blockStore := s.GetBus().GetPersistenceModule().GetBlockStore()
	blockBz, err := blockStore.Get(utils.HeightToBytes(height))
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	block := new(coreTypes.Block)
	if err := codec.GetCodec().Unmarshal(blockBz, block); err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	rpcBlock, err := s.blockToRPCBlock(block)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	allTxs := rpcBlock.Transactions
	if sortDesc {
		for i, j := 0, len(allTxs)-1; i < j; i, j = i+1, j-1 {
			allTxs[i], allTxs[j] = allTxs[j], allTxs[i]
		}
	}

	start, end, totalPages, err := getPageIndexes(len(allTxs), int(body.Page), int(body.PerPage))
	if err != nil && !errors.Is(err, errNoItems) {
		return ctx.String(http.StatusBadRequest, err.Error())
	}
	if totalPages == 0 || errors.Is(err, errNoItems) {
		return ctx.JSON(http.StatusOK, QueryTxsResponse{})
	}

	return ctx.JSON(http.StatusOK, QueryTxsResponse{
		Transactions: allTxs[start : end+1],
		TotalTxs:     int64(len(allTxs)),
		Page:         body.Page,
		TotalPages:   int64(totalPages),
	})
}

func (s *rpcServer) PostV1QueryFisherman(ctx echo.Context) error {
	var body QueryAddressHeight
	if err := ctx.Bind(&body); err != nil {
		return ctx.String(http.StatusBadRequest, "bad request")
	}

	// Get latest stored block height
	height := body.Height
	if height == 0 {
		currentHeight := int64(s.GetBus().GetConsensusModule().CurrentHeight())
		if currentHeight > 0 {
			currentHeight -= 1
		}
		height = currentHeight
	}
	readCtx, err := s.GetBus().GetPersistenceModule().NewReadContext(height)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	defer readCtx.Release() //nolint:errcheck // We only need to make sure the readCtx is released

	addrBz, err := hex.DecodeString(body.Address)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	fisherman, err := readCtx.GetFisherman(addrBz, height)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	actor := protocolActorToRPCProtocolActor(fisherman)
	return ctx.JSON(http.StatusOK, actor)
}

func (s *rpcServer) PostV1QueryFishermen(ctx echo.Context) error {
	var body QueryHeightPaginated
	if err := ctx.Bind(&body); err != nil {
		return ctx.String(http.StatusBadRequest, "bad request")
	}

	// Get latest stored block height
	height := body.Height
	if height == 0 {
		currentHeight := int64(s.GetBus().GetConsensusModule().CurrentHeight())
		if currentHeight > 0 {
			currentHeight -= 1
		}
		height = currentHeight
	}
	readCtx, err := s.GetBus().GetPersistenceModule().NewReadContext(height)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	defer readCtx.Release() //nolint:errcheck // We only need to make sure the readCtx is released

	allFishermen, err := readCtx.GetAllFishermen(height)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	start, end, totalPages, err := getPageIndexes(len(allFishermen), int(body.Page), int(body.PerPage))
	if err != nil && !errors.Is(err, errNoItems) {
		return ctx.String(http.StatusBadRequest, err.Error())
	}
	if totalPages == 0 || errors.Is(err, errNoItems) {
		return ctx.JSON(http.StatusOK, QueryFishermenResponse{})
	}

	rpcFishermen := make([]ProtocolActor, 0)
	for _, fisher := range allFishermen[start : end+1] {
		actor := protocolActorToRPCProtocolActor(fisher)
		rpcFishermen = append(rpcFishermen, actor)
	}

	return ctx.JSON(http.StatusOK, QueryFishermenResponse{
		Fishermen:      rpcFishermen,
		TotalFishermen: int64(len(allFishermen)),
		Page:           body.Page,
		TotalPages:     int64(totalPages),
	})
}

func (s *rpcServer) GetV1QueryHeight(ctx echo.Context) error {
	// Get latest stored block height
	currentHeight := s.GetBus().GetConsensusModule().CurrentHeight()
	if currentHeight > 0 {
		currentHeight -= 1
	}

	return ctx.JSON(http.StatusOK, QueryHeight{
		Height: int64(currentHeight),
	})
}

func (s *rpcServer) PostV1QueryParam(ctx echo.Context) error {
	var body QueryParameter
	if err := ctx.Bind(&body); err != nil {
		return ctx.String(http.StatusBadRequest, "bad request")
	}

	// Get latest stored block height
	currentHeight := int64(s.GetBus().GetConsensusModule().CurrentHeight())
	if currentHeight > 0 {
		currentHeight -= 1
	}
	height := body.Height
	if height == 0 || height > currentHeight {
		height = currentHeight
	}

	readCtx, err := s.GetBus().GetPersistenceModule().NewReadContext(height)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	defer readCtx.Release() //nolint:errcheck // We only need to make sure the readCtx is released

	paramValue, err := readCtx.GetStringParam(body.ParamName, height)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, Parameter{
		ParameterName:  body.ParamName,
		ParameterValue: paramValue,
	})
}

func (s *rpcServer) PostV1QueryServicer(ctx echo.Context) error {
	var body QueryAddressHeight
	if err := ctx.Bind(&body); err != nil {
		return ctx.String(http.StatusBadRequest, "bad request")
	}

	// Get latest stored block height
	height := body.Height
	if height == 0 {
		currentHeight := int64(s.GetBus().GetConsensusModule().CurrentHeight())
		if currentHeight > 0 {
			currentHeight -= 1
		}
		height = currentHeight
	}
	readCtx, err := s.GetBus().GetPersistenceModule().NewReadContext(height)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	defer readCtx.Release() //nolint:errcheck // We only need to make sure the readCtx is released

	addrBz, err := hex.DecodeString(body.Address)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	servicer, err := readCtx.GetServicer(addrBz, height)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	actor := protocolActorToRPCProtocolActor(servicer)
	return ctx.JSON(http.StatusOK, actor)
}

func (s *rpcServer) PostV1QueryServicers(ctx echo.Context) error {
	var body QueryHeightPaginated
	if err := ctx.Bind(&body); err != nil {
		return ctx.String(http.StatusBadRequest, "bad request")
	}

	// Get latest stored block height
	height := body.Height
	if height == 0 {
		currentHeight := int64(s.GetBus().GetConsensusModule().CurrentHeight())
		if currentHeight > 0 {
			currentHeight -= 1
		}
		height = currentHeight
	}
	readCtx, err := s.GetBus().GetPersistenceModule().NewReadContext(height)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	defer readCtx.Release() //nolint:errcheck // We only need to make sure the readCtx is released

	allServicers, err := readCtx.GetAllServicers(height)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	start, end, totalPages, err := getPageIndexes(len(allServicers), int(body.Page), int(body.PerPage))
	if err != nil && !errors.Is(err, errNoItems) {
		return ctx.String(http.StatusBadRequest, err.Error())
	}
	if totalPages == 0 || errors.Is(err, errNoItems) {
		return ctx.JSON(http.StatusOK, QueryServicersResponse{})
	}

	rpcServicers := make([]ProtocolActor, 0)
	for _, servicer := range allServicers[start : end+1] {
		actor := protocolActorToRPCProtocolActor(servicer)
		rpcServicers = append(rpcServicers, actor)
	}

	return ctx.JSON(http.StatusOK, QueryServicersResponse{
		Servicers:      rpcServicers,
		TotalServicers: int64(len(allServicers)),
		Page:           body.Page,
		TotalPages:     int64(totalPages),
	})
}

func (s *rpcServer) PostV1QuerySupply(ctx echo.Context) error {
	var body QueryHeight
	if err := ctx.Bind(&body); err != nil {
		return ctx.String(http.StatusBadRequest, "bad request")
	}

	// Get latest stored block height
	height := body.Height
	if height == 0 {
		currentHeight := int64(s.GetBus().GetConsensusModule().CurrentHeight())
		if currentHeight > 0 {
			currentHeight -= 1
		}
		height = currentHeight
	}
	readCtx, err := s.GetBus().GetPersistenceModule().NewReadContext(height)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	defer readCtx.Release() //nolint:errcheck // We only need to make sure the readCtx is released

	pools, err := readCtx.GetAllPools(height)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	rpcPools := make([]Pool, 0)
	total := new(big.Int)
	for _, pool := range pools {
		name := coreTypes.PoolAddressToFriendlyName(pool.Address)
		amount, success := new(big.Int).SetString(pool.Amount, 10)
		if !success {
			return ctx.String(http.StatusInternalServerError, "failed to convert amount to big.Int")
		}
		total = total.Add(total, amount)
		rpcPools = append(rpcPools, Pool{
			Address: pool.Address,
			Name:    name,
			Amount:  pool.Amount,
			Denom:   "upokt",
		})
	}

	return ctx.JSON(http.StatusOK, QuerySupplyResponse{
		Pools: rpcPools,
		Total: Coin{
			Amount: total.String(),
			Denom:  "upokt",
		},
	})
}

func (s *rpcServer) PostV1QuerySupportedChains(ctx echo.Context) error {
	var body QueryHeight
	if err := ctx.Bind(&body); err != nil {
		return ctx.String(http.StatusBadRequest, "bad request")
	}

	// Get latest stored block height
	height := body.Height
	if height == 0 {
		currentHeight := int64(s.GetBus().GetConsensusModule().CurrentHeight())
		if currentHeight > 0 {
			currentHeight -= 1
		}
		height = currentHeight
	}
	readCtx, err := s.GetBus().GetPersistenceModule().NewReadContext(height)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	defer readCtx.Release() //nolint:errcheck // We only need to make sure the readCtx is released

	chains, err := readCtx.GetSupportedChains(height)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, QuerySupportedChainsResponse{
		SupportedChains: chains,
	})
}

func (s *rpcServer) PostV1QueryTx(ctx echo.Context) error {
	var body QueryHash
	if err := ctx.Bind(&body); err != nil {
		return ctx.String(http.StatusBadRequest, "bad request")
	}

	hashBz, err := hex.DecodeString(body.Hash)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	txIndexer := s.GetBus().GetPersistenceModule().GetTxIndexer()
	txResult, err := txIndexer.GetByHash(hashBz)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	rpcTx, err := s.txResultToRPCTransaction(txResult)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, rpcTx)
}

func (s *rpcServer) PostV1QueryUnconfirmedTx(ctx echo.Context) error {
	var body QueryHash
	if err := ctx.Bind(&body); err != nil {
		return ctx.String(http.StatusBadRequest, "bad request")
	}

	mempool := s.GetBus().GetUtilityModule().GetMempool()
	uncTx := mempool.Get(body.Hash)
	if uncTx == nil {
		return ctx.String(http.StatusBadRequest, fmt.Sprintf("hash not found in mempool: %s", body.Hash))
	}

	rpcUncTxs, err := s.txProtoBytesToRPCTransactions([][]byte{uncTx})
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, rpcUncTxs[0])
}

func (s *rpcServer) PostV1QueryUnconfirmedTxs(ctx echo.Context) error {
	var body QueryPaginated
	if err := ctx.Bind(&body); err != nil {
		return ctx.String(http.StatusBadRequest, "bad request")
	}

	mempool := s.GetBus().GetUtilityModule().GetMempool()
	uncTxs := mempool.GetAll()

	start, end, totalPages, err := getPageIndexes(len(uncTxs), int(body.Page), int(body.PerPage))
	if err != nil && !errors.Is(err, errNoItems) {
		return ctx.String(http.StatusBadRequest, err.Error())
	}
	if totalPages == 0 || errors.Is(err, errNoItems) {
		return ctx.JSON(http.StatusOK, QueryTxsResponse{})
	}

	rpcUncTxs, err := s.txProtoBytesToRPCTransactions(uncTxs[start : end+1])
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, QueryTxsResponse{
		Transactions: rpcUncTxs,
		TotalTxs:     int64(len(uncTxs)),
		Page:         body.Page,
		TotalPages:   int64(totalPages),
	})
}

func (s *rpcServer) PostV1QueryUpgrade(ctx echo.Context) error {
	var body QueryHeight
	if err := ctx.Bind(&body); err != nil {
		return ctx.String(http.StatusBadRequest, "bad request")
	}

	// Get latest stored block height
	height := body.Height
	if height == 0 {
		currentHeight := int64(s.GetBus().GetConsensusModule().CurrentHeight())
		if currentHeight > 0 {
			currentHeight -= 1
		}
		height = currentHeight
	}
	reatCtx, err := s.GetBus().GetPersistenceModule().NewReadContext(height)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	version, err := reatCtx.GetVersionAtHeight(height)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, QueryUpgradeResponse{
		Height:  height,
		Version: version,
	})
}

func (s *rpcServer) PostV1QueryValidator(ctx echo.Context) error {
	var body QueryAddressHeight
	if err := ctx.Bind(&body); err != nil {
		return ctx.String(http.StatusBadRequest, "bad request")
	}

	// Get latest stored block height
	height := body.Height
	if height == 0 {
		currentHeight := int64(s.GetBus().GetConsensusModule().CurrentHeight())
		if currentHeight > 0 {
			currentHeight -= 1
		}
		height = currentHeight
	}
	readCtx, err := s.GetBus().GetPersistenceModule().NewReadContext(height)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	defer readCtx.Release() //nolint:errcheck // We only need to make sure the readCtx is released

	addrBz, err := hex.DecodeString(body.Address)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	validator, err := readCtx.GetValidator(addrBz, height)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	actor := protocolActorToRPCProtocolActor(validator)
	return ctx.JSON(http.StatusOK, actor)
}

func (s *rpcServer) PostV1QueryValidators(ctx echo.Context) error {
	var body QueryHeightPaginated
	if err := ctx.Bind(&body); err != nil {
		return ctx.String(http.StatusBadRequest, "bad request")
	}

	// Get latest stored block height
	height := body.Height
	if height == 0 {
		currentHeight := int64(s.GetBus().GetConsensusModule().CurrentHeight())
		if currentHeight > 0 {
			currentHeight -= 1
		}
		height = currentHeight
	}
	readCtx, err := s.GetBus().GetPersistenceModule().NewReadContext(height)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	defer readCtx.Release() //nolint:errcheck // We only need to make sure the readCtx is released

	allValidators, err := readCtx.GetAllValidators(height)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	start, end, totalPages, err := getPageIndexes(len(allValidators), int(body.Page), int(body.PerPage))
	if err != nil && !errors.Is(err, errNoItems) {
		return ctx.String(http.StatusBadRequest, err.Error())
	}
	if totalPages == 0 || errors.Is(err, errNoItems) {
		return ctx.JSON(http.StatusOK, QueryValidatorsResponse{})
	}

	rpcValidators := make([]ProtocolActor, 0)
	for _, val := range allValidators[start : end+1] {
		actor := protocolActorToRPCProtocolActor(val)
		rpcValidators = append(rpcValidators, actor)
	}

	return ctx.JSON(http.StatusOK, QueryValidatorsResponse{
		Validators:      rpcValidators,
		TotalValidators: int64(len(allValidators)),
		Page:            body.Page,
		TotalPages:      int64(totalPages),
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
