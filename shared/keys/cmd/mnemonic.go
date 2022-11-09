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
	"crypto/sha256"
	"fmt"
	"github.com/cosmos/go-bip39"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/input"
)

const (
	flagUserEntropy = "unsafe-entropy"

	mnemonicEntropySize = 256
)

// mnemonicCmd represents the mnemonic command
var mnemonicCmd = &cobra.Command{
	Use:   "mnemonic",
	Short: "Computing BIP-39 mnemonic phrases",
	Long:  `Computing and output seed phrases based on BIP-39 and system entropy. Passing your own entropy, use --unsafe-entropy`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// entropy source
		var entropySeed []byte

		// get entropy from users if --unsafe-entropy flag is passed
		if userEntropy, _ := cmd.Flags().GetBool(flagUserEntropy); userEntropy {
			// user entropy buffer storage
			buf := bufio.NewReader(cmd.InOrStdin())

			// prompt user to enter entropy safely
			inputEntropy, err := input.GetString(
				"WARNING: Generate at least 256-bits of entropy and enter the results here:",
				buf)
			if err != nil {
				return err
			}

			// check entropy input validity
			if len(inputEntropy) < 43 {
				return fmt.Errorf(
					"256-bits is 43 characters in Base-64, and 100 in Base-6. You entered %v, and probably want more",
					len(inputEntropy))
			}

			conf, err := input.GetConfirmation(
				fmt.Sprintf("> Input length: %d", len(inputEntropy)),
				buf,
				cmd.ErrOrStderr())
			if err != nil {
				return err
			}

			// return if user didn't confirm with "yes"
			if !conf {
				return nil
			}

			// hash input entropy to get entropy seed
			hashedEntropy := sha256.Sum256([]byte(inputEntropy))
			entropySeed = hashedEntropy[:] // creating a slice reference to the hashedEntropy array
		} else {
			// getting entropy seed from crypto.Rand
			var err error
			entropySeed, err = bip39.NewEntropy(mnemonicEntropySize)
			if err != nil {
				return err
			}
		}

		mnemonic, err := bip39.NewMnemonic(entropySeed)
		if err != nil {
			return err
		}

		// mnemonic is shown only once to user and raw text never stored
		cmd.Println(mnemonic)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(mnemonicCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// mnemonicCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Local Flags
	mnemonicCmd.Flags().BoolP(flagUserEntropy, "u", false, "Prompt the user to supply their own entropy, instead of relying on the system")
}
