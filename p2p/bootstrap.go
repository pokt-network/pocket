package p2p

import (
	"context"
	"encoding/csv"
	"fmt"
	rpcPSP "github.com/pokt-network/pocket/p2p/providers/peerstore_provider/rpc"
	"regexp"
	"strconv"
	"strings"

	"github.com/pokt-network/pocket/p2p/providers/peerstore_provider"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/p2p/utils"
	"github.com/pokt-network/pocket/runtime/defaults"
	sharedUtils "github.com/pokt-network/pocket/shared/utils"
)

// configureBootstrapNodes parses the bootstrap nodes from the config and validates them
func (m *p2pModule) configureBootstrapNodes() error {
	p2pCfg := m.GetBus().GetRuntimeMgr().GetConfig().P2P

	bootstrapNodesCsv := strings.Trim(p2pCfg.BootstrapNodesCsv, " ")
	if bootstrapNodesCsv == "" {
		bootstrapNodesCsv = defaults.DefaultP2PBootstrapNodesCsv
	}
	csvReader := csv.NewReader(strings.NewReader(bootstrapNodesCsv))
	bootStrapNodes, err := csvReader.Read()
	if err != nil {
		return fmt.Errorf("error parsing bootstrap nodes: %w", err)
	}

	// validate the bootstrap nodes
	for i, node := range bootStrapNodes {
		bootStrapNodes[i] = strings.Trim(node, " ")
		if !isValidHostnamePort(bootStrapNodes[i]) {
			return fmt.Errorf("invalid bootstrap node: %s", bootStrapNodes[i])
		}
	}
	m.bootstrapNodes = bootStrapNodes
	return nil
}

// bootstrap attempts to bootstrap from a bootstrap node
func (m *p2pModule) bootstrap() error {
	// CONSIDERATION: add a config value to indicate whether this node is intended
	// to be a bootstrap node. It would assume the `persistencePeerstoreProvider`
	// is accurate and would not attempt to perform health checks, use an
	// `rpcPeerstoreProvider`, nor attempt to connect

	if err := m.bootstrapFromRPC(); err != nil {
		m.logger.Error().
			Err(err).
			Msg("error bootstrap via RPC")
	}

	return nil
}

func (m *p2pModule) bootstrapFromRPC() error {
	// TECHDEBT(#595): add ctx to interface methods and propagate down.
	ctx := context.TODO()

	// TECHDEBT(#811): use `bus.GetPeerstoreProvider()` after peerstore provider
	// is retrievable as a proper submodule
	pstoreProvider, err := peerstore_provider.GetPeerstoreProvider(m.GetBus())
	if err != nil {
		return err
	}

	currentHeightProvider := m.GetBus().GetCurrentHeightProvider()

	// NB: consider `m.cfg.MaxBootstrapConcurrency` number of bootstrap nodes concurrently.
	bootstrapNodeLimiter := sharedUtils.NewLimiter(int(m.cfg.MaxBootstrapConcurrency))
	// NB: check health of and attempt to connect to `m.cfg.MaxBootstrapConcurrency`
	// number of staked peers concurrently.
	peerLimiter := sharedUtils.NewLimiter(int(m.cfg.MaxBootstrapConcurrency))
	for _, serviceURL := range m.bootstrapNodes {
		bootstrapNodeLimiter.Go(ctx, m.bootstrapFunc(ctx, serviceURL, peerLimiter))
	}
	bootstrapNodeLimiter.Close()

	// NB: re-registering previous peerstore provider.
	m.GetBus().RegisterModule(pstoreProvider)

	// NB: re-registering previous current height provider.
	m.GetBus().RegisterModule(currentHeightProvider)

	// TECHDEBT(#859): determine bootstrapping success/error conditions.
	return nil
}

