package dkg

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/gob"
	"log"
	"math/big"
	"sort"

	"pocket/consensus/pkg/types"
	"pocket/shared"
	"pocket/shared/context"

	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/coinbase/kryptology/pkg/dkg/gennaro"
	v1 "github.com/coinbase/kryptology/pkg/sharing/v1"
)

// TODO: Call this directly from the app specific bus.
func (module *dkgModule) HandleMessage(ctx *context.PocketContext, m *DKGMessage) {
	if m.Recipient != nil && m.Recipient == &module.NodeId {
		log.Printf("[%d] Received message from %d but should not have for DKG Round %d.\n", module.NodeId, m.Sender, m.Round)
		return
	}

	switch m.Round {
	case DKGRound1:
		module.handleDKGRound1()
	case DKGRound2:
		module.handleDKGRound2(m)
	case DKGRound3:
		module.handleDKGRound3(m)
	case DKGRound4:
		module.handleDKGRound4(m)
	}
}

func (module *dkgModule) handleDKGRound1() {
	log.Printf("[%d][DEBUG] Handling DKG Messages Round1 \n", module.NodeId)

	if module.DKGParticipant == nil {
		log.Printf("[ERROR][%d] handleDKGRound1 called with nil DKGParticipant...\n", module.NodeId)
		return
	}

	bcast, p2psend, err := module.DKGParticipant.Round1(nil)
	if err != nil {
		log.Printf("[ERROR][%d] Failed to perform round 1... %v \n", module.NodeId, err)
		return
	}

	broadcastMessage := &DKGMessage{
		Round:       DKGRound2,
		MessageData: encodeBroadcast1(bcast),
		MessageType: DKGBroadcast,
		Sender:      module.NodeId,
	}
	module.broadcastToNodes(broadcastMessage)

	valMap := shared.GetPocketState().ValidatorMap
	for nodeIdNumber, p2pPacket := range p2psend {
		dest, ok := valMap[types.NodeId(nodeIdNumber)]
		if !ok {
			log.Printf("[ERROR][%d] Failed to find nodeId %d in ValMap...\n", module.NodeId, nodeIdNumber)
			continue
		}
		p2pSendMessage := &DKGMessage{
			Round:       DKGRound2,
			MessageData: encodeSend1(p2pPacket),
			MessageType: DKGP2PSend,
			Sender:      module.NodeId,
			Recipient:   &dest.NodeId,
		}
		module.sendToNode(p2pSendMessage, &dest.NodeId)
	}
}

func (module *dkgModule) handleDKGRound2(m *DKGMessage) {
	log.Printf("[%d][DEBUG] Handling DKG Messages Round2 \n", module.NodeId)

	valMap := shared.GetPocketState().ValidatorMap

	module.DKGMessagePool[m.Round] = append(module.DKGMessagePool[m.Round], *m)
	numExpectedMessages := len(valMap)*2 - 1 // 1 bcast (including self) and 1 send from every validator. TODO: Add byzantine threshold.
	if len(module.DKGMessagePool[m.Round]) < numExpectedMessages {
		return
	}
	log.Printf("[%d] Starting DKG Round 2...\n", module.NodeId)

	rnd1Bcast := make(map[uint32]gennaro.Round1Bcast, len(valMap))
	rnd1P2p := make(map[uint32]*gennaro.Round1P2PSendPacket, len(valMap))
	for _, m := range module.DKGMessagePool[m.Round] {
		src := uint32(m.Sender)
		switch m.MessageType {
		case DKGBroadcast:
			rnd1Bcast[src] = decodeBroadcast1(m.MessageData)
		case DKGP2PSend:
			rnd1P2p[src] = decodeSend1(m.MessageData)
		}

	}
	module.DKGMessagePool[m.Round] = nil

	bcast2, err := module.DKGParticipant.Round2(rnd1Bcast, rnd1P2p)
	if err != nil {
		log.Printf("[ERROR][%d] Failed to perform round 2... %v\n", module.NodeId, err)
		return
	}

	broadcastMessage := &DKGMessage{
		Round:       DKGRound3,
		MessageData: encodeBroadcast2(bcast2),
		MessageType: DKGBroadcast,
		Sender:      module.NodeId,
	}
	module.broadcastToNodes(broadcastMessage)
}

