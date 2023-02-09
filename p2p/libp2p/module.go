package libp2p

import (
	"context"
	"io"
	"log"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"
	"google.golang.org/protobuf/proto"

	"github.com/pokt-network/pocket/p2p/common"
	"github.com/pokt-network/pocket/p2p/libp2p/identity"
	"github.com/pokt-network/pocket/p2p/stdnetwork"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/shared/modules"
)

type libp2pModule struct {
	bus          modules.Bus
	cfg          *configs.P2PConfig
	identity     libp2p.Option
	listenAddrs  libp2p.Option
	host         host.Host
	pubsub       *pubsub.PubSub
	topic        *pubsub.Topic
	subscription *pubsub.Subscription
	network      typesP2P.Network
}

// NewReadStreamDeadline returns a future deadline
// based on the read stream timeout duration.
func NewReadStreamDeadline() time.Time {
	return time.Now().Add(readStreamTimeoutDuration)
}

var (
	// TODO: consider security exposure of and "safe minimum" for timeout.
	// TODO: parameterize and expose via config.
	// ReadStreamTimeout is the duration to wait for a read operation on a
	// stream to complete, after which the stream is closed ("timed out").
	readStreamTimeoutDuration = time.Second * 10
	// ErrModule wraps errors which occur within the libp2pModule implementation.
	ErrModule = common.NewErrFactory("LibP2P module error")
	// ErrCloseStream wraps errors which occur when attempting to close a peer stream.
	ErrCloseStream = common.NewErrFactory("an error occurred while closing peer stream")
	// PoktProtocolID is the libp2p protocol ID matching current version of the pokt protocol.
	PoktProtocolID = protocol.ID("pokt/v1.0.0")
	// DefaultTopicStr is a "default" pubsub topic string for use when subscribing.
	DefaultTopicStr = "pokt/default"
)

func Create(bus modules.Bus) (modules.Module, error) {
	return new(libp2pModule).Create(bus)
}

func (mod *libp2pModule) GetModuleName() string {
	// TODO: double check if this should change.
	return modules.P2PModuleName
}

func (mod *libp2pModule) Create(bus modules.Bus) (modules.Module, error) {
	*mod = libp2pModule{}
	if err := bus.RegisterModule(mod); err != nil {
		return nil, ErrModule("unable to register module", err)
	}

	log.Println("Creating libp2p-backed network module")

	// TODO: support RainTree network
	if mod.cfg.UseRainTree {
		return nil, ErrModule("raintree is not yet compatible with libp2p", nil)
	}

	// TODO: how should this effect things?
	//if !cfg.ClientDebugMode {
	//}

	runtimeMgr := bus.GetRuntimeMgr()
	cfg := runtimeMgr.GetConfig()
	mod.cfg = cfg.P2P

	// TODO(future): investigate any unnecessary
	// key exposure / duplication in memory
	secretKey, err := identity.NewLibP2PPrivateKey(mod.cfg.PrivateKey)
	if err != nil {
		// TODO: wrap error
		return nil, err
	}

	mod.identity = libp2p.Identity(secretKey)
	mod.listenAddrs = libp2p.ListenAddrStrings()

	if err := bus.RegisterModule(mod); err != nil {
		// TODO: wrap error
		return nil, ErrModule("unable to register module", err)
	}

	return mod, nil
}

