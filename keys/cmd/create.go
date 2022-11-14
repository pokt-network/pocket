package cmd

// TODO: modify module structure to prevent crypto/ copy redundancy

import (
	"bufio"
	"encoding/json"
	"errors"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/go-bip39"
	"github.com/spf13/cobra"
	"github.com/syndtr/goleveldb/leveldb"
	cryptoPocket "keys/crypto"
	"log"
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
	- Address: the address related to the private key
	- mnemonic: mnemonic used to generated key, empty if not saved
*/
type key struct {
	Name       string                  `json:"name"`
	PublicKey  cryptoPocket.PublicKey  `json:"publickey"`
	PrivateKey cryptoPocket.PrivateKey `json:"privatekey"`
	Address    cryptoPocket.Address    `json:"address"`
	Mnemonic   string                  `json:"mnemonic"`
}

// Future updates
// - determine a safer keystore location (team discuss)
// - confirmation from user for overriding existing key
// - implement key phrase intput from user secure keys
func runAddCmd(cmd *cobra.Command, args []string) error {
	var inBuf = bufio.NewReader(cmd.InOrStdin())

	name := args[0]

	//////////////////////////
	//  Mnemonic Generation //
	//////////////////////////

	// Get bip39 mnemonic
	var mnemonic string
	var bip39Passphrase string = ""

	// User can recover private key from mnemonic
	recover, err := cmd.Flags().GetBool(flagRecover)
	if err != nil {
		return err
	}

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

	var kb *leveldb.DB
	if kb, err = leveldb.OpenFile("/.keybase/poktKeys.db", nil); err != nil {
		return err
	}

	defer kb.Close() // execute at the conclusion of the function

	// Creating a private key with ED25519 and mnemonic seed phrases
	privateKey, err := cryptoPocket.NewPrivateKeyFromSeed([]byte(mnemonic + bip39Passphrase))
	if err != nil {
		return err
	}

	keystore := key{name, privateKey.PublicKey(), privateKey, privateKey.Address(), mnemonic}

	//////////////////
	// Storing keys //
	//////////////////

	// TODO: ask users for passphrase for key protection
	var data []byte
	if data, err = json.Marshal(keystore); err != nil {
		return err
	}

	if err = kb.Put([]byte(name), data, nil); err != nil {
		return err
	}

	///////////////
	// Print Key //
	///////////////
	if err = printKey(keystore); err != nil {
		return err
	}

	return nil
}

func init() {
	rootCmd.AddCommand(createCmd)

	// Local Flags
	f := createCmd.Flags()
	f.Bool(flagRecover, false, "Provide seed phrase to recover existing key instead of creating")
}

// Utility functions

// Print out key in indented JSON format
func printKey(keystore key) error {
	output, err := json.MarshalIndent(keystore, "", "\t")
	if err != nil {
		return err
	}
	log.Printf("%s\n", output)

	return nil
}
