package p2p

import (
	"context"
	"encoding/csv"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	rpcCHP "github.com/pokt-network/pocket/p2p/providers/current_height_provider/rpc"
	rpcPSP "github.com/pokt-network/pocket/p2p/providers/peerstore_provider/rpc"
	"github.com/pokt-network/pocket/rpc"
	"github.com/pokt-network/pocket/runtime/defaults"
	"github.com/pokt-network/pocket/shared/utils"
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
	limiter := utils.NewLimiter(int(m.cfg.MaxBootstrapConcurrency))

	for _, serviceURL := range m.bootstrapNodes {
		// concurrent bootstrapping
		// TECHDEBT(#595): add ctx to interface methods and propagate down.
		limiter.Go(context.TODO(), func() {
			m.bootstrapFromRPC(strings.Clone(serviceURL))
		})
	}

	limiter.Close()

	return nil
}

// bootstrapFromRPC fetches the peerstore of the peer at `serviceURL` via RPC
// and adds it to this host's peerstore after performing a health check.
// TECHDEBT(SOLID): refactor; this method has more than one reason to change
func (m *p2pModule) bootstrapFromRPC(serviceURL string) {
	m.logger.Info().Str("endpoint", serviceURL).Msg("Attempting to bootstrap from bootstrap node")

	client, err := rpc.NewClientWithResponses(serviceURL)
	if err != nil {
		return
	}
	healthCheck, err := client.GetV1Health(context.TODO())
	if err != nil || healthCheck == nil || healthCheck.StatusCode != http.StatusOK {
		m.logger.Warn().Str("serviceURL", serviceURL).Msg("Error getting a green health check from bootstrap node")
		return
	}

	// fetch `serviceURL`'s  peerstore
	pstoreProvider := rpcPSP.NewRPCPeerstoreProvider(
		rpcPSP.WithP2PConfig(
			m.GetBus().GetRuntimeMgr().GetConfig().P2P,
		),
		rpcPSP.WithCustomRPCURL(serviceURL),
	)

	currentHeightProvider := rpcCHP.NewRPCCurrentHeightProvider(rpcCHP.WithCustomRPCURL(serviceURL))

	pstore, err := pstoreProvider.GetStakedPeerstoreAtHeight(currentHeightProvider.CurrentHeight())
	if err != nil {
		m.logger.Warn().Err(err).Str("endpoint", serviceURL).Msg("Error getting address book from bootstrap node")
		return
	}

	// add `serviceURL`'s peers to this node's peerstore
	for _, peer := range pstore.GetPeerList() {
		m.logger.Debug().Str("address", peer.GetAddress().String()).Msg("Adding peer to network")
		if err := m.network.AddPeer(peer); err != nil {
			m.logger.Error().Err(err).
				Str("pokt_address", peer.GetAddress().String()).
				Msg("adding peer")
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
