package consensus

// TODO: Split this file into multiple helpers (e.g. signatures.go, hotstuff_helpers.go, etc...)
import (
	"encoding/base64"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"google.golang.org/protobuf/proto"
)

// These constants and variables are wrappers around the autogenerated protobuf types and were
// added to simply make the code in the `consensus` module more readable.
const (
	NewRound  = typesCons.HotstuffStep_HOTSTUFF_STEP_NEWROUND
	Prepare   = typesCons.HotstuffStep_HOTSTUFF_STEP_PREPARE
	PreCommit = typesCons.HotstuffStep_HOTSTUFF_STEP_PRECOMMIT
	Commit    = typesCons.HotstuffStep_HOTSTUFF_STEP_COMMIT
	Decide    = typesCons.HotstuffStep_HOTSTUFF_STEP_DECIDE

	Propose = typesCons.HotstuffMessageType_HOTSTUFF_MESSAGE_PROPOSE
	Vote    = typesCons.HotstuffMessageType_HOTSTUFF_MESSAGE_VOTE

	ByzantineThreshold = float64(2) / float64(3)
)

var HotstuffSteps = [...]typesCons.HotstuffStep{NewRound, Prepare, PreCommit, Commit, Decide}

// ** Hotstuff Helpers ** //

// IMPROVE: Avoid having the `ConsensusModule` be a receiver of this; making it more functional.
// TODO: Add unit tests for all quorumCert creation & validation logic...
func (m *consensusModule) getQuorumCertificate(height uint64, step typesCons.HotstuffStep, round uint64) (*typesCons.QuorumCertificate, error) {
	var pss []*typesCons.PartialSignature
	for !m.hotstuffMempool[step].IsEmpty() {
		msg, err := m.hotstuffMempool[step].Pop()
		if err != nil {
			return nil, err
		}
		if msg.GetPartialSignature() == nil {

			m.logger.Warn().Fields(
				map[string]any{
					"height": msg.GetHeight(),
					"step":   msg.GetStep(),
					"round":  msg.GetRound(),
				},
			).Msg("No partial signature found which should not happen...")

			continue
		}
		if msg.GetHeight() != height || msg.GetStep() != step || msg.GetRound() != round {

			m.logger.Warn().Fields(
				map[string]any{
					"height": msg.GetHeight(),
					"step":   msg.GetStep(),
					"round":  msg.GetRound(),
				},
			).Msg("Message in pool does not match (height, step, round) of QC being generated")

			continue
		}

		ps := msg.GetPartialSignature()
		if ps.Signature == nil || ps.Address == "" {

			m.logger.Warn().Fields(
				map[string]any{
					"height": msg.GetHeight(),
					"step":   msg.GetStep(),
					"round":  msg.GetRound(),
				},
			).Msg("Partial signature is incomplete which should not happen...")
			continue
		}
		pss = append(pss, msg.GetPartialSignature())
	}

	validators, err := m.getValidatorsAtHeight(height)
	if err != nil {
		return nil, err
	}

	if err := m.isOptimisticThresholdMet(len(pss), validators); err != nil {
		return nil, err
	}

	thresholdSig := getThresholdSignature(pss)

	return &typesCons.QuorumCertificate{
		Height:             height,
		Step:               step,
		Round:              round,
		Block:              m.block,
		ThresholdSignature: thresholdSig,
	}, nil
}

func (m *consensusModule) findHighQC(msgs []*typesCons.HotstuffMessage) (qc *typesCons.QuorumCertificate) {
	for _, m := range msgs {
		if m.GetQuorumCertificate() == nil {
			continue
		}
		// TODO: Make sure to validate the "highest QC" first and add tests
		if qc == nil || m.GetQuorumCertificate().Height > qc.Height {
			qc = m.GetQuorumCertificate()
		}
	}
	return
}

func getThresholdSignature(partialSigs []*typesCons.PartialSignature) *typesCons.ThresholdSignature {
	thresholdSig := new(typesCons.ThresholdSignature)
	thresholdSig.Signatures = make([]*typesCons.PartialSignature, len(partialSigs))
	copy(thresholdSig.Signatures, partialSigs)
	return thresholdSig
}

