/*
Copyright © 2022 Jason You <jason.you1995@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
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
	Use:   "delete <key> ...",
	Short: "Delete the given key from the keystore",
	Long: `Delete the public key from the backend keystore offline
Note: Delete key does not delete private key stored in a ledger device.
`,
	Args: cobra.MinimumNArgs(1),
	RunE: deleteKey,
}

func deleteKey(cmd *cobra.Command, args []string) error {
	var err error

	// TODO: check if the file exist or not

	// TODO: delete multiple keys
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

	fmt.Println("Key deleted!")

	defer kb.Close()

	return nil
}

func init() {
	rootCmd.AddCommand(deleteCmd)

	// Local flags
	//f := deleteCmd.Flags()
	//f.BoolP(flagYes, "y", false, "Skip confirmation prompt when deleting offline or ledger key references")
}
