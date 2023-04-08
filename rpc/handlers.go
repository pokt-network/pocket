package rpc

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math"
	"net/http"
	"regexp"

	"github.com/labstack/echo/v4"
	"github.com/pokt-network/pocket/app"
	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/utility"
	utilTypes "github.com/pokt-network/pocket/utility/types"
)

var (
	paramValueRegex *regexp.Regexp
)

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
	if body.Sort != nil && !(*body.Sort == "asc" || *body.Sort == "desc") {
		return ctx.String(http.StatusBadRequest, "sort must be either asc or desc")
	}
	sortDesc := true
	if *body.Sort == "asc" {
		sortDesc = false
	}

	txIndexer := s.GetBus().GetPersistenceModule().GetTxIndexer()
	txResults, err := txIndexer.GetBySender(body.Address, sortDesc)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	totalPages := uint64(math.Ceil(float64(len(txResults)) / float64(body.PerPage)))
	//start := (body.Page - 1) * body.PerPage
	end := (body.Page * body.PerPage) - 1
	totalTxs := int64(len(txResults))
	fmt.Println(totalTxs)
	if end >= totalTxs {
		end = totalTxs
	}

	pageTxs := make([]Transaction, 0)
	//for i := start; i < end; i++ {
	//txResult := txResults[i]
	for _, txResult := range txResults {
		hashBz, err := txResult.Hash()
		if err != nil {
			return ctx.String(http.StatusInternalServerError, err.Error())
		}
		hexHashStr := hex.EncodeToString(hashBz)
		txStr := base64.StdEncoding.EncodeToString(txResult.GetTx())
		if err != nil {
			return ctx.String(http.StatusInternalServerError, err.Error())
		}
		stdTx, err := generateStdTx(txResult.GetTx(), txResult.GetMessageType())
		if err != nil {
			return ctx.String(http.StatusInternalServerError, err.Error())
		}
		pageTxs = append(pageTxs, Transaction{
			Hash:   hexHashStr,
			Height: txResult.GetHeight(),
			Index:  txResult.GetIndex(),
			TxResult: TxResult{
				Tx:            txStr,
				Height:        txResult.GetHeight(),
				Index:         txResult.GetIndex(),
				ResultCode:    txResult.GetResultCode(),
				SignerAddr:    txResult.GetSignerAddr(),
				RecipientAddr: txResult.GetRecipientAddr(),
				MessageType:   txResult.GetMessageType(),
			},
			StdTx: *stdTx,
		})
	}

	return ctx.JSON(http.StatusOK, QueryAccountTxsResponse{
		Txs:        pageTxs,
		Page:       body.Page,
		TotalTxs:   totalTxs,
		TotalPages: totalPages,
	})
}