func isSignatureValid(msg *typesCons.HotstuffMessage, pubKeyString string, signature []byte) bool {
	pubKey, err := cryptoPocket.NewPublicKey(pubKeyString)
	if err != nil {
		logger.Global.Warn().Err(err).Msgf("Error getting PublicKey from bytes")
		return false
	}
	bytesToVerify, err := getSignableBytes(msg)
	if err != nil {
		logger.Global.Warn().Err(err).Msgf("Error getting bytes to verify")
		return false
	}
	return pubKey.Verify(bytesToVerify, signature)
}

func (m *consensusModule) didReceiveEnoughMessageForStep(step typesCons.HotstuffStep) error {
	validators, err := m.getValidatorsAtHeight(m.CurrentHeight())
	if err != nil {
		return err
	}
	return m.isOptimisticThresholdMet(int(m.hotstuffMempool[step].Size()), validators)
}

func (m *consensusModule) isOptimisticThresholdMet(numSignatures int, validators []*coreTypes.Actor) error {
	numValidators := len(validators)
	if !(float64(numSignatures) > ByzantineThreshold*float64(numValidators)) {
		return typesCons.ErrByzantineThresholdCheck(numSignatures, ByzantineThreshold*float64(numValidators))
	}
	return nil
}

func protoHash(m proto.Message) string {
	b, err := codec.GetCodec().Marshal(m)
	if err != nil {
		logger.Global.Fatal().Err(err).Msg("Could not marshal proto message")
	}
	return base64.StdEncoding.EncodeToString(b)
}

/*** P2P Helpers ***/

func (m *consensusModule) sendToLeader(msg *typesCons.HotstuffMessage) {
	m.logger.Info().Fields(
		map[string]any{
			"node_id": m.leaderId,
			"height":  msg.GetHeight(),
			"step":    msg.GetStep(),
			"round":   msg.GetRound(),
		},
	).Msg("✉️ Sending message ✉️")

	// TODO: This can happen due to a race condition with the pacemaker.
	if m.leaderId == nil {
		m.logger.Error().Msg(typesCons.ErrNilLeaderId.Error())
		return
	}

	anyConsensusMessage, err := codec.GetCodec().ToAny(msg)
	if err != nil {
		m.logger.Error().Err(err).Msg(typesCons.ErrCreateConsensusMessage.Error())
		return
	}

	validators, err := m.getValidatorsAtHeight(m.CurrentHeight())
	if err != nil {
		m.logger.Error().Err(err).Msg(typesCons.ErrPersistenceGetAllValidators.Error())
	}

	idToValAddrMap := typesCons.NewActorMapper(validators).GetIdToValAddrMap()

	if err := m.GetBus().GetP2PModule().Send(cryptoPocket.AddressFromString(idToValAddrMap[*m.leaderId]), anyConsensusMessage); err != nil {
		m.logger.Error().Err(err).Msg(typesCons.ErrSendMessage.Error())
		return
	}
}

// Star-like (O(n)) broadcast - send to all nodes directly
// INVESTIGATE: Re-evaluate if we should be using our structured broadcast (RainTree O(log3(n))) algorithm instead
func (m *consensusModule) broadcastToValidators(msg *typesCons.HotstuffMessage) {
	m.logger.Info().Fields(
		map[string]any{
			"height": m.CurrentHeight(),
			"step":   m.step,
			"round":  m.round,
		},
	).Msg("📣 Broadcasting message 📣")

	anyConsensusMessage, err := codec.GetCodec().ToAny(msg)
	if err != nil {
		m.logger.Error().Err(err).Msg(typesCons.ErrCreateConsensusMessage.Error())
		return
	}

	validators, err := m.getValidatorsAtHeight(m.CurrentHeight())
	if err != nil {
		m.logger.Error().Err(err).Msg(typesCons.ErrPersistenceGetAllValidators.Error())
	}

	for _, val := range validators {
		if err := m.GetBus().GetP2PModule().Send(cryptoPocket.AddressFromString(val.GetAddress()), anyConsensusMessage); err != nil {
			m.logger.Error().Err(err).Msg(typesCons.ErrBroadcastMessage.Error())
		}
	}
}

