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

var (
	Config       *config.Config
	P2pMod       modules.P2PModule
	ConsensusMod modules.ConsensusModule

	modInitOnce sync.Once
)

func Init(remoteCLIURL string) {

	modInitOnce.Do(func() {

		pk, err := crypto.GeneratePrivateKey()
		if err != nil {
			log.Fatalf("[ERROR] Failed to generate private key: %v", err)
		}

		Config := &config.Config{
			GenesisSource: &genesis.GenesisSource{
				Source: &genesis.GenesisSource_File{
					File: &genesis.GenesisFile{
						Path: "build/config/genesis.json",
					},
				},
			},

			// Not used - only set to avoid `GetNodeState(_)` from crashing
			PrivateKey: pk.(crypto.Ed25519PrivateKey),

			// Used to access the validator map
			Consensus: &config.ConsensusConfig{
				Pacemaker: &config.PacemakerConfig{},
			},

			// Not used - only set to avoid `p2p.Create()` from crashing
			P2P: &config.P2PConfig{
				ConsensusPort:  9999,
				UseRainTree:    true,
				ConnectionType: config.TCPConnection,
			},
			EnableTelemetry: false,

			RemoteCLIURL: remoteCLIURL,
		}
		if err := Config.HydrateGenesisState(); err != nil {
			log.Fatalf("[ERROR] Failed to hydrate the genesis state: %v", err.Error())
		}

		ConsensusMod, err = consensus.Create(Config)
		if err != nil {
			log.Fatalf("[ERROR] Failed to create consensus module: %v", err.Error())
		}

		P2pMod, err = p2p.Create(Config)
		if err != nil {
			log.Fatalf("[ERROR] Failed to create p2p module: %v", err.Error())
		}
		// This telemetry module instance is a NOOP because the 'enable_telemetry' flag in the `cfg` above is set to false.
		// Since this client mimics partial - networking only - functionality of a full node, some of the telemetry-related
		// code paths are executed. To avoid those messages interfering with the telemetry data collected, a non-nil telemetry
		// module that NOOPs (per the configs above) is injected.
		telemetryMod, err := telemetry.Create(Config)
		if err != nil {
			log.Fatalf("[ERROR] Failed to create NOOP telemetry module: " + err.Error())
		}

		_ = shared.CreateBusWithOptionalModules(nil, P2pMod, nil, ConsensusMod, telemetryMod, Config)

		P2pMod.Start()
	})

}