func generateStdTx(txBz []byte, messageType string) (*StdTx, error) {
	tx, err := coreTypes.TxFromBytes(txBz)
	if err != nil {
		return nil, err
	}
	sig := tx.GetSignature()
	txMsg, err := tx.GetMessage()
	if err != nil {
		return nil, err
	}
	anypb, err := codec.GetCodec().ToAny(txMsg)
	if err != nil {
		return nil, err
	}
	stdTx := StdTx{
		Nonce: tx.GetNonce(),
		Signature: Signature{
			PublicKey: hex.EncodeToString(sig.GetPublicKey()),
			Signature: hex.EncodeToString(sig.GetSignature()),
		},
	}
	switch messageType {
	case "MessageSend":
		m := new(utilTypes.MessageSend)
		if err := anypb.UnmarshalTo(m); err != nil {
			return nil, err
		}
		stdTx.Message = MessageSend{
			FromAddr: hex.EncodeToString(m.GetFromAddress()),
			ToAddr:   hex.EncodeToString(m.GetToAddress()),
			Amount:   m.Amount,
			Denom:    "upokt",
		}
	case "MessageStake":
		m := new(utilTypes.MessageStake)
		if err := anypb.UnmarshalTo(m); err != nil {
			return nil, err
		}
		stdTx.Message = MessageStake{
			ActorType:     protocolActorToRPCActorTypeEnum(m.GetActorType()),
			PublicKey:     hex.EncodeToString(m.GetPublicKey()),
			Chains:        m.GetChains(),
			ServiceUrl:    m.GetServiceUrl(),
			OutputAddress: hex.EncodeToString(m.GetOutputAddress()),
			Signer:        hex.EncodeToString(m.GetSigner()),
			Amount:        m.GetAmount(),
			Denom:         "upokt",
		}
	case "MessageEditStake":
		m := new(utilTypes.MessageEditStake)
		if err := anypb.UnmarshalTo(m); err != nil {
			return nil, err
		}
		stdTx.Message = MessageEditStake{
			ActorType:  protocolActorToRPCActorTypeEnum(m.GetActorType()),
			Chains:     m.GetChains(),
			ServiceUrl: m.GetServiceUrl(),
			Address:    hex.EncodeToString(m.GetAddress()),
			Signer:     hex.EncodeToString(m.GetSigner()),
			Amount:     m.GetAmount(),
			Denom:      "upokt",
		}
	case "MessageUnstake":
		m := new(utilTypes.MessageUnstake)
		if err := anypb.UnmarshalTo(m); err != nil {
			return nil, err
		}
		stdTx.Message = MessageUnstake{
			ActorType: protocolActorToRPCActorTypeEnum(m.GetActorType()),
			Address:   hex.EncodeToString(m.GetAddress()),
			Signer:    hex.EncodeToString(m.GetSigner()),
		}
	case "MessageUnpause":
		m := new(utilTypes.MessageUnpause)
		if err := anypb.UnmarshalTo(m); err != nil {
			return nil, err
		}
		stdTx.Message = MessageUnpause{
			ActorType: protocolActorToRPCActorTypeEnum(m.GetActorType()),
			Address:   hex.EncodeToString(m.GetAddress()),
			Signer:    hex.EncodeToString(m.GetSigner()),
		}
	case "MessageChangeParameter":
		m := new(utilTypes.MessageChangeParameter)
		if err := anypb.UnmarshalTo(m); err != nil {
			return nil, err
		}
		values := paramValueRegex.FindStringSubmatch(m.GetParameterValue().String())
		value := ""
		if len(values) > 1 {
			value = values[1]
		}
		stdTx.Message = MessageChangeParameter{
			Signer: hex.EncodeToString(m.GetSigner()),
			Owner:  hex.EncodeToString(m.GetOwner()),
			Parameter: Parameter{
				ParameterName:  m.GetParameterKey(),
				ParameterValue: value,
			},
		}
	default:
		return nil, fmt.Errorf("unknown message type: %s", messageType)
	}

	return &stdTx, nil
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

// protocolActorToRPCActorTypeEnum converts a protocol actor type to the rpc actor type enum
func protocolActorToRPCActorTypeEnum(protocolActorType coreTypes.ActorType) ActorTypesEnum {
	switch protocolActorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		return Application
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		return Fisherman
	case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
		return Servicer
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		return Validator
	default:
		panic("invalid actor type")
	}
}

// getProtocolActorGetter returns the correct protocol actor getter function based on the actor type parameter
func getProtocolActorGetter(persistenceContext modules.PersistenceReadContext, params GetV1P2pStakedActorsAddressBookParams) func(height int64) ([]*coreTypes.Actor, error) {
	var protocolActorGetter func(height int64) ([]*coreTypes.Actor, error) = persistenceContext.GetAllStakedActors
	if params.ActorType == nil {
		return persistenceContext.GetAllStakedActors
	}
	switch *params.ActorType {
	case Application:
		protocolActorGetter = persistenceContext.GetAllApps
	case Fisherman:
		protocolActorGetter = persistenceContext.GetAllFishermen
	case Servicer:
		protocolActorGetter = persistenceContext.GetAllServicers
	case Validator:
		protocolActorGetter = persistenceContext.GetAllValidators
	}
	return protocolActorGetter
}
