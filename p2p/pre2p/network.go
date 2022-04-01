package pre2p

import (
	"fmt"
	"log"
	"net"

	pre2ptypes "github.com/pokt-network/pocket/p2p/pre2p/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types"
)

const (
	NetworkProtocol = "tcp4"
)

var _ pre2ptypes.Network = &network{}

type network struct {
	// TODO(team): This address book is currently static and does not update dynamically as new peers come on/offline.
	// TODO(olshansky): Make sure that self (the current node) is not added to the list to avoid self-broadcasts.
	AddrBook []*pre2ptypes.NetworkPeer
}

func ConnectToValidatorNetwork(validators types.ValMap) (n pre2ptypes.Network) {
	n = &network{}
	for _, v := range validators {
		err := n.(*network).connectToValidator(v)
		if err != nil {
			log.Println("[WARN] Error connecting to validator: ", err)
			continue
		}
	}
	return
}

// TODO(olshansky): How do we avoid self-broadcasts given that `AddrBook` may contain self in the current pre2p implementation?
func (n *network) NetworkBroadcast(data []byte) error {
	for _, peer := range n.AddrBook {
		client, err := net.DialTCP(NetworkProtocol, nil, peer.ConsensusAddr)
		if err != nil {
			log.Println("Error connecting to one of the peers during broadcast: ", err)
			continue
		}
		defer client.Close()

		_, err = client.Write(data)
		if err != nil {
			log.Println("Error writing to one of the peers during broadcast: ", err)
			continue
		}
	}
	return nil
}

func (n *network) NetworkSend(data []byte, address cryptoPocket.Address) error {
	for _, peer := range n.AddrBook {
		// TODO(team): If the address book is a map instead of a list, we wouldn't have to do this loop.
		if address.String() != peer.PublicKey.Address().String() {
			continue
		}

		client, err := net.DialTCP(NetworkProtocol, nil, peer.ConsensusAddr)
		if err != nil {
			log.Println("Error connecting to peer during send: ", err)
			return err
		}
		defer client.Close()

		_, err = client.Write(data)
		if err != nil {
			log.Println("Error writing to peer during send: ", err)
			return err
		}

		break // During a send, only one peer needs to receive the message
	}

	return nil
}

// TODO(hack): Publically exposed for testing purposes only.
func (n *network) GetAddrBook() []*pre2ptypes.NetworkPeer {
	return n.AddrBook
}

func (n *network) connectToValidator(v *types.Validator) error {
	tcpAddr, err := net.ResolveTCPAddr(NetworkProtocol, fmt.Sprintf("%s:%d", v.Host, v.Port))
	if err != nil {
		return fmt.Errorf("error resolving addr: %v", err)
	}

	peer := &pre2ptypes.NetworkPeer{
		ConsensusAddr: tcpAddr,
		PublicKey:     v.PublicKey,
	}

	n.AddrBook = append(n.AddrBook, peer)
	return nil
}
