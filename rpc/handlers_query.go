package rpc

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/utils"
)

// This file contains the handlers for the v1/query path in the RPC specification
// It pertains to all the user facing/public RPC endpoints to query the pocket
// network, its actors and state.

const (
	denom = "upokt"
)

func (s *rpcServer) PostV1QueryAccount(ctx echo.Context) error {
	var body QueryAccountHeight
	if err := ctx.Bind(&body); err != nil {
		return ctx.String(http.StatusBadRequest, "bad request")
	}

	height := s.getQueryHeight(body.Height)
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
		Coins:   []Coin{{Amount: amount, Denom: denom}},
	})
}

func (s *rpcServer) PostV1QueryAccounts(ctx echo.Context) error {
	var body QueryHeightPaginated
	if err := ctx.Bind(&body); err != nil {
		return ctx.String(http.StatusBadRequest, "bad request")
	}

	height := s.getQueryHeight(body.Height)
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
			Coins:   []Coin{{Amount: account.Amount, Denom: denom}},
		})
	}

	return ctx.JSON(http.StatusOK, QueryAccountsResponse{
		Result:     accounts,
		Page:       body.Page,
		TotalPages: int64(totalPages),
	})
}

func (s *rpcServer) PostV1QueryAccountTxs(ctx echo.Context) error {
	var body QueryAccountPaginated
	if err := ctx.Bind(&body); err != nil {
		return ctx.String(http.StatusBadRequest, "bad request")
	}
	sortDesc := checkSortDesc(*body.Sort)

	txIndexer := s.GetBus().GetPersistenceModule().GetTxIndexer()
	idxTxs, err := txIndexer.GetBySender(body.Address, sortDesc)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	start, end, totalPages, err := getPageIndexes(len(idxTxs), int(body.Page), int(body.PerPage))
	if err != nil && !errors.Is(err, errNoItems) {
		return ctx.String(http.StatusBadRequest, err.Error())
	}
	if totalPages == 0 || errors.Is(err, errNoItems) {
		return ctx.JSON(http.StatusOK, QueryAccountTxsResponse{})
	}

	pageTxs := make([]IndexedTransaction, 0)
	for _, idxTx := range idxTxs[start : end+1] {
		rpcTx, err := s.idxTxToRPCIdxTx(idxTx)
		if err != nil {
			return ctx.String(http.StatusInternalServerError, err.Error())
		}
		pageTxs = append(pageTxs, *rpcTx)
	}

	return ctx.JSON(http.StatusOK, QueryAccountTxsResponse{
		Txs:        pageTxs,
		Page:       body.Page,
		TotalTxs:   int64(len(idxTxs)),
		TotalPages: int64(totalPages),
	})
}

func (s *rpcServer) GetV1QueryAllChainParams(ctx echo.Context) error {
	height := s.getQueryHeight(0)
	readCtx, err := s.GetBus().GetPersistenceModule().NewReadContext(height)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	defer readCtx.Release() //nolint:errcheck // We only need to make sure the readCtx is released

	paramsSlice, err := readCtx.GetAllParams()
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	resp := make([]Parameter, 0)
	for i := 0; i < len(paramsSlice); i++ {
		resp = append(resp, Parameter{
			ParameterName:  paramsSlice[i][0],
			ParameterValue: paramsSlice[i][1],
		})
	}
	return ctx.JSON(200, resp)
}

func (s *rpcServer) PostV1QueryApp(ctx echo.Context) error {
	var body QueryAccountHeight
	if err := ctx.Bind(&body); err != nil {
		return ctx.String(http.StatusBadRequest, "bad request")
	}

	height := s.getQueryHeight(body.Height)
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

	height := s.getQueryHeight(body.Height)
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

	rpcApps := protocolActorToRPCProtocolActors(allApps[start : end+1])

	return ctx.JSON(http.StatusOK, QueryAppsResponse{
		Apps:       rpcApps,
		TotalApps:  int64(len(allApps)),
		Page:       body.Page,
		TotalPages: int64(totalPages),
	})
}

