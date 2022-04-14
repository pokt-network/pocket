package shared

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/jackc/pgx/v4"

	"github.com/pokt-network/pocket/p2p/pre2p"
	"github.com/pokt-network/pocket/shared/config"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/utility"

	"github.com/pokt-network/pocket/consensus"
	"github.com/pokt-network/pocket/persistence/pre_persistence"
	"github.com/pokt-network/pocket/shared/types"
	typesGenesis "github.com/pokt-network/pocket/shared/types/genesis"

	"github.com/pokt-network/pocket/shared/modules"
)

var _ modules.Module = &Node{}

type Node struct {
	bus modules.Bus

	Address cryptoPocket.Address
}

func pocketDbStuff() {
	ctx := context.TODO()

	url := "postgres://postgres:postgres@pocket-db:5432/postgres"
	conn, err := pgx.Connect(context.Background(), url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	schema := os.Getenv("POSTGRES_SCHEMA")
	conn.Exec(ctx, fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schema))
	conn.Exec(ctx, fmt.Sprintf("set search_path TO %s", schema))

	_, err = conn.Exec(ctx, `
		create table IF NOT EXISTS users (id int);`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create users table: %v\n", err)
		os.Exit(1)
	}

	rand.Seed(time.Now().UnixNano())
	_, err = conn.Exec(ctx, fmt.Sprintf("INSERT INTO users (id) VALUES (%d)", rand.Int31()))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to insert random int into users table: %v\n", err)
		os.Exit(1)
	}
}

func Create(cfg *config.Config) (n *Node, err error) {
	// TODO(design): initialize the state singleton until we have a proper solution for this.
	_ = typesGenesis.GetNodeState(cfg)

	pocketDbStuff()

	// persistenceMod, err := persistence.Create(cfg)
	prePersistenceMod, err := pre_persistence.Create(cfg)
	if err != nil {
		return nil, err
	}
	// TODO(derrandz): Replace with real P2P module
	// p2pMod, err := p2p.Create(cfg)
	pre2pMod, err := pre2p.Create(cfg)
	if err != nil {
		return nil, err
	}

	// TODO(andrew): Replace with real Utility module
	utilityMod, err := utility.Create(cfg)
	// mockedUtilityMod, err := utility.CreateMockedModule(cfg)
	if err != nil {
		return nil, err
	}

	consensusMod, err := consensus.Create(cfg)
	if err != nil {
		return nil, err
	}

	bus, err := CreateBus(prePersistenceMod, pre2pMod, utilityMod, consensusMod)
	if err != nil {
		return nil, err
	}

	return &Node{
		bus:     bus,
		Address: cfg.PrivateKey.Address(),
	}, nil
}

func (node *Node) Start() error {
	log.Println("Starting pocket node...")

	// NOTE: Order of module startup here matters.

	if err := node.GetBus().GetPersistenceModule().Start(); err != nil {
		return err
	}

	if err := node.GetBus().GetP2PModule().Start(); err != nil {
		return err
	}

	if err := node.GetBus().GetUtilityModule().Start(); err != nil {
		return err
	}

	if err := node.GetBus().GetConsensusModule().Start(); err != nil {
		return err
	}

	// TODO(olshansky): discuss if we need a special type/event for this.
	signalNodeStartedEvent := &types.PocketEvent{Topic: types.PocketTopic_POCKET_NODE_TOPIC, Data: nil}
	node.GetBus().PublishEventToBus(signalNodeStartedEvent)

	// While loop lasting throughout the entire lifecycle of the node.
	for {
		event := node.GetBus().GetBusEvent()
		if err := node.handleEvent(event); err != nil {
			log.Println("Error handling event: ", err)
		}
	}
}

func (node *Node) Stop() error {
	log.Println("Stopping pocket node...")
	return nil
}

func (m *Node) SetBus(bus modules.Bus) {
	m.bus = bus
}

func (m *Node) GetBus() modules.Bus {
	if m.bus == nil {
		log.Fatalf("PocketBus is not initialized")
	}
	return m.bus
}
