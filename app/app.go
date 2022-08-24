package app

import (
	"log"
	"sync"

	"github.com/pokt-network/pocket/consensus"
	"github.com/pokt-network/pocket/p2p"
	"github.com/pokt-network/pocket/shared"
	"github.com/pokt-network/pocket/shared/config"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/telemetry"
	"github.com/pokt-network/pocket/shared/types/genesis"
)

const (
	configPath  string = "build/config/config1.json"
	genesisPath string = "build/config/genesis.json"
)

var (
	// A P2P module is initialized in order to broadcast a message to the local network
	p2pMod modules.P2PModule

	// A consensus module is initialized in order to get a list of the validator network
	consensusMod modules.ConsensusModule

	modInitOnce sync.Once
)

func Init(remoteCLIURL string) {

	modInitOnce.Do(func() {

		// HACK: rain tree will detect if trying to send to addr=self and not send it
		var err error
		clientPrivateKey, err := pocketCrypto.GeneratePrivateKey()
		if err != nil {
			log.Fatalf(err.Error())
		}

		config, genesis := test_artifacts.ReadConfigAndGenesisFiles(configPath, genesisPath)
		config.Base.PrivateKey = clientPrivateKey.String()

		consensusMod, err = consensus.Create(config, genesis)
		if err != nil {
			log.Fatalf("[ERROR] Failed to create consensus module: %v", err.Error())
		}

		p2pMod, err = p2p.Create(config, genesis)
		if err != nil {
			log.Fatalf("[ERROR] Failed to create p2p module: %v", err.Error())
		}

		// This telemetry module instance is a NOOP because the 'enable_telemetry' flag in the `cfg` above is set to false.
		// Since this client mimics partial - networking only - functionality of a full node, some of the telemetry-related
		// code paths are executed. To avoid those messages interfering with the telemetry data collected, a non-nil telemetry
		// module that NOOPs (per the configs above) is injected.
		telemetryMod, err := telemetry.Create(config, genesis)
		if err != nil {
			log.Fatalf("[ERROR] Failed to create NOOP telemetry module: " + err.Error())
		}

		_ = shared.CreateBusWithOptionalModules(nil, p2pMod, nil, consensusMod, telemetryMod, config, genesis)

		p2pMod.Start()
	})

}