func (s *rpcServer) PostV1QueryBalance(ctx echo.Context) error {
	var body QueryAccountHeight
	if err := ctx.Bind(&body); err != nil {
		return ctx.String(http.StatusBadRequest, "bad request")
	}

	height := s.getQueryHeight(body.Height)
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

	height := uint64(s.getQueryHeight(body.Height))
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

	height := uint64(s.getQueryHeight(body.Height))
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
	var body QueryAccountHeight
	if err := ctx.Bind(&body); err != nil {
		return ctx.String(http.StatusBadRequest, "bad request")
	}

	height := s.getQueryHeight(body.Height)
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

	height := s.getQueryHeight(body.Height)
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

	rpcFishermen := protocolActorToRPCProtocolActors(allFishermen[start : end+1])

	return ctx.JSON(http.StatusOK, QueryFishermenResponse{
		Fishermen:      rpcFishermen,
		TotalFishermen: int64(len(allFishermen)),
		Page:           body.Page,
		TotalPages:     int64(totalPages),
	})
}

func (s *rpcServer) GetV1QueryHeight(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, QueryHeight{
		Height: s.getQueryHeight(0),
	})
}

func (s *rpcServer) PostV1QueryParam(ctx echo.Context) error {
	var body QueryParameter
	if err := ctx.Bind(&body); err != nil {
		return ctx.String(http.StatusBadRequest, "bad request")
	}

	height := s.getQueryHeight(body.Height)
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
	var body QueryAccountHeight
	if err := ctx.Bind(&body); err != nil {
		return ctx.String(http.StatusBadRequest, "bad request")
	}

	height := s.getQueryHeight(body.Height)
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

	height := s.getQueryHeight(body.Height)
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

	rpcServicers := protocolActorToRPCProtocolActors(allServicers[start : end+1])

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

	height := s.getQueryHeight(body.Height)
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
			Denom:   denom,
		})
	}

	return ctx.JSON(http.StatusOK, QuerySupplyResponse{
		Pools: rpcPools,
		Total: Coin{
			Amount: total.String(),
			Denom:  denom,
		},
	})
}

func (s *rpcServer) PostV1QuerySupportedChains(ctx echo.Context) error {
	var body QueryHeight
	if err := ctx.Bind(&body); err != nil {
		return ctx.String(http.StatusBadRequest, "bad request")
	}

	height := s.getQueryHeight(body.Height)
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
	idxTx, err := txIndexer.GetByHash(hashBz)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	rpcTx, err := s.idxTxToRPCIdxTx(idxTx)
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

	rpcUncTxs, err := s.txProtoBytesToRPCIdxTxs([][]byte{uncTx})
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

	rpcUncTxs, err := s.txProtoBytesToRPCIdxTxs(uncTxs[start : end+1])
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

	height := s.getQueryHeight(body.Height)
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
	var body QueryAccountHeight
	if err := ctx.Bind(&body); err != nil {
		return ctx.String(http.StatusBadRequest, "bad request")
	}

	height := s.getQueryHeight(body.Height)
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

	height := s.getQueryHeight(body.Height)
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

	rpcValidators := protocolActorToRPCProtocolActors(allValidators[start : end+1])

	return ctx.JSON(http.StatusOK, QueryValidatorsResponse{
		Validators:      rpcValidators,
		TotalValidators: int64(len(allValidators)),
		Page:            body.Page,
		TotalPages:      int64(totalPages),
	})
}

func (s *rpcServer) PostV1QueryNodeRoles(ctx echo.Context) error {
	actorModules := s.GetBus().GetUtilityModule().GetActorModules()
	roles := make([]string, 0)
	for _, m := range actorModules {
		roles = append(roles, m.GetModuleName())
	}
	return ctx.JSON(200, QueryNodeRolesResponse{NodeRoles: roles})
}
