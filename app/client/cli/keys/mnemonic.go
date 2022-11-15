package keys

import (
	"bufio"
	"crypto/sha256"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/go-bip39"
	"github.com/spf13/cobra"
)

const (
	flagUserEntropy = "unsafe-entropy"

	mnemonicEntropySize = 256
)

// MnemonicCmd represents the mnemonic command
var MnemonicCmd = &cobra.Command{
	Use:   "mnemonic",
	Short: "Computing BIP-39 mnemonic phrases",
	Long:  `Computing and output seed phrases based on BIP-39 and system entropy. Passing your own entropy, use --unsafe-entropy`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// entropy source
		var entropySeed []byte

		// get entropy from users if --unsafe-entropy flag is passed
		var userEntropy bool
		var err error
		if userEntropy, err = cmd.Flags().GetBool(flagUserEntropy); err != nil {
			return err
		}

		if userEntropy {
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
	// Local Flags
	MnemonicCmd.Flags().BoolP(flagUserEntropy, "u", false, "Prompt the user to supply their own entropy, instead of relying on the system")
}
