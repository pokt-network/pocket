package rpc

import (
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/pokt-network/pocket/app"
	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/utils"
	"github.com/pokt-network/pocket/utility"
	"golang.org/x/exp/slices"
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
	currentHeight := int64(s.GetBus().GetConsensusModule().CurrentHeight())
	if currentHeight > 0 {
		currentHeight -= 1
	}
	height := body.Height
	if height == 0 {
		height = currentHeight
	}
	readCtx, err := s.GetBus().GetPersistenceModule().NewReadContext(currentHeight)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
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
	currentHeight := int64(s.GetBus().GetConsensusModule().CurrentHeight())
	if currentHeight > 0 {
		currentHeight -= 1
	}
	height := body.Height
	if height == 0 {
		height = currentHeight
	}
	readCtx, err := s.GetBus().GetPersistenceModule().NewReadContext(currentHeight)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
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

func (s *rpcServer) PostV1QueryAccounttxs(ctx echo.Context) error {
	var body QueryAddressPaginated
	if err := ctx.Bind(&body); err != nil {
		return ctx.String(http.StatusBadRequest, "bad request")
	}
	if body.Sort != nil && *body.Sort != "asc" && *body.Sort != "desc" {
		return ctx.String(http.StatusBadRequest, "sort must be either asc or desc")
	}
	sortDesc := true
	if body.Sort != nil && *body.Sort == "asc" {
		sortDesc = false
	}

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
	currentHeight := int64(s.GetBus().GetConsensusModule().CurrentHeight())
	if currentHeight > 0 {
		currentHeight -= 1
	}
	height := body.Height
	if height == 0 {
		height = currentHeight
	}
	readCtx, err := s.GetBus().GetPersistenceModule().NewReadContext(currentHeight)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	allApps, err := readCtx.GetAllApps(height)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	idx := slices.IndexFunc(allApps, func(app *coreTypes.Actor) bool { return app.Address == body.Address })
	if idx == -1 {
		return ctx.String(http.StatusBadRequest, fmt.Sprintf("no app found with address: %s", body.Address))
	}

	return ctx.JSON(http.StatusOK, App{
		Address:         allApps[idx].Address,
		ActorType:       protocolActorToRPCActorTypeEnum(allApps[idx].ActorType),
		PublicKey:       allApps[idx].PublicKey,
		Chains:          allApps[idx].Chains,
		StakedAmount:    allApps[idx].StakedAmount,
		PausedHeight:    allApps[idx].PausedHeight,
		UnstakingHeight: allApps[idx].UnstakingHeight,
		OutputAddr:      allApps[idx].Output,
	})
}

func (s *rpcServer) PostV1QueryApps(ctx echo.Context) error {
	var body QueryHeightPaginated
	if err := ctx.Bind(&body); err != nil {
		return ctx.String(http.StatusBadRequest, "bad request")
	}

	// Get latest stored block height
	currentHeight := int64(s.GetBus().GetConsensusModule().CurrentHeight())
	if currentHeight > 0 {
		currentHeight -= 1
	}
	height := body.Height
	if height == 0 {
		height = currentHeight
	}
	readCtx, err := s.GetBus().GetPersistenceModule().NewReadContext(currentHeight)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
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

	rpcApps := make([]App, 0)
	for _, app := range allApps[start : end+1] {
		rpcApps = append(rpcApps, App{
			Address:         app.Address,
			ActorType:       protocolActorToRPCActorTypeEnum(app.ActorType),
			PublicKey:       app.PublicKey,
			Chains:          app.Chains,
			StakedAmount:    app.StakedAmount,
			PausedHeight:    app.PausedHeight,
			UnstakingHeight: app.UnstakingHeight,
			OutputAddr:      app.Output,
		})
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
	currentHeight := int64(s.GetBus().GetConsensusModule().CurrentHeight())
	if currentHeight > 0 {
		currentHeight -= 1
	}
	height := body.Height
	if height == 0 {
		height = currentHeight
	}
	readCtx, err := s.GetBus().GetPersistenceModule().NewReadContext(currentHeight)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

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
	currentHeight := s.GetBus().GetConsensusModule().CurrentHeight()
	if currentHeight > 0 {
		currentHeight -= 1
	}
	height := uint64(body.Height)
	if height == 0 || height > currentHeight {
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

func (s *rpcServer) PostV1QueryBlocktxs(ctx echo.Context) error {
	var body QueryHeightPaginated
	if err := ctx.Bind(&body); err != nil {
		return ctx.String(http.StatusBadRequest, "bad request")
	}
	if body.Sort != nil && *body.Sort != "asc" && *body.Sort != "desc" {
		return ctx.String(http.StatusBadRequest, "sort must be either asc or desc")
	}
	sortDesc := true
	if body.Sort != nil && *body.Sort == "asc" {
		sortDesc = false
	}

	// Get latest stored block height
	currentHeight := s.GetBus().GetConsensusModule().CurrentHeight()
	if currentHeight > 0 {
		currentHeight -= 1
	}
	height := uint64(body.Height)
	if height == 0 || height > currentHeight {
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

func (s *rpcServer) PostV1QueryUnconfirmedtx(ctx echo.Context) error {
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

func (s *rpcServer) PostV1QueryUnconfirmedtxs(ctx echo.Context) error {
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

// Broadcast to the entire validator set
func (s *rpcServer) broadcastMessage(msgBz []byte) error {
	utilityMsg, err := utility.PrepareTxGossipMessage(msgBz)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to prepare transaction gossip message")
		return err
	}

	if err := s.GetBus().GetP2PModule().Broadcast(utilityMsg); err != nil {
		s.logger.Error().Err(err).Msg("Failed to broadcast utility message")
		return err
	}

	return nil
}
