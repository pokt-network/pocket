package cmd

import (
	"bufio"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/spf13/cobra"
	"github.com/syndtr/goleveldb/leveldb"
)

const (
	flagYes = "yes"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete <name> ...",
	Short: "Delete the given key from the keystore",
	Long: `Delete the public key from the backend keystore offline
Note: Delete key does not delete private key stored in a ledger device.
`,
	Args: cobra.MinimumNArgs(1),
	RunE: deleteKey,
}

// Future updates
// - Check if the file exist or not
// - Delete multiple keys
// - Skip confirmation prompt when deleting offline or ledger key references
func deleteKey(cmd *cobra.Command, args []string) error {
	var err error

	name := args[0]

	// Open keybase DB
	kb, err := leveldb.OpenFile("./.keybase/poktKeys.db", nil)
	if err != nil {
		return err
	}

	// Check if the key name is in the DB
	_, err = kb.Get([]byte(name), nil)
	if err != nil {
		return err
	}

	// confirm error
	buf := bufio.NewReader(cmd.InOrStdin())
	if yes, err := input.GetConfirmation("Key reference will be deleted. Continue?", buf, cmd.ErrOrStderr()); err != nil {
		return err
	} else if !yes {
		return nil
	}

	// remove key based on name KEY ID
	err = kb.Delete([]byte(name), nil)
	if err != nil {
		return err
	}

	fmt.Printf("Key (%s) deleted!\n", name)

	defer kb.Close()

	return nil
}

func init() {
	rootCmd.AddCommand(deleteCmd)

	// Local flags
}
