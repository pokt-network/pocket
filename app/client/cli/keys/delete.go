package keys

import (
	"bufio"
	"log"

	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/spf13/cobra"
	"github.com/syndtr/goleveldb/leveldb"
)

const (
	flagYes = "yes"
)

// DeleteCmd represents the delete command
var DeleteCmd = &cobra.Command{
	Use:   "delete <name> ...",
	Short: "Delete the given key from the keystore",
	Long: `Delete the public key from the backend keystore offline
Note: Delete key does not delete private key stored in a ledger device.
`,
	Args: cobra.ExactArgs(1),
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
	var kb *leveldb.DB
	if kb, err = leveldb.OpenFile("./.keybase/poktKeys.db", nil); err != nil {
		return err
	}

	defer kb.Close()

	// Check if the key name is in the DB
	if _, err = kb.Get([]byte(name), nil); err != nil {
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
	if err = kb.Delete([]byte(name), nil); err != nil {
		return err
	}

	log.Printf("Key (%s) deleted!\n", name)

	return nil
}
