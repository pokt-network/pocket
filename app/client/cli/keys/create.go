package keys

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"log"
	"os"
	"path"

	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/go-bip39"
	"github.com/spf13/cobra"
	"github.com/syndtr/goleveldb/leveldb"
)

const (
	flagInteractive = "interactive"
	flagRecover     = "recover"

	mnemonicEntropySize = 256
)

// CreateCmd represents the create command
var CreateCmd = &cobra.Command{
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
 - privatekey: the private key (encrypted and ignored in JSON)
 - Address: the address related to the private key
 - mnemonic: mnemonic used to recover private key (ignored in JSON)
*/
type key struct {
	Name       string                 `json:"name"`
	PublicKey  cryptoPocket.PublicKey `json:"publickey"`
	PrivateKey string                 `json:"-"`
	Address    cryptoPocket.Address   `json:"address"`
	Mnemonic   string                 `json:"-"`
}

func runAddCmd(cmd *cobra.Command, args []string) error {
	var inBuf = bufio.NewReader(cmd.InOrStdin())
	var err error

	name := args[0]
	var interactive, recover bool
	if interactive, err = cmd.Flags().GetBool(flagInteractive); err != nil {
		return err
	}
	if recover, err = cmd.Flags().GetBool(flagRecover); err != nil {
		return err
	}

	////////////////////
	//  Keybase Setup //
	////////////////////

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

	// Ask user to override existing key if exists
	if !recover {
		if err = overrideKey(kb, name, inBuf, cmd); err != nil {
			return err
		}
	}

	//////////////////////////
	//  Mnemonic Generation //
	//////////////////////////

	// Get bip39 mnemonic
	var mnemonic, bip39Passphrase string

	if recover {
		if mnemonic, err = input.GetString("Enter your bip39 mnemonic", inBuf); err != nil {
			return err
		}

		if !bip39.IsMnemonicValid(mnemonic) {
			return errors.New("invalid mnemonic")
		}
	} else if interactive {
		if mnemonic, err = input.GetString("Enter your bip39 mnemonic, or hit enter to generate one.", inBuf); err != nil {
			return err
		}

		if !bip39.IsMnemonicValid(mnemonic) && mnemonic != "" {
			return errors.New("invalid mnemonic")
		}
	}

	// generate new mnemonic when user doesn't have one
	if len(mnemonic) == 0 {
		// read entropy seed straight from tmcrypto.Rand and convert to mnemonic
		var entropySeed []byte
		if entropySeed, err = bip39.NewEntropy(mnemonicEntropySize); err != nil {
			return err
		}

		if mnemonic, err = bip39.NewMnemonic(entropySeed); err != nil {
			return err
		}
	}

	/////////////////////
	// Keys Generation //
	/////////////////////

	// Interactive key generation (override empty bip39 passphrase with user inputs)
	if interactive {
		if bip39Passphrase, err = input.GetString(
			"Enter your bip39 passphrase. This is combined with the mnemonic to derive the seed. "+
				"(default \"\")", inBuf); err != nil {
			return err
		}

		// if they use one, make them re-enter it
		if len(bip39Passphrase) != 0 {
			var passphraseRepeat string
			if passphraseRepeat, err = input.GetString("Repeat the passphrase:", inBuf); err != nil {
				return err
			}

			if bip39Passphrase != passphraseRepeat {
				return errors.New("passphrases don't match")
			}
		}
	}

	// Creating a private key with ED25519 and mnemonic seed phrases
	var privateKey cryptoPocket.PrivateKey
	if privateKey, err = cryptoPocket.NewPrivateKeyFromSeed([]byte(mnemonic + bip39Passphrase)); err != nil {
		return err
	}

	var encryptedKey, userKey string

	// create 32 bytes long key for AES encryption based on user passphrase
	if userKey, err = generateKeyFromPassPhrase(bip39Passphrase); err != nil {
		return err
	}

	// encrypt private key with AES (use decrypt func from utility to decrypt)
	if encryptedKey, err = encrypt(privateKey.String(), userKey); err != nil {
		return err
	}

	// store encrypted private key and mnemonic not stored by default
	keystore := key{
		name,
		privateKey.PublicKey(),
		encryptedKey,
		privateKey.Address(),
		"",
	}

	//////////////////
	// Storing keys //
	//////////////////

	var data []byte
	if data, err = json.Marshal(keystore); err != nil {
		return err
	}

	if err = kb.Put([]byte(name), data, nil); err != nil {
		return err
	}

	//////////////
	// Log Keys //
	//////////////
	if err = logInfo(keystore, mnemonic, recover); err != nil {
		return err
	}

	return nil
}

// Check if key name already exists and ask user to override or not
func overrideKey(kb *leveldb.DB, name string, inBuf *bufio.Reader, cmd *cobra.Command) error {
	if _, err := kb.Get([]byte(name), nil); err == nil {
		log.Printf("Key \"%s\" alredy exists", name)

		// account exists, ask for user confirmation
		var response bool
		var err2 error
		if response, err2 = input.GetConfirmation(
			fmt.Sprintf("override the existing name %s", name), inBuf, cmd.ErrOrStderr()); err2 != nil {
			return err2
		}

		if !response {
			return errors.New("aborted")
		}

		if err2 = kb.Delete([]byte(name), nil); err2 != nil {
			return err2
		}
	}

	return nil
}

func init() {

	// Local Flags
	f := CreateCmd.Flags()
	f.Bool(flagRecover, false, "Provide seed phrase to recover existing key instead of creating")
	f.BoolP(flagInteractive, "i", false, "Interactively prompt user for BIP39 passphrase and mnemonic")
}