func (mod *libp2pModule) Start() error {
	// TODO: receive context in interface methods?
	ctx := context.Background()

	// TODO: metrics integration.
	var err error

	mod.host, err = libp2p.New(
		mod.identity,
		mod.listenAddrs,
		// TODO: transport security!
	)
	if err != nil {
		return ErrModule("unable to create libp2p host", err)
	}

	mod.host.SetStreamHandler(PoktProtocolID, mod.handleStream)

	// TODO: use RandomSub or GossipSub once we're on more stable ground.
	// TODO: consider supporting multiple router types via config.
	mod.pubsub, err = pubsub.NewFloodSub(ctx, mod.host)
	if err != nil {
		return ErrModule("unable to create pubsub", err)
	}

	mod.topic, err = mod.pubsub.Join(DefaultTopicStr)
	if err != nil {
		return ErrModule("unable to join pubsub topic", err)
	}

	mod.subscription, err = mod.topic.Subscribe()
	if err != nil {
		return ErrModule("unable to subscribe to pubsub topic", err)
	}

	mod.network, err = stdnetwork.NewLibp2pNetwork(mod.bus, mod.host, mod.topic)
	if err != nil {
		return ErrModule("unable to create network", err)
	}

	go mod.readFromSubscription(ctx)

	return nil
}

func (mod *libp2pModule) Stop() error {
	return mod.host.Close()
}

func (mod *libp2pModule) SetBus(bus modules.Bus) {
	// INVESTIGATE: Can the code flow be modified to set the bus here?
	// m.network.SetBus(m.GetBus())
	mod.bus = bus
}

func (mod *libp2pModule) GetBus() modules.Bus {
	if mod.bus == nil {
		log.Printf("[WARN]: PocketBus is not initialized")
		return nil
	}
	return mod.bus
}

// handleStream is called each time a peer establishes a new stream with this
// module's libp2p `host.Host`.
func (mod *libp2pModule) handleStream(stream network.Stream) {
	poktPeer, err := identity.PoktPeerFromStream(stream)
	if err != nil {
		// TODO: conventional error logging.
		log.Printf("%s", ErrModule("unable to parse remote peer's public key", err))

		if err = stream.Close(); err != nil {
			log.Printf("%s", ErrCloseStream("in libp2pModule#handleStream", err))
		}
	}

	if err := mod.network.AddPeerToAddrBook(poktPeer); err != nil {
		// TODO: conventional error logging.
		log.Printf("%s", ErrModule("unable to parse remote peer's public key", err))
	}

	go mod.readStream(stream)
}

func (mod *libp2pModule) readStream(stream network.Stream) {
	// NB: time out if no data is sent to free resources.
	if err := stream.SetReadDeadline(NewReadStreamDeadline()); err != nil {
		// TODO: conventional error logging.
		log.Printf("%s", ErrModule("unable to read from stream", err))
	}

	data, err := io.ReadAll(stream)
	if err != nil {
		// TODO: conventional error logging.
		log.Printf("%s", ErrModule("unable to read from stream", err))
		if err := stream.Close(); err != nil {
			log.Printf("%s", ErrCloseStream("in libp2pModule#readStream", err))
		}
		// NB: abort this goroutine
		// TODO: signal this somewhere?
		return
	}

	mod.handleNetworkData(data)
}

func (mod *libp2pModule) readFromSubscription(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := mod.subscription.Next(ctx)
			if err != nil {
				// TODO: is there a more conventional way to log (e.g. logging module)?
				log.Printf("%s", ErrModule("unable to read from subscription", err))
			}

			// NB: ignore messages from self
			if msg.ReceivedFrom == mod.host.ID() {
				continue
			}

			mod.handleNetworkData(msg.Data)
		}
	}
}

func (mod *libp2pModule) handleNetworkData(data []byte) {
	appMsgData, err := mod.network.HandleNetworkData(data)
	if err != nil {
		// TODO: is there a more conventional way to log (e.g. logging module)?
		log.Printf("%s", ErrModule("", err))
	}

	// There was no error, but we don't need to forward this to the app-specific bus.
	// For example, the message has already been handled by the application.
	if appMsgData == nil {
		return
	}

	networkMessage := messaging.PocketEnvelope{}
	if err := proto.Unmarshal(appMsgData, &networkMessage); err != nil {
		// TODO: is there a more conventional way to log (e.g. logging module)?
		log.Printf("%s", ErrModule("Error decoding network message", err))
		return
	}

	event := messaging.PocketEnvelope{
		Content: networkMessage.Content,
	}

	mod.GetBus().PublishEventToBus(&event)
}
