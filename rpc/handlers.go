package rpc

import (
	"encoding/hex"
	"fmt"
	"math"
	"net/http"
	"regexp"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/pokt-network/pocket/app"
	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/utils"
	"github.com/pokt-network/pocket/utility"
)

var paramValueRegex *regexp.Regexp

func init() {
	paramValueRegex = regexp.MustCompile(`value:"(.+)"`)
}

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

	// get the account from the persistence module
	currentHeight := s.GetBus().GetConsensusModule().CurrentHeight()
	readCtx, err := s.GetBus().GetPersistenceModule().NewReadContext(int64(currentHeight))
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	accBz, err := hex.DecodeString(body.Address)
	if err != nil {
		return ctx.String(http.StatusBadRequest, err.Error())
	}
	height := body.Height
	if height == 0 {
		height = currentHeight
	}
	amount, err := readCtx.GetAccountAmount(accBz, int64(height))
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
	if body.Page == 0 || body.PerPage == 0 {
		return ctx.String(http.StatusBadRequest, "page and per_page must be greater than 0")
	}

	// get the account from the persistence module
	currentHeight := s.GetBus().GetConsensusModule().CurrentHeight()
	readCtx, err := s.GetBus().GetPersistenceModule().NewReadContext(int64(currentHeight))
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	height := body.Height
	if height == 0 {
		height = currentHeight
	}
	allAccounts, err := readCtx.GetAllAccounts(int64(height))
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	totalPages := uint64(math.Ceil(float64(len(allAccounts)) / float64(body.PerPage)))
	start := (body.Page - 1) * body.PerPage
	end := (body.Page * body.PerPage) - 1
	accounts := make([]Account, 0)
	for i := start; i <= end; i++ {
		accounts = append(accounts, Account{
			Address: allAccounts[i].Address,
			Coins:   []Coin{{Amount: allAccounts[i].Amount, Denom: "upokt"}},
		})
	}

	return ctx.JSON(http.StatusOK, QueryAccountsResponse{
		Result:     accounts,
		Page:       body.Page,
		TotalPages: totalPages,
	})
}

func (s *rpcServer) PostV1QueryAccounttxs(ctx echo.Context) error {
	var body QueryAddressPaginated
	if err := ctx.Bind(&body); err != nil {
		return ctx.String(http.StatusBadRequest, "bad request")
	}
	if body.Page == 0 || body.PerPage == 0 {
		return ctx.String(http.StatusBadRequest, "page and per_page must be greater than 0")
	}
	if body.PerPage > 1000 {
		return ctx.String(http.StatusBadRequest, "per_page has a max value of 1000")
	}
	if body.Sort != nil && *body.Sort != "asc" && *body.Sort != "desc" {
		return ctx.String(http.StatusBadRequest, "sort must be either asc or desc")
	}
	sortDesc := true
	if *body.Sort == "asc" {
		sortDesc = false
	}

	currHeight := s.GetBus().GetConsensusModule().CurrentHeight()
	readCtx, err := s.GetBus().GetPersistenceModule().NewReadContext(int64(currHeight))
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	// TODO: (h5law) figure out how to query all transactions from an address
	txIndexer := s.GetBus().GetPersistenceModule().GetTxIndexer()
	txResults, err := txIndexer.GetByHeight(3, sortDesc)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	totalPages := uint64(math.Ceil(float64(len(txResults)) / float64(body.PerPage)))
	start := (body.Page - 1) * body.PerPage
	if int(start) > len(txResults)-1 {
		return ctx.String(http.StatusBadRequest, fmt.Sprintf("starting page too large: got %d, max: %d", body.Page, totalPages))
	}
	end := (body.Page * body.PerPage) - 1
	totalTxs := int64(len(txResults))
	if end >= totalTxs {
		end = totalTxs - 1
	}

	pageTxs := make([]Transaction, 0)
	for i := start; i <= end; i++ {
		rpcTx, err := txResultToRPCTransaction(readCtx, txResults[i])
		if err != nil {
			return ctx.String(http.StatusInternalServerError, err.Error())
		}
		pageTxs = append(pageTxs, *rpcTx)
	}

	return ctx.JSON(http.StatusOK, QueryAccountTxsResponse{
		Txs:        pageTxs,
		Page:       body.Page,
		TotalTxs:   totalTxs,
		TotalPages: totalPages,
	})
}

func (s *rpcServer) GetV1QueryAllChainParams(ctx echo.Context) error {
	currHeight := s.GetBus().GetConsensusModule().CurrentHeight()
	readCtx, err := s.GetBus().GetPersistenceModule().NewReadContext(int64(currHeight))
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

func (s *rpcServer) PostV1QueryBalance(ctx echo.Context) error {
	var body QueryAddressHeight
	if err := ctx.Bind(&body); err != nil {
		return ctx.String(http.StatusBadRequest, "bad request")
	}

	currentHeight := s.GetBus().GetConsensusModule().CurrentHeight()
	readCtx, err := s.GetBus().GetPersistenceModule().NewReadContext(int64(currentHeight))
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	accBz, err := hex.DecodeString(body.Address)
	if err != nil {
		return ctx.String(http.StatusBadRequest, err.Error())
	}
	height := body.Height
	if height == 0 {
		height = currentHeight
	}
	amountStr, err := readCtx.GetAccountAmount(accBz, int64(height))
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

	currentHeight := s.GetBus().GetConsensusModule().CurrentHeight()
	height := uint64(body.Height)
	if height == 0 {
		height = currentHeight
	}

	// TODO: (h5law) figure out how to get transactions from the block protobuf
	blockStore := s.GetBus().GetPersistenceModule().GetBlockStore()
	blockBz, err := blockStore.Get(utils.HeightToBytes(height))
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	block := new(coreTypes.Block)
	if err := codec.GetCodec().Unmarshal(blockBz, block); err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	rpcBlock, err := blockToRPCBlock(block)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, rpcBlock)
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
