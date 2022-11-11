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

// TODO: modify module structure to prevent crypto/ copy redundancy

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/go-bip39"
	"github.com/spf13/cobra"
	"github.com/syndtr/goleveldb/leveldb"
	cryptoPocket "keys/crypto"
)

const (
	flagRecover = "recover"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Creating an encrypted private key and save to <name> file as the key pair identifier",
	Long: `Derive a new private key and encrypt to disk.

Allow users to use BIP39 mnemonic and to secure the mnemonic. Take key ID <name> from user and store key under the <name>.
`,
	Args: cobra.ExactArgs(1),
	RunE: runAddCmd,
}

/*
key type
	- name: the unique ID for the key
	- publickey: the public key
	- privatekey: the private key
	- mnemonic: mnemonic used to generated key, empty if not saved
*/
type key struct {
	Name       string                  `json:"name"`
	PublicKey  cryptoPocket.PublicKey  `json:"publickey"`
	PrivateKey cryptoPocket.PrivateKey `json:"privatekey"`
	Address    cryptoPocket.Address    `json:"address"`
	Mnemonic   string                  `json:"mnemonic"`
}

/*
[Miniature Keybase] Using level.db and ED25519 for public/private keys generation
input
  - bip39 mnemonic
  - bip39 passphrase
  - bip44 path
  - local encryption password

output
  - armor encrypted private key (saved to file)
*/
func runAddCmd(cmd *cobra.Command, args []string) error {
	var err error
	var inBuf = bufio.NewReader(cmd.InOrStdin())

	name := args[0]

	//////////////////////////
	//  Mnemonic Generation //
	//////////////////////////

	// Get bip39 mnemonic
	var mnemonic string
	var bip39Passphrase string = "" // TODO: implement to take pass phrases from user

	// User can recover private key from mnemonic
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

	/////////////////////
	// Keys Generation //
	/////////////////////

	// TODO: determine a safer keystore location (team discuss)
	kb, err := leveldb.OpenFile("./.keybase/poktKeys.db", nil)

	// TODO: is this generating order with the provided ED25519 correct? (team discuss)

	// Creating a private key with ED25519 and mnemonic seed phrases
	privateKey, err := cryptoPocket.NewPrivateKeyFromSeed([]byte(mnemonic + bip39Passphrase))
	if err != nil {
		return err
	}

	// Creating a public key
	publicKey := privateKey.PublicKey()

	// Creating an address
	address := privateKey.Address()

	// JSON encoding
	var keystore = key{name, publicKey, privateKey, address, mnemonic}

	//////////////////
	// Storing keys //
	//////////////////

	data, err := json.Marshal(keystore)
	if err != nil {
		return err
	}
	err = kb.Put([]byte(name), data, nil)
	if err != nil {
		return err
	}

	//////////////////
	// Print Output //
	//////////////////

	// Print out indented JSON
	output, err := json.MarshalIndent(keystore, "", "\t")
	fmt.Printf("%v\n", output)

	defer kb.Close()

	return nil
}

func init() {
	rootCmd.AddCommand(createCmd)

	// Local Flags
	f := createCmd.Flags()
	f.Bool(flagRecover, false, "Provide seed phrase to recover existing key instead of creating")
}
