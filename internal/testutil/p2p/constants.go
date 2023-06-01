package p2p_testutil

import (
	"fmt"

	"github.com/pokt-network/pocket/runtime/defaults"
)

var (
	// IP4ServiceURL is a string representing a valid IPv4 based ServiceURL using the loopback interface.
	IP4ServiceURL = fmt.Sprintf("127.0.0.1:%d", defaults.DefaultP2PPort)
	// IP6ServiceURL is a string representing a valid IPv6 based ServiceURL.
	// (see: https://www.rfc-editor.org/rfc/rfc3986#section-3.2.2)
	IP6ServiceURL = fmt.Sprintf("[2a00:1450:4005:802::2004]:%d", defaults.DefaultP2PPort)
)
