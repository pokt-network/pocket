package state_sync

import (
	"fmt"
	"sync"
	"time"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/shared/modules"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	stateSyncModuleName       = "stateSyncModule"
	committedBlocsChannelSize = 100
	blockWaitingPeriod        = 30 * time.Second
	metadataChannelSize       = 1000
)

type StateSyncModule interface {
	modules.Module
	StateSyncServerModule

	HandleStateSyncBlockCommittedEvent(message *anypb.Any) error
	HandleStateSyncMetadataResponse(metadataRes *typesCons.StateSyncMetadataResponse)

	// SetActiveSyncHeight sets the active sync height
	SetActiveSyncHeight(height uint64)

	// CatchMsgHeight starts active state sync from current height to activeSyncHeight
	CatchMsgHeight()
}

var (
	_ modules.Module        = &stateSync{}
	_ StateSyncModule       = &stateSync{}
	_ StateSyncServerModule = &stateSync{}
)

type stateSync struct {
	bus                    modules.Bus
	logger                 *modules.Logger
	committedBlocksChannel chan uint64
	activeSyncHeight       uint64
	wg                     sync.WaitGroup

	// metadata responses received from peers are collected in this channel
	metadataReceived chan *typesCons.StateSyncMetadataResponse
}

func CreateStateSync(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	return new(stateSync).Create(bus, options...)
}

func (*stateSync) Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	m := &stateSync{}

	for _, option := range options {
		option(m)
	}

	bus.RegisterModule(m)

	m.logger = logger.Global.CreateLoggerForModule(m.GetModuleName())
	m.committedBlocksChannel = make(chan uint64, committedBlocsChannelSize)
	m.metadataReceived = make(chan *typesCons.StateSyncMetadataResponse, metadataChannelSize)

	return m, nil
}

func (m *stateSync) SetActiveSyncHeight(height uint64) {
	if height > m.activeSyncHeight {
		m.activeSyncHeight = height
	}
}

// Start performs passive state sync process, starting from the consensus module's current height to the aggragated metadata height
func (m *stateSync) Start() error {
	m.wg.Add(2)
	go func() {
		defer m.wg.Done()
		m.metadataSyncLoop()
	}()
	go func() {
		defer m.wg.Done()
		m.blockRequestLoop()
	}()

	return nil
}

// Stop stops the state sync process, and sends `Consensus_IsSyncedValidator` FSM event
func (m *stateSync) Stop() error {
	close(m.metadataReceived)
	m.wg.Wait()
	return nil
}

func (m *stateSync) SetBus(pocketBus modules.Bus) {
	m.bus = pocketBus
}

func (m *stateSync) GetBus() modules.Bus {
	if m.bus == nil {
		logger.Global.Fatal().Msg("PocketBus is not initialized")
	}
	return m.bus
}

func (m *stateSync) GetModuleName() string {
	return stateSyncModuleName
}

func (m *stateSync) HandleStateSyncBlockCommittedEvent(event *anypb.Any) error {
	evt, err := codec.GetCodec().FromAny(event)
	if err != nil {
		return err
	}

	switch event.MessageName() {
	case messaging.StateSyncBlockCommittedEventType:
		newCommitBlockEvent, ok := evt.(*typesCons.StateSyncBlockCommittedEvent)
		if !ok {
			return fmt.Errorf("failed to cast event to StateSyncBlockCommittedEvent")
		}

		m.committedBlocksChannel <- newCommitBlockEvent.Height
	}
	return nil
}

func (m *stateSync) HandleStateSyncMetadataResponse(metadataRes *typesCons.StateSyncMetadataResponse) {
	m.metadataReceived <- metadataRes
}

// TODO(#352): Implement this function, currently a placeholder.
// metadataSyncLoop periodically sends metadata requests to its peers, as part of passive state sync process
func (m *stateSync) metadataSyncLoop() {
	// runs as a background process
	// requests metadata from peers
	// sends received metadata to the metadataReceived channel
}

// TODO(#352): Implement this function, currently a placeholder.
// blockRequestLoop periodically sends metadata requests to its peers, as part of passive state sync process
func (m *stateSync) blockRequestLoop() {
	// runs as a background process
	// aggregates metadata in the metadataReceived channel
	// requests blocks from the current height to the aggregated metadata height
	// sends received blocks to the blockReceived channel
}

// CatchMsgHeight performs active state sync, starting from the consensus module's current height to the activeSyncHeight
func (m *stateSync) CatchMsgHeight() {
	consensusMod := m.bus.GetConsensusModule()
	currentHeight := consensusMod.CurrentHeight()
	nodeAddress := consensusMod.GetNodeAddress()
	readCtx, err := m.GetBus().GetPersistenceModule().NewReadContext(int64(currentHeight))
	if err != nil {
		m.logger.Err(err).Msg("failed to create read context for persistence module")
		return
	}
	defer readCtx.Release()

	//get the current validators
	validators, err := readCtx.GetAllValidators(int64(currentHeight))
	if err != nil {
		m.logger.Err(err).Msg("failed to get all validators from persistence module")
		return
	}

	// requests blocks from the current height to the aggregated metadata height
	for currentHeight <= m.activeSyncHeight {
		m.logger.Info().Msgf("Sync is requesting block: %d, ending height: %d", currentHeight, m.activeSyncHeight)

		// form the get block request message
		stateSyncGetBlockMessage := &typesCons.StateSyncMessage{
			Message: &typesCons.StateSyncMessage_GetBlockReq{
				GetBlockReq: &typesCons.GetBlockRequest{
					PeerAddress: nodeAddress,
					Height:      currentHeight,
				},
			},
		}

		// broadcast the get block request message to all validators
		for _, val := range validators {
			if err := m.sendStateSyncMessage(stateSyncGetBlockMessage, cryptoPocket.AddressFromString(val.GetAddress())); err != nil {
				m.logger.Err(err).Msg("failed to send block request message")
				return
			}
		}

		// wait for the requested block to be received and committed by consensus module
		<-m.committedBlocksChannel

		// requested block is received and committed, continue to the next block from the current height
		currentHeight = consensusMod.CurrentHeight()
	}

	m.logger.Info().Msg("Active syncing is complete!")

	if err = m.GetBus().GetStateMachineModule().SendEvent(coreTypes.StateMachineEvent_Consensus_IsSyncedValidator); err != nil {
		m.logger.Err(err).Msg("failed to send IsSyncedValidator event to the FSM")
	}
}
