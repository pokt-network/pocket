package rpc

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math"
	"regexp"
	"strings"

	conTypes "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/utility"
	utilTypes "github.com/pokt-network/pocket/utility/types"
)

var (
	paramValueRegex *regexp.Regexp
	errNoItems      = fmt.Errorf("no items found")
)

func init() {
	paramValueRegex = regexp.MustCompile(`value:"(.+)"`)
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

func checkSort(sort string) string {
	if strings.ToLower(sort) == "asc" {
		return "asc"
	}
	return "desc"
}

func getPageIndexes(totalItems, page, per_page int) (startIdx, endIdx, totalPages int, err error) {
	if totalItems == 0 {
		err = errNoItems
		return
	}
	if per_page > 1000 {
		err = fmt.Errorf("per_page has a max value of 1000")
		return
	}
	if page == 0 || per_page == 0 {
		err = fmt.Errorf("page and per_page must both be greater than 0")
		return
	}

	totalPages = int(math.Ceil(float64(totalItems) / float64(per_page)))
	startIdx = (page - 1) * per_page
	if startIdx > totalItems-1 {
		err = fmt.Errorf("starting page too high: got %d, total pages: %d", page, totalPages)
		return
	}
	endIdx = (page * per_page) - 1
	if endIdx >= totalItems {
		endIdx = totalItems - 1 //  Last Index
	}

	return startIdx, endIdx, totalPages, nil
}

// protocolActorToRPCProtocolActor converts the coreTypes.Actor to an RPC ProtocolActor
func protocolActorToRPCProtocolActor(actor *coreTypes.Actor) ProtocolActor {
	return ProtocolActor{
		Address:         actor.Address,
		ActorType:       protocolActorToRPCActorTypeEnum(actor.ActorType),
		PublicKey:       actor.PublicKey,
		Chains:          actor.Chains,
		ServiceUrl:      actor.ServiceUrl,
		StakedAmount:    actor.StakedAmount,
		PausedHeight:    actor.PausedHeight,
		UnstakingHeight: actor.UnstakingHeight,
		OutputAddr:      actor.Output,
	}
}

// txResultToRPCTransaction converts the txResult protobuf into the RPC Transaction type
func (s *rpcServer) txResultToRPCTransaction(txResult *coreTypes.TxResult) (*Transaction, error) {
	hash := coreTypes.TxHash(txResult.GetTx())
	txStr := base64.StdEncoding.EncodeToString(txResult.GetTx())
	stdTx, err := s.transactionBytesToRPCStdTx(txResult.GetTx(), txResult.GetMessageType())
	if err != nil {
		return nil, err
	}
	return &Transaction{
		Hash:   hash,
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
	}, nil
}

// transactionBytesToRPCStdTx generates a StdTx from a serialised byte slice of a Transaction protobuf and message type
func (s *rpcServer) transactionBytesToRPCStdTx(txBz []byte, messageType string) (*StdTx, error) {
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
		fee, err := s.calculateMessageFeeForActor(m.GetActorType(), messageType)
		if err != nil {
			return nil, err
		}
		stdTx.Fee = Fee{
			Amount: fee,
			Denom:  "upokt",
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
		fee, err := s.calculateMessageFeeForActor(m.GetActorType(), messageType)
		if err != nil {
			return nil, err
		}
		stdTx.Fee = Fee{
			Amount: fee,
			Denom:  "upokt",
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
		fee, err := s.calculateMessageFeeForActor(m.GetActorType(), messageType)
		if err != nil {
			return nil, err
		}
		stdTx.Fee = Fee{
			Amount: fee,
			Denom:  "upokt",
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
		fee, err := s.calculateMessageFeeForActor(m.GetActorType(), messageType)
		if err != nil {
			return nil, err
		}
		stdTx.Fee = Fee{
			Amount: fee,
			Denom:  "upokt",
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
		fee, err := s.calculateMessageFeeForActor(m.GetActorType(), messageType)
		if err != nil {
			return nil, err
		}
		stdTx.Fee = Fee{
			Amount: fee,
			Denom:  "upokt",
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
		fee, err := s.calculateMessageFeeForActor(m.GetActorType(), messageType)
		if err != nil {
			return nil, err
		}
		stdTx.Fee = Fee{
			Amount: fee,
			Denom:  "upokt",
		}
		values := paramValueRegex.FindStringSubmatch(m.GetParameterValue().String())
		if len(values) < 2 {
			return nil, fmt.Errorf("unable to extract parameter value: %s", m.GetParameterValue().String())
		}
		stdTx.Message = MessageChangeParameter{
			Signer: hex.EncodeToString(m.GetSigner()),
			Owner:  hex.EncodeToString(m.GetOwner()),
			Parameter: Parameter{
				ParameterValue: values[1],
			},
		}
	default:
		return nil, fmt.Errorf("unknown message type: %s", messageType)
	}

	return &stdTx, nil
}

// calculateMessageFeeForActor calculates the fee for a transaction given the actor type and message type
func (s *rpcServer) calculateMessageFeeForActor(actorType coreTypes.ActorType, messageType string) (string, error) {
	height := int64(s.GetBus().GetConsensusModule().CurrentHeight())
	readCtx, err := s.GetBus().GetPersistenceModule().NewReadContext(height)
	if err != nil {
		return "", err
	}
	if messageType == "MessageSend" {
		return readCtx.GetStringParam(utilTypes.MessageSendFee, height)
	}
	if messageType == "MessageChangeParameter" {
		return readCtx.GetStringParam(utilTypes.MessageChangeParameterFee, height)
	}
	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		switch messageType {
		case "MessageStake":
			return readCtx.GetStringParam(utilTypes.MessageStakeAppFee, height)
		case "MessageEditStake":
			return readCtx.GetStringParam(utilTypes.MessageEditStakeAppFee, height)
		case "MessageUnstake":
			return readCtx.GetStringParam(utilTypes.MessageUnstakeAppFee, height)
		case "MessageUnpause":
			return readCtx.GetStringParam(utilTypes.MessageUnpauseAppFee, height)
		}
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		switch messageType {
		case "MessageStake":
			return readCtx.GetStringParam(utilTypes.MessageStakeFishermanFee, height)
		case "MessageEditStake":
			return readCtx.GetStringParam(utilTypes.MessageEditStakeFishermanFee, height)
		case "MessageUnstake":
			return readCtx.GetStringParam(utilTypes.MessageUnstakeFishermanFee, height)
		case "MessageUnpause":
			return readCtx.GetStringParam(utilTypes.MessageUnpauseFishermanFee, height)
		}
	case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
		switch messageType {
		case "MessageStake":
			return readCtx.GetStringParam(utilTypes.MessageStakeServicerFee, height)
		case "MessageEditStake":
			return readCtx.GetStringParam(utilTypes.MessageEditStakeServicerFee, height)
		case "MessageUnstake":
			return readCtx.GetStringParam(utilTypes.MessageUnstakeServicerFee, height)
		case "MessageUnpause":
			return readCtx.GetStringParam(utilTypes.MessageUnpauseServicerFee, height)
		}
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		switch messageType {
		case "MessageStake":
			return readCtx.GetStringParam(utilTypes.MessageStakeValidatorFee, height)
		case "MessageEditStake":
			return readCtx.GetStringParam(utilTypes.MessageEditStakeValidatorFee, height)
		case "MessageUnstake":
			return readCtx.GetStringParam(utilTypes.MessageUnstakeValidatorFee, height)
		case "MessageUnpause":
			return readCtx.GetStringParam(utilTypes.MessageUnpauseValidatorFee, height)
		}
	default:
		return "", fmt.Errorf("invalid actor type: %s", actorType.GetName())
	}
	return "", fmt.Errorf("unhandled message type: %s", messageType)
}

// txProtoBytesToRPCTransactions converts a slice of serialised Transaction protobufs to a slice of RPC transactions
func (s *rpcServer) txProtoBytesToRPCTransactions(txProtoBytes [][]byte) ([]Transaction, error) {
	currentHeight := s.GetBus().GetConsensusModule().CurrentHeight()
	uow, err := s.GetBus().GetUtilityModule().NewUnitOfWork(int64(currentHeight))
	if err != nil {
		return nil, err
	}
	defer uow.Release() //nolint:errcheck // We only need to make sure the UOW is released

	txs := make([]Transaction, 0)
	for idx, txBz := range txProtoBytes {
		tx := new(coreTypes.Transaction)
		if err := codec.GetCodec().Unmarshal(txBz, tx); err != nil {
			return nil, err
		}
		txResult, er := uow.HydrateTxResult(tx, idx)
		if er != nil {
			return nil, er
		}
		rpcTx, err := s.txResultToRPCTransaction(txResult)
		if err != nil {
			return nil, err
		}
		txs = append(txs, *rpcTx)
	}

	return txs, nil
}

// blockToRPCBlock converts a block protobuf to the RPC block type
func (s *rpcServer) blockToRPCBlock(protoBlock *coreTypes.Block) (*Block, error) {
	txs, err := s.txProtoBytesToRPCTransactions(protoBlock.GetTransactions())
	if err != nil {
		return nil, err
	}

	qc := new(conTypes.QuorumCertificate)
	if err := codec.GetCodec().Unmarshal(protoBlock.BlockHeader.GetQuorumCertificate(), qc); err != nil {
		return nil, err
	}
	partialSigs := make([]PartialSignature, 0)
	for _, sig := range qc.GetThresholdSignature().GetSignatures() {
		ps := PartialSignature{
			Signature: hex.EncodeToString(sig.GetSignature()),
			Address:   sig.GetAddress(),
		}
		partialSigs = append(partialSigs, ps)
	}

	qcTxs := make([]string, 0)
	for _, txBz := range qc.GetBlock().GetTransactions() {
		tx := base64.StdEncoding.EncodeToString(txBz)
		qcTxs = append(qcTxs, tx)
	}

	qcBlockBz, err := codec.GetCodec().Marshal(qc.GetBlock())
	if err != nil {
		return nil, err
	}
	qcBlock := base64.StdEncoding.EncodeToString(qcBlockBz)

	return &Block{
		BlockHeader: BlockHeader{
			Height:        int64(protoBlock.BlockHeader.GetHeight()),
			NetworkId:     protoBlock.BlockHeader.GetNetworkId(),
			StateHash:     protoBlock.BlockHeader.GetStateHash(),
			PrevStateHash: protoBlock.BlockHeader.GetPrevStateHash(),
			ProposerAddr:  hex.EncodeToString(protoBlock.BlockHeader.GetProposerAddress()),
			QuorumCert: QuorumCertificate{
				Height: int64(qc.GetHeight()),
				Round:  int64(qc.GetRound()),
				Step:   qc.GetStep().String(),
				Block:  qcBlock,
				ThresholdSig: ThresholdSignature{
					Signatures: partialSigs,
				},
				Transactions: qcTxs,
			},
			Timestamp: protoBlock.BlockHeader.GetTimestampt().AsTime().String(),
		},
		Transactions: txs,
	}, nil
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
	var protocolActorGetter = persistenceContext.GetAllStakedActors
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
