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
	"errors"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/go-bip39"
	"github.com/spf13/cobra"
	"github.com/syndtr/goleveldb/leveldb"
)

const (
	flagRecover = "recover"

	// FlagPublicKey represents the user's public key on the command line.
	FlagPublicKey = "pubkey"

	//// For output formats
	//OutputFormatText = "text"
	//OutputFormatJSON = "json"

	// DefaultKeyPass contains the default key password for genesis transactions
	DefaultKeyPass = "12345678"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Creating an encrypted private key and save to <name> file as the key pair identifier",
	Long: `Derive a new private key and encrypt to disk.

Allow users to use BIP39 mnemonic and set passphrases to secure the mnemonic. Supports BIP32 Hierarchical
Deterministic (HD) path to derive a specific account. Take key ID <name> from user and store key under the <name>.
Key file is encrypted with the given password (required).
	
	--dry-run	Generate/Recover a key without stored it to the local keystore.

	-i			Prompt the user for BIP44 path, BIP39 mnemonic, and passphrase.

	--recover 	Recover a key from a seed passphrase.
`,
	Args: cobra.ExactArgs(1),
	RunE: runAddCmdPrepare,
}

func runAddCmdPrepare(cmd *cobra.Command, args []string) error {
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}

	buf := bufio.NewReader(clientCtx.Input)
	return runAddCmd(clientCtx, cmd, args, buf)
}

/*
input
  - bip39 mnemonic
  - bip39 passphrase
  - bip44 path
  - local encryption password

output
  - armor encrypted private key (saved to file)
*/
func runAddCmd(ctx client.Context, cmd *cobra.Command, args []string, inBuf *bufio.Reader) error {
	var err error

	// [Miniature Keybase] Using level.db for keybase and ED25519 for public/private key generation

	name := args[0]

	// TODO: determine proper keystore location later
	kb, err := leveldb.OpenFile("./.keybase/poktKeys.db", nil)

	// Creating key with ED25519
	pubKey, privKey, err := ed25519.GenerateKey(nil) // TODO: using mnemonic to generate (read)
	keyMap := make(map[string][]byte)
	keyMap["pubKey"] = pubKey
	keyMap["privKey"] = privKey

	// TODO: bette print function (support Test and JSON)
	fmt.Printf("Name (KEY UID): %s\n", name)
	fmt.Printf("Public key: %v\n", pubKey)
	fmt.Printf("Private key: %v\n", privKey)

	// Store the raw key to keybase (TODO: check if key already exists; user input sanitize)

	// TODO: store JSON formated name -> {pubkey, privkey, ...}
	err = kb.Put([]byte(name), privKey, nil)
	if err != nil {
		return err
	}

	// TODO: mnemonic key generation and recovery feature

	// Get bip39 mnemonic
	var mnemonic string
	var bip39Passphrase string = ""

	recover, _ := cmd.Flags().GetBool(flagRecover)
	if recover {
		mnemonic, err = input.GetString("Enter your bip39 mnemonic", inBuf)
		if err != nil {
			return err
		}

		if !bip39.IsMnemonicValid(mnemonic) {
			return errors.New("invalid mnemonic")
		}
	} else {
		// read entropy seed straight from tmcrypto.Rand and convert to mnemonic
		entropySeed, err := bip39.NewEntropy(mnemonicEntropySize)
		if err != nil {
			return err
		}

		mnemonic, err = bip39.NewMnemonic(entropySeed)
		if err != nil {
			return err
		}
	}

	// TODO: Using Mnemonic to generate keys

	// TODO: Import the CryptoPocket!
	ed25519.NewKeyFromSeed()

	defer kb.Close()

	return nil
}

func init() {
	rootCmd.AddCommand(createCmd)

	// Local Flags
	f := createCmd.Flags()
	f.Bool(flagRecover, false, "Provide seed phrase to recover existing key instead of creating")
}
