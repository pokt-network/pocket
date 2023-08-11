//go:build debug

package p2p

import (
	"fmt"
	"os"

	"github.com/pokt-network/pocket/shared/modules"
)

type RouterType string

const (
	StakedRouterType   RouterType = "staked"
	UnstakedRouterType RouterType = "unstaked"
	AllRouterTypes     RouterType = "all"
	Libp2pHost         RouterType = "libp2p_host"
)

func LogSelfAddress(bus modules.Bus) error {
	p2pModule := bus.GetP2PModule()
	if p2pModule == nil {
		return fmt.Errorf("no p2p module found on the bus")
	}

	selfAddr, err := p2pModule.GetAddress()
	if err != nil {
		return fmt.Errorf("getting self address: %w", err)
	}

	_, err = fmt.Fprintf(os.Stdout, "self address: %s\n", selfAddr.String())
	return err
}