func (module *dkgModule) handleDKGRound3(m *DKGMessage) {
	log.Printf("[%d][DEBUG] Handling DKG Messages Round3 \n", module.NodeId)

	valMap := shared.GetPocketState().ValidatorMap

	module.DKGMessagePool[m.Round] = append(module.DKGMessagePool[m.Round], *m)
	if len(module.DKGMessagePool[m.Round]) != len(valMap) {
		return
	}
	log.Printf("[%d] Starting DKG Round 3...\n", module.NodeId)

	rnd2Bcast := make(map[uint32]gennaro.Round2Bcast, len(valMap))
	for _, m := range module.DKGMessagePool[m.Round] {
		rnd2Bcast[uint32(m.Sender)] = decodeBroadcast2(m.MessageData)
	}
	module.DKGMessagePool[m.Round] = nil

	verificationKey, secretKey, err := module.DKGParticipant.Round3(rnd2Bcast)
	if err != nil {
		log.Printf("[ERROR][%d] Failed to perform round 3... %v \n", module.NodeId, err)
		return
	}

	module.ThresholdSigningKey = secretKey // TODO: Is this necessary?

	broadcastMessage := &DKGMessage{
		Round:       DKGRound4,
		MessageData: encodeBroadcast3(verificationKey),
		MessageType: DKGBroadcast,
		Sender:      module.NodeId,
	}
	module.broadcastToNodes(broadcastMessage)
}

func (module *dkgModule) handleDKGRound4(m *DKGMessage) {
	log.Printf("[%d][DEBUG] Handling DKG Messages Round4 \n", module.NodeId)

	valMap := shared.GetPocketState().ValidatorMap

	module.DKGMessagePool[m.Round] = append(module.DKGMessagePool[m.Round], *m)
	if len(module.DKGMessagePool[m.Round]) != len(valMap) {
		return
	}

	rnd3Bcast := make(map[uint32]*gennaro.Round3Bcast, len(valMap))
	for _, m := range module.DKGMessagePool[m.Round] {
		rnd3Bcast[uint32(m.Sender)] = decodeBroadcast3(m.MessageData)
	}
	module.DKGMessagePool[m.Round] = nil

	_, err := module.DKGParticipant.Round4()
	if err != nil {
		log.Printf("[ERROR][%d] Failed to perform round 3... %v \n", module.NodeId, err)
		return
	}
}

func (module *dkgModule) addNewDKGParticipant() {
	valMap := shared.GetPocketState().ValidatorMap

	valIdx := uint32(module.NodeId)
	threshold := uint32(2) // TODO: 2/3 of the ValMap?
	scalar := curves.NewEd25519Scalar()
	curve := v1.Ed25519()
	generator, err := curves.NewScalarBaseMult(curve, getRandomBigInt())
	if err != nil {
		log.Fatalf("Failed to create generator: %v", err)
	}
	otherIds := make([]uint32, len(valMap)-1) // -1 because we're assuming the current validator is in the set.
	idx := 0                                  // TODO: Loop is confusing because the node idx doesn't necessary equal index in otherIds array.
	for id := range valMap {
		if valIdx == uint32(id) {
			continue
		}
		otherIds[idx] = uint32(id)
		idx++
	}
	sort.Slice(otherIds, func(i, j int) bool { return otherIds[i] < otherIds[j] }) // TODO: Is this necessary?

	participant, err := gennaro.NewParticipant(valIdx, threshold, generator, scalar, otherIds...)
	if err != nil {
		log.Fatalf("Failed to create participant: %v", err)
	}
	module.DKGParticipant = participant
}