/*** Persistence Helpers ***/

// TECHDEBT(#388): Integrate this with the `persistence` module or a real mempool.
func (m *consensusModule) clearMessagesPool() {
	for _, step := range HotstuffSteps {
		m.hotstuffMempool[step].Clear()
	}
}

func (m *consensusModule) initMessagesPool() {
	for _, step := range HotstuffSteps {
		m.hotstuffMempool[step] = NewHotstuffFIFOMempool(m.consCfg.MaxMempoolBytes)
	}
}

/*** Leader Election Helpers ***/
func (m *consensusModule) isReplica() bool {
	return !m.IsLeader()
}

func (m *consensusModule) clearLeader() {
	m.logPrefix = DefaultLogPrefix
	m.leaderId = nil
}

func (m *consensusModule) electNextLeader(message *typesCons.HotstuffMessage) error {
	leaderId, err := m.leaderElectionMod.ElectNextLeader(message)
	if err != nil || leaderId == 0 {

		m.logger.Error().Err(err).Fields(
			map[string]any{
				"leaderId": leaderId,
				"height":   m.height,
				"round":    m.round,
			},
		).Msg("leader election failed: Validator cannot take part in consensus")

		m.clearLeader()
		return err
	}
	m.leaderId = &leaderId

	validators, err := m.getValidatorsAtHeight(m.CurrentHeight())
	if err != nil {
		return err
	}

	idToValAddrMap := typesCons.NewActorMapper(validators).GetIdToValAddrMap()

	if m.IsLeader() {
		m.setLogPrefix("LEADER")
		m.logger.Info().Fields(
			map[string]any{
				"leaderId": idToValAddrMap[*m.leaderId],
				"height":   m.height,
				"round":    m.round,
			},
		).Msg("👑 I am the leader 👑")
	} else {
		m.setLogPrefix("REPLICA")
		m.logger.Info().Fields(
			map[string]any{
				"leaderId": idToValAddrMap[*m.leaderId],
				"height":   m.height,
				"round":    m.round,
			},
		).Msg("🙇 Elected leader 🙇")
	}

	return nil
}

/*** General Infrastructure Helpers ***/
func (m *consensusModule) setLogPrefix(logPrefix string) {
	logger.Global.UpdateFields(map[string]any{
		"kind": logPrefix,
	})
	// INVESTIGATE: Do we need to create a new logger here?
	m.logger = logger.Global.CreateLoggerForModule("consensus")
}

func (m *consensusModule) getValidatorsAtHeight(height uint64) ([]*coreTypes.Actor, error) {
	persistenceReadContext, err := m.GetBus().GetPersistenceModule().NewReadContext(int64(height))
	if err != nil {
		return nil, err
	}
	defer persistenceReadContext.Close()

	return persistenceReadContext.GetAllValidators(int64(height))
}

// TODO: This is a temporary solution, cache this in Consensus module. This field will be populated once with a single query to the persistence module.
func (m *consensusModule) IsValidator() (bool, error) {
	validators, err := m.getValidatorsAtHeight(m.CurrentHeight())
	if err != nil {
		return false, err
	}

	for _, actor := range validators {
		if actor.Address == m.nodeAddress {
			return true, nil
		}
	}

	return false, nil
}

// CONSIDER: Below are same as the ones on statesync helper. We should probably move them to a common place.
func (m *consensusModule) logHelper(receiverPeerId string) map[string]any {
	return map[string]any{
		"height":         m.CurrentHeight(),
		"senderPeerId":   m.GetNodeAddress(),
		"receiverPeerId": receiverPeerId,
	}
}

func (m *consensusModule) maximumPersistedBlockHeight() (uint64, error) {
	currentHeight := m.CurrentHeight()
	persistenceContext, err := m.GetBus().GetPersistenceModule().NewReadContext(int64(currentHeight))
	if err != nil {
		return 0, err
	}
	defer persistenceContext.Close()

	maxHeight, err := persistenceContext.GetMaximumBlockHeight()
	if err != nil {
		return 0, err
	}

	return maxHeight, nil
}
