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
	rpcABP "github.com/pokt-network/pocket/p2p/providers/peerstore_provider/rpc"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/rpc"
	"github.com/pokt-network/pocket/runtime/defaults"
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
	var pstore typesP2P.Peerstore

	for _, bootstrapNode := range m.bootstrapNodes {
		m.logger.Info().Str("endpoint", bootstrapNode).Msg("Attempting to bootstrap from bootstrap node")

		client, err := rpc.NewClientWithResponses(bootstrapNode)
		if err != nil {
			continue
		}
		healthCheck, err := client.GetV1Health(context.TODO())
		if err != nil || healthCheck == nil || healthCheck.StatusCode != http.StatusOK {
			m.logger.Warn().Str("bootstrapNode", bootstrapNode).Msg("Error getting a green health check from bootstrap node")
			continue
		}

		pstoreProvider := rpcABP.NewRPCPeerstoreProvider(
			rpcABP.WithP2PConfig(
				m.GetBus().GetRuntimeMgr().GetConfig().P2P,
			),
			rpcABP.WithCustomRPCURL(bootstrapNode),
		)

		currentHeightProvider := rpcCHP.NewRPCCurrentHeightProvider(rpcCHP.WithCustomRPCURL(bootstrapNode))

		pstore, err = pstoreProvider.GetStakedPeerstoreAtHeight(currentHeightProvider.CurrentHeight())
		if err != nil {
			m.logger.Warn().Err(err).Str("endpoint", bootstrapNode).Msg("Error getting address book from bootstrap node")
			continue
		}
	}

	for _, peer := range pstore.GetPeerList() {
		m.logger.Debug().Str("address", peer.GetAddress().String()).Msg("Adding peer to router")
		if err := m.router.AddPeer(peer); err != nil {
			m.logger.Error().Err(err).
				Str("pokt_address", peer.GetAddress().String()).
				Msg("adding peer")
		}
	}

	if m.router.GetPeerstore().Size() == 0 {
		return fmt.Errorf("bootstrap failed")
	}
	return nil
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
