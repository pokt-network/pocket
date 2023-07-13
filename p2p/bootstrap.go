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
// TECHDEBT(#859): refactor bootstrapping.
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

		pstoreProvider, err := rpcPSP.Create(
			m.GetBus(),
			rpcPSP.WithCustomRPCURL(bootstrapNode),
		)
		if err != nil {
			return fmt.Errorf("creating RPC peerstore provider: %w", err)
		}

		currentHeightProvider, err := rpcCHP.Create(
			m.GetBus(),
			rpcCHP.WithCustomRPCURL(bootstrapNode),
		)
		if err != nil {
			m.logger.Warn().Err(err).Str("endpoint", bootstrapNode).Msg("Error getting current height from bootstrap node")
			continue
		}

		pstore, err = pstoreProvider.GetStakedPeerstoreAtHeight(currentHeightProvider.CurrentHeight())
		if err != nil {
			m.logger.Warn().Err(err).Str("endpoint", bootstrapNode).Msg("Error getting address book from bootstrap node")
			continue
		}

		for _, peer := range pstore.GetPeerList() {
			m.logger.Debug().Str("address", peer.GetAddress().String()).Msg("Adding peer to router")
			isStaked, err := m.isStakedActor()
			if err != nil {
				m.logger.Error().Err(err).Msg("checking if node is staked")
			}
			if isStaked {
				if err := m.stakedActorRouter.AddPeer(peer); err != nil {
					m.logger.Error().Err(err).
						Str("pokt_address", peer.GetAddress().String()).
						Msg("adding peer to staked actor router")
				}
			}

			if err := m.unstakedActorRouter.AddPeer(peer); err != nil {
				m.logger.Error().Err(err).
					Str("pokt_address", peer.GetAddress().String()).
					Msg("adding peer to unstaked actor router")
			}
		}
	}

	// TECHDEBT(#859): determine bootstrapping success/error conditions.
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
