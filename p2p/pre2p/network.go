package pre2p

import (
	"fmt"
	"log"
	"net"
	"pocket/p2p/pre2p/types"
)

var _ types.Network = &network{}

type network struct {
	AddrBook []*types.NetworkPeer
}

func ConnectToValidatorNetwork(validators types.ValMap) (n types.Network) {
	n = &network{}
	for nodeId, v := range validators {
		err := n.(*network).connectToValidator(nodeId, v)
		if err != nil {
			log.Println("[WARN] Error connecting to validator: ", err)
			continue
		}
	}
	return
}

func (n *network) NetworkBroadcast(data []byte, self types.NodeId) error {
	// TODO(team): This address book is currently static and does not update dynamically as new peers come on/offline.
	for _, peer := range n.AddrBook {
		// TODO(team): Discuss how self-broadcasts should be handled. A the moment, the consensus
		// module has custom logic that makes a leader take extra actions to also behave as a replica
		// rather than having it be "generalized" through the P2P layer.
		if self == peer.NodeId {
			continue
		}
		client, err := net.DialTCP("tcp", nil, peer.ConsensusAddr)
		if err != nil {
			log.Println("Error connecting to peer: ", err)
			continue
		}
		client.Write(data)
		client.Close()
	}
	return nil
}

func (n *network) NetworkSend(data []byte, node types.NodeId) error {
	for _, peer := range n.AddrBook {
		// TODO(olshansky): Quick hack to avoid sending network messages to self.
		if node != peer.NodeId {
			continue
		}
		client, err := net.DialTCP("tcp", nil, peer.ConsensusAddr)
		if err != nil {
			log.Println("Error connecting to peer: ", err)
			continue
		}
		client.Write(data)
		client.Close()
		break
	}

	return nil
}

// TODO(hack): Publically exposed for testing purposes only.
func (n *network) GetAddrBook() []*types.NetworkPeer {
	return n.AddrBook
}

func (n *network) connectToValidator(nodeId types.NodeId, v *types.Validator) error {
	var tcpAddr *net.TCPAddr
	var err, errDebug error

	tcpAddr, err = net.ResolveTCPAddr("tcp4", fmt.Sprintf("%s:%d", v.Host, v.Port))
	if err != nil || errDebug != nil {
		return fmt.Errorf("error resolving addr: %v %v", err, errDebug)
	}

	peer := &types.NetworkPeer{
		ConsensusAddr: tcpAddr,
		NodeId:        nodeId,
		PublicKey:     v.PublicKey,
	}

	n.AddrBook = append(n.AddrBook, peer)
	return nil
}
