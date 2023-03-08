package p2p

import (
	"context"
	"encoding/csv"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	rpcABP "github.com/pokt-network/pocket/p2p/providers/addrbook_provider/rpc"
	rpcCHP "github.com/pokt-network/pocket/p2p/providers/current_height_provider/rpc"
	"github.com/pokt-network/pocket/rpc"
	"github.com/pokt-network/pocket/runtime/defaults"
	sharedP2P "github.com/pokt-network/pocket/shared/p2p"
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
	var pstore sharedP2P.Peerstore

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

		addressBookProvider := rpcABP.NewRPCAddrBookProvider(
			rpcABP.WithP2PConfig(
				m.GetBus().GetRuntimeMgr().GetConfig().P2P,
			),
			rpcABP.WithCustomRPCURL(bootstrapNode),
		)

		currentHeightProvider := rpcCHP.NewRPCCurrentHeightProvider(rpcCHP.WithCustomRPCURL(bootstrapNode))

		pstore, err = addressBookProvider.GetStakedAddrBookAtHeight(currentHeightProvider.CurrentHeight())
		if err != nil {
			m.logger.Warn().Err(err).Str("endpoint", bootstrapNode).Msg("Error getting address book from bootstrap node")
			continue
		}
	}

	if pstore.Size() == 0 {
		return fmt.Errorf("bootstrap failed")
	}

	for _, peer := range pstore.GetAllPeers() {
		m.logger.Debug().Str("address", peer.GetAddress().String()).Msg("Adding peer to pstore")
		// TECHDEBT: either remove the returned error from the interface OR log the error here.
		if err := m.network.AddPeer(peer); err != nil {
			return err
		}
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
