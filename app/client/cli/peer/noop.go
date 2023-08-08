//go:build !debug

package peer

import (
	"github.com/spf13/cobra"
)

func NewPeerCommand() *cobra.Command {
	return nil
}