func getRandomBigInt() *big.Int {
	// Max random value, a 130-bits integer, i.e 2^130 - 1
	max := new(big.Int)
	max.Exp(big.NewInt(2), big.NewInt(130), nil).Sub(max, big.NewInt(1))

	//Generate cryptographically strong pseudo-random between 0 - max
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		log.Fatalf("Failed to generate random number: %v", err)
	}

	// HACK - TODO: This value must be the same for all participants, so it must either
	// be a shared nonce that is distributed or part of genesis.
	n = big.NewInt(10)

	return n
}

type RoundBroadcastWrapper struct {
	PointsByteArrs [][]byte
}

func encodeBroadcast1(bcast gennaro.Round1Bcast) []byte {
	bBytes := make([][]byte, len(bcast))
	for idx, b := range bcast {
		bBytes[idx] = b.Bytes()
	}
	bcastMessage := &RoundBroadcastWrapper{
		PointsByteArrs: bBytes,
	}
	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)
	if err := enc.Encode(bcastMessage); err != nil {
		log.Fatalf("Failed to encode message: %v", err)
	}
	return buff.Bytes()
}

func decodeBroadcast1(data []byte) (bcast gennaro.Round1Bcast) {
	var messageBuff = bytes.NewBuffer(data)
	dec := gob.NewDecoder(messageBuff)
	m := new(RoundBroadcastWrapper)
	if err := dec.Decode(m); err != nil {
		log.Println("[ERROR] Failed to decode message:", err)
	}

	for _, pointBytes := range m.PointsByteArrs {
		point, err := curves.PointFromBytesUncompressed(v1.Ed25519(), pointBytes)
		if err != nil {
			log.Println("Failed to decode point: ", err)
			continue
		}
		bcast = append(bcast, point)
	}
	return
}

type RoundP2PSendWrapper struct {
	SecretShare   []byte
	BlindingShare []byte
}

func shamirShareFromBytes(data []byte) (shamirShare *v1.ShamirShare) {
	field := curves.NewField(v1.Ed25519().Params().N)
	identifier := binary.BigEndian.Uint32(data[:4])
	return v1.NewShamirShare(identifier, data[4:], field)
}

func encodeSend1(p2pSend *gennaro.Round1P2PSendPacket) []byte {
	p2pPacketWrapper := &RoundP2PSendWrapper{
		SecretShare:   p2pSend.SecretShare.Bytes(),
		BlindingShare: p2pSend.BlindingShare.Bytes(),
	}

	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)
	if err := enc.Encode(p2pPacketWrapper); err != nil {
		log.Fatalf("Failed to encode message: %v", err)
	}

	return buff.Bytes()
}

func decodeSend1(data []byte) *gennaro.Round1P2PSendPacket {
	var messageBuff = bytes.NewBuffer(data)
	dec := gob.NewDecoder(messageBuff)
	p2pPacketWrapper := new(RoundP2PSendWrapper)
	if err := dec.Decode(p2pPacketWrapper); err != nil {
		log.Println("[ERROR] Failed to decode message:", err)
	}

	p2pPacket := &gennaro.Round1P2PSendPacket{
		SecretShare:   shamirShareFromBytes(p2pPacketWrapper.SecretShare),
		BlindingShare: shamirShareFromBytes(p2pPacketWrapper.BlindingShare),
	}

	return p2pPacket
}

func encodeBroadcast2(bcast gennaro.Round2Bcast) []byte {
	return encodeBroadcast1(bcast) // Different types defined by kryptology but they are equivalent.
}

func decodeBroadcast2(data []byte) gennaro.Round2Bcast {
	return decodeBroadcast1(data) // Different types defined by kryptology but they are equivalent.
}

func encodeBroadcast3(bcast *gennaro.Round3Bcast) []byte {
	return bcast.Bytes()
}

func decodeBroadcast3(data []byte) (bcase *gennaro.Round3Bcast) {
	point, err := curves.PointFromBytesUncompressed(v1.Ed25519(), data)
	if err != nil {
		log.Println("Failed to decode point: ", err)
		return nil
	}
	return point
}
