package keys

import (
	"bufio"
	"log"
	"os"
	"path"

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

	// Open local key database
	var kb *leveldb.DB
	var kbPath string // local keystore path
	if kbPath, err = os.UserHomeDir(); err != nil {
		panic(err)
	}
	kbPath = path.Join(kbPath, ".keybase", "poktKeys.db")
	log.Printf("Keys stored in local path: %s\n", kbPath)

	if kb, err = leveldb.OpenFile(kbPath, nil); err != nil {
		return err
	}

	defer kb.Close() // execute at the conclusion of the function

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