// bootstrapFunc is intended to be run in a goroutine for each configured bootstrap
// node (m.bootstrapNodes), with concurrency limited to `m.cfg.MaxBootstrapConcurrency`.
func (m *p2pModule) bootstrapFunc(
	ctx context.Context,
	bsNodeServiceURL string,
	peerLimiter *sharedUtils.Limiter,
) func() {
	return func() {
		// check health of bootstrap node
		// get staked peers from peerstore provider

		m.logger.Info().
			Str("url", bsNodeServiceURL).
			Msg("attempting to bootstrap from bootstrap node")

		if err := utils.CheckHealth(ctx, bsNodeServiceURL, m.logger); err != nil {
			// NB: errors are logged in `checkHealth()`.
			// abort bootstrap attempt if bootstrap node is unhealthy.
			return
		}

		// TECHDEBT: only use `rpcPeerstoreProvider` if this node is not configured
		// to be a bootstrap node.

		// NB: this overwrites the peerstore provider in the registry each time,
		// finally leaving an `RPCPeerstoreProvider` registered.
		// CONSIDERATION: decoupling sub/module creation from registration would
		// eliminate the need to check for an existing module and to restore it.
		// Alternatively, we could add a distinct "slot" in the registry for the
		// RPC peerstore provider.
		rpcPStoreProvider, err := rpcPSP.Create(
			m.GetBus(),
			rpcPSP.WithCustomRPCURL(bsNodeServiceURL),
		)
		if err != nil {
			m.logger.Error().Err(err).
				Str("url", bsNodeServiceURL).
				Msg("error creating RPC peerstore provider while bootstrapping")
			// abort bootstrap attempt if RPC peerstore provider cannot be created.
			return
		}

		// NB: Only add bootstrap peer if it is present in the peerstore. If
		// the peerstore provider is an `rpcPeerstoreProvider`, this will
		// ensure that the bootstrap node is a part of the network. Otherwise,
		// `persistencePeerstore` should be in use which will ensure that
		// only peers which are staked at the last known height are added
		// to the router.
		pstore, err := rpcPStoreProvider.GetStakedPeerstoreAtCurrentHeight()
		if err != nil {
			rtr.logger.Error().Err(err).Msg("error getting staked peerstore at current height")
		}

		for _, peer := range pstore.GetPeerList() {
			checkHealth := bsNodeServiceURL == peer.GetServiceURL()
			peerLimiter.Go(ctx, bootstrapStakedPeerFunc(ctx, checkHealth, peer))
		}
	}
}

func (m *p2pModule) bootstrapStakedPeerFunc(
	ctx context.Context,
	checkHealth bool,
	peer typesP2P.Peer,
) func() {
	return func() {
		if checkHealth {
			if err := utils.CheckHealth(ctx, peer.GetServiceURL(), m.logger); err != nil {
				//   isStaked, err := m.isStakedActor()
				//   if err != nil { ... }
				//   if isStaked {
				//     add to staked actor router
				//   }
				//   add to unstaked actor router
				//   add to libp2p host
				//   attempt to connect
				//   if err != nil {
				//     attemptCount <= maxAttempts {
				//       attemptCount++
				//       retry
				//     } else {
				//	   continue  // (give up on this staked peer)
				//     }
				//   }
			}
		}

		// NB: as long as `rpcPStoreProvider` is registered, `m.isStakedActor()`
		// will look for this node's address in the staked peerstore returned
		// from the RPC provider (i.e. bootstrap node's staked peerstore).
		isStaked, err := m.isStakedActor()
		if err != nil {
			m.logger.Error().Err(err).Msg("checking if node is staked")
		}

		if isStaked {
			if err := m.stakedActorRouter.Bootstrap(ctx, limiter, m.bootstrapNodes); err != nil {
				m.logger.Error().Err(err).Msg("error bootstrapping stake actor router")
			}
		}

		if err := m.unstakedActorRouter.Bootstrap(ctx, limiter, m.bootstrapNodes); err != nil {
			m.logger.Error().Err(err).Msg("error bootstrapping unstaked actor router")
		}
	}
}

func isValidHostnamePort(str string) bool {
	pattern := regexp.MustCompile(`^(https?)://([a-zA-Z0-9.-]+):(\d{1,5})$`)
	matches := pattern.FindStringSubmatch(str)
	if len(matches) != 4 {
		return false
	}
	protocol := matches[1]
	if protocol != "http" && protocol != "https" {
		return false
	}
	port, err := strconv.Atoi(matches[3])
	if err != nil || port < 0 || port > 65535 {
		return false
	}
	return true
}
