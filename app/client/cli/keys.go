package cli

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/utils"

	"github.com/spf13/cobra"
)

var (
	outputFile string
	inputFile  string
	exportAs   string
	importAs   string
	hint       string
	newPwd     string
	storeChild bool
	childPwd   string
	childHint  string
)

func init() {
	rootCmd.AddCommand(NewKeysCommand())
}

func NewKeysCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "Keys",
		Short:   "Key specific commands",
		Aliases: []string{"keys"},
		Args:    cobra.ExactArgs(0),
	}

	cmd.AddCommand(keysCreateCommands()...)
	cmd.AddCommand(keysUpdateCommands()...)
	cmd.AddCommand(keysDeleteCommands()...)
	cmd.AddCommand(keysGetCommands()...)
	cmd.AddCommand(keysExportCommands()...)
	cmd.AddCommand(keysImportCommands()...)
	cmd.AddCommand(keysSignMsgCommands()...)
	cmd.AddCommand(keysSignTxCommands()...)
	cmd.AddCommand(keysSlipCommands()...)

	return cmd
}

func keysCreateCommands() []*cobra.Command {
	cmds := []*cobra.Command{
		{
			Use:     "Create",
			Short:   "Create new key",
			Long:    "Creates a new key and stores it in the keybase",
			Aliases: []string{"create"},
			Args:    cobra.ExactArgs(0),
			RunE: func(cmd *cobra.Command, args []string) error {
				kb, err := keybaseForCLI()
				if err != nil {
					return err
				}

				if !nonInteractive {
					pwd = readPassphrase(pwd)
					confirmPassphrase(pwd)
				}

				kp, err := kb.Create(pwd, hint)
				if err != nil {
					return err
				}

				if err := kb.Stop(); err != nil {
					return err
				}

				logger.Global.Info().Str("address", kp.GetAddressString()).Msg("New Key Created")

				return nil
			},
		},
	}

	// Add --pwd and --hint flags
	applySubcommandOptions(cmds, attachPwdFlagToSubcommands())
	applySubcommandOptions(cmds, attachHintFlagToSubcommands())
	// Add --keybase flag
	applySubcommandOptions(cmds, attachKeybaseFlagsToSubcommands())

	return cmds
}

func keysUpdateCommands() []*cobra.Command {
	cmds := []*cobra.Command{
		{
			Use:     "Update <addrHex> [--pwd] [--new_pwd] [--hint]",
			Short:   "Updates the key to have a new passphrase and hint",
			Long:    "Updates the passphrase and hint of <addrHex> in the keybase, using either the values from the flags provided or from the CLI prompts.",
			Aliases: []string{"update"},
			Args:    cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				// Unpack CLI args
				addrHex := args[0]

				kb, err := keybaseForCLI()
				if err != nil {
					return err
				}

				if !nonInteractive {
					pwd = readPassphrase(pwd)
					newPwd = readPassphraseMessage(newPwd, "New passphrase: ")
					confirmPassphrase(newPwd)
				}

				err = kb.UpdatePassphrase(addrHex, pwd, newPwd, hint)
				if err != nil {
					return err
				}

				if err := kb.Stop(); err != nil {
					return err
				}

				logger.Global.Info().Str("address", addrHex).Msg("Key updated")

				return nil
			},
		},
	}

	// Add --pwd, --new_pwd and --hint flags
	applySubcommandOptions(cmds, attachPwdFlagToSubcommands())
	applySubcommandOptions(cmds, attachNewPwdFlagToSubcommands())
	applySubcommandOptions(cmds, attachHintFlagToSubcommands())
	// Add --keybase flag
	applySubcommandOptions(cmds, attachKeybaseFlagsToSubcommands())

	return cmds
}

func keysDeleteCommands() []*cobra.Command {
	cmds := []*cobra.Command{
		{
			Use:     "Delete <addrHex>",
			Short:   "Deletes the key from the keybase",
			Long:    "Deletes <addrHex> from the keybase",
			Aliases: []string{"delete"},
			Args:    cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				// Unpack CLI args
				addrHex := args[0]

				kb, err := keybaseForCLI()
				if err != nil {
					return err
				}

				if !nonInteractive {
					pwd = readPassphrase(pwd)
				}

				err = kb.Delete(addrHex, pwd)
				if err != nil {
					return err
				}

				if err := kb.Stop(); err != nil {
					return err
				}

				logger.Global.Info().Str("address", addrHex).Msg("Key deleted")

				return nil
			},
		},
	}

	// Add --pwd flag
	applySubcommandOptions(cmds, attachPwdFlagToSubcommands())

	// Add --keybase flag
	applySubcommandOptions(cmds, attachKeybaseFlagsToSubcommands())

	return cmds
}

func keysGetCommands() []*cobra.Command {
	cmds := []*cobra.Command{
		{
			Use:     "List",
			Short:   "List all keys",
			Long:    "List all of the hex addresses of the keys stored in the keybase",
			Aliases: []string{"list"},
			Args:    cobra.ExactArgs(0),
			RunE: func(cmd *cobra.Command, args []string) error {
				kb, err := keybaseForCLI()
				if err != nil {
					return err
				}

				addresses, _, err := kb.GetAll()
				if err != nil {
					return err
				}

				if err := kb.Stop(); err != nil {
					return err
				}

				logger.Global.Info().Strs("addresses", addresses).Msg("Get all keys")

				return nil
			},
		},
		{
			Use:     "Get <addrHex>",
			Short:   "Get the address and public key from the keybase",
			Long:    "Get the address and public key of <addrHex> from the keybase, provided it is stored",
			Aliases: []string{"get"},
			Args:    cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				// Unpack CLI args
				addrHex := args[0]

				kb, err := keybaseForCLI()
				if err != nil {
					return err
				}

				kp, err := kb.Get(addrHex)
				if err != nil {
					return err
				}

				if err := kb.Stop(); err != nil {
					return err
				}

				logger.Global.Info().Str("address", addrHex).Str("public_key", kp.GetPublicKey().String()).Msg("Found key")

				return nil
			},
		},
	}

	// Add --keybase flag
	applySubcommandOptions(cmds, attachKeybaseFlagsToSubcommands())

	return cmds
}

func keysExportCommands() []*cobra.Command {
	cmds := []*cobra.Command{
		{
			Use:     "Export <addrHex> [--export_format] [--output_file]",
			Short:   "Exports the private key as a raw string or JSON to either STDOUT or to a file",
			Long:    "Exports the private key of <addrHex> as a raw or JSON encoded string depending on [--output_format], written to STDOUT or [--output_file]",
			Aliases: []string{"export"},
			Args:    cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				// Unpack CLI args
				addrHex := args[0]

				kb, err := keybaseForCLI()
				if err != nil {
					return err
				}

				if !nonInteractive {
					pwd = readPassphrase(pwd)
				}

				// Select the correct format to export private key
				var exportString string
				switch strings.ToLower(exportAs) {
				case "json":
					exportString, err = kb.ExportPrivJSON(addrHex, pwd)
					if err != nil {
						return err
					}
				case "raw":
					exportString, err = kb.ExportPrivString(addrHex, pwd)
					if err != nil {
						return err
					}
				default:
					return fmt.Errorf("invalid export format: got %s, want [raw]/[json]", exportAs)
				}

				if err := kb.Stop(); err != nil {
					return err
				}

				// Write to stdout or file
				if outputFile == "" {
					logger.Global.Info().Str("private_key", exportString).Msg("Key exported")
					return nil
				}

				logger.Global.Info().Str("output_file", outputFile).Msg("Exporting private key string to file...")

				return utils.WriteOutput(exportString, outputFile)
			},
		},
	}

	// Add --pwd, --output_file and --export_format flags
	applySubcommandOptions(cmds, attachPwdFlagToSubcommands())
	applySubcommandOptions(cmds, attachOutputFlagToSubcommands())
	applySubcommandOptions(cmds, attachExportFlagToSubcommands())

	// Add --keybase flag
	applySubcommandOptions(cmds, attachKeybaseFlagsToSubcommands())

	return cmds
}

func keysImportCommands() []*cobra.Command {
	cmds := []*cobra.Command{
		{
			Use:     "Import [privateKeyString] [--input_file] [--import_format]",
			Short:   "Imports a key from a string or from a file",
			Long:    "Imports [privateKeyString] or from [--input_file] into the keybase, provided it is in the form of [--import_format]",
			Aliases: []string{"import"},
			Args:    cobra.MaximumNArgs(1),
			RunE: func(cmd *cobra.Command, args []string) (err error) {
				// Get import string
				var privateKeyString string
				if len(args) == 1 {
					privateKeyString = args[0]
				} else if inputFile != "" {
					privateKeyBz, err := utils.ReadInput(inputFile)
					privateKeyString = string(privateKeyBz)
					if err != nil {
						return err
					}
				} else {
					return fmt.Errorf("no input file or argument provided")
				}

				kb, err := keybaseForCLI()
				if err != nil {
					return err
				}

				if !nonInteractive {
					pwd = readPassphrase(pwd)
				}

				// Determine correct way to import the private key
				var kp crypto.KeyPair
				switch strings.ToLower(importAs) {
				case "json":
					kp, err = kb.ImportFromJSON(privateKeyString, pwd)
					if err != nil {
						return err
					}
				case "raw":
					// it is unarmoured so we need to confirm the passphrase
					if !nonInteractive {
						confirmPassphrase(pwd)
					}
					kp, err = kb.ImportFromString(privateKeyString, pwd, hint)
					if err != nil {
						return err
					}
				default:
					return fmt.Errorf("invalid import format: got %s, want [raw]/[json]", exportAs)
				}

				if err := kb.Stop(); err != nil {
					return err
				}

				logger.Global.Info().Str("address", kp.GetAddressString()).Msg("Key imported")

				return nil
			},
		},
	}

	// Add --pwd, --hint, --input_file and --import_format flags
	applySubcommandOptions(cmds, attachPwdFlagToSubcommands())
	applySubcommandOptions(cmds, attachHintFlagToSubcommands())
	applySubcommandOptions(cmds, attachInputFlagToSubcommands())
	applySubcommandOptions(cmds, attachImportFlagToSubcommands())

	// Add --keybase flag
	applySubcommandOptions(cmds, attachKeybaseFlagsToSubcommands())

	return cmds
}

func keysSignMsgCommands() []*cobra.Command {
	cmds := []*cobra.Command{
		{
			Use:     "Sign <addrHex> <messageHex>",
			Short:   "Signs a message using the key provided",
			Long:    "Signs <messageHex> with <addrHex> from the keybase, returning the signature",
			Aliases: []string{"sign"},
			Args:    cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				// Unpack CLI args
				addrHex := args[0]
				msgHex := args[1]
				msgBz, err := hex.DecodeString(msgHex)
				if err != nil {
					return err
				}

				kb, err := keybaseForCLI()
				if err != nil {
					return err
				}

				if !nonInteractive {
					pwd = readPassphrase(pwd)
				}

				sigBz, err := kb.Sign(addrHex, pwd, msgBz)
				if err != nil {
					return err
				}

				if err := kb.Stop(); err != nil {
					return err
				}

				sigHex := hex.EncodeToString(sigBz)

				logger.Global.Info().Str("signature", sigHex).Str("address", addrHex).Msg("Message signed")

				return nil
			},
		},
		{
			Use:     "Verify <addrHex> <messageHex> <signatureHex>",
			Short:   "Verifies the signature is valid from the signer",
			Long:    "Verify that <signatureHex> is a valid signature of <messageHex> signed by <addrHex>",
			Aliases: []string{"verify"},
			Args:    cobra.ExactArgs(3),
			RunE: func(cmd *cobra.Command, args []string) error {
				// Unpack CLI args
				addrHex := args[0]
				msgHex := args[1]
				msgBz, err := hex.DecodeString(msgHex)
				if err != nil {
					return err
				}
				sigHex := args[2]
				sigBz, err := hex.DecodeString(sigHex)
				if err != nil {
					return err
				}

				kb, err := keybaseForCLI()
				if err != nil {
					return err
				}

				valid, err := kb.Verify(addrHex, msgBz, sigBz)
				if err != nil {
					return err
				}

				if err := kb.Stop(); err != nil {
					return err
				}

				logger.Global.Info().Str("address", addrHex).Bool("valid", valid).Msg("Signature checked")

				return nil
			},
		},
	}

	// Add --pwd flag
	applySubcommandOptions(cmds, attachPwdFlagToSubcommands())
	// Add --keybase flag
	applySubcommandOptions(cmds, attachKeybaseFlagsToSubcommands())

	return cmds
}

func keysSignTxCommands() []*cobra.Command {
	cmds := []*cobra.Command{
		{
			Use:     "SignTx <addrHex> [--input_file] [--output_file]",
			Short:   "Signs a transaction using the key provided",
			Long:    "Signs [--input_file] with <addrHex> from the keybase, writing the signed transaction to [--output_file]",
			Aliases: []string{"signtx"},
			Args:    cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				// Unpack CLI args
				addrHex := args[0]

				if inputFile == "" {
					return fmt.Errorf("no input file provided")
				} else if outputFile == "" {
					return fmt.Errorf("no output file provided")
				}

				kb, err := keybaseForCLI()
				if err != nil {
					return err
				}

				if !nonInteractive {
					pwd = readPassphrase(pwd)
				}

				privKey, err := kb.GetPrivKey(addrHex, pwd)
				if err != nil {
					return err
				}

				if err := kb.Stop(); err != nil {
					return err
				}

				// Unmarshal Tx from input file
				txBz, err := utils.ReadInput(inputFile)
				if err != nil {
					return err
				}
				txProto := new(coreTypes.Transaction)
				if err := codec.GetCodec().Unmarshal(txBz, txProto); err != nil {
					return err
				}

				// Sign the serialised transaction
				txSigBz, err := txProto.SignableBytes()
				if err != nil {
					return err
				}

				sigBz, err := privKey.Sign(txSigBz)
				if err != nil {
					return err
				}

				// Add signature to the transaction
				sig := new(coreTypes.Signature)
				sig.PublicKey = privKey.PublicKey().Bytes()
				sig.Signature = sigBz
				txProto.Signature = sig

				// Re-serealise the transaction and write to output_file
				txBz, err = codec.GetCodec().Marshal(txProto)
				if err != nil {
					return err
				}

				if err := utils.WriteOutput(txBz, outputFile); err != nil {
					return err
				}

				logger.Global.Info().Str("signed_transaction_file", outputFile).Str("address", addrHex).Msg("Message signed")

				return nil
			},
		},
		{
			Use:     "VerifyTx <addrHex> [--input_file]",
			Short:   "Verifies the transaction's signature is valid from the signer",
			Long:    "Verify that [--input_file] contains a valid signature for the transaction in the file signed by <addrHex>",
			Aliases: []string{"verifytx"},
			Args:    cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				// Unpack CLI args
				addrHex := args[0]

				if inputFile == "" {
					return fmt.Errorf("no input file provided")
				}

				kb, err := keybaseForCLI()
				if err != nil {
					return err
				}

				pubKey, err := kb.GetPubKey(addrHex)
				if err != nil {
					return err
				}

				// Unmarshal Tx from input file
				txBz, err := utils.ReadInput(inputFile)
				if err != nil {
					return err
				}
				txProto := new(coreTypes.Transaction)
				if err := codec.GetCodec().Unmarshal(txBz, txProto); err != nil {
					return err
				}

				// Extract signature and begin verification
				var valid bool
				sigBz := txProto.Signature.Signature
				sigPub := txProto.Signature.PublicKey

				// First check public keys are the same
				if !bytes.Equal(sigPub, pubKey.Bytes()) {
					valid = false
				} else {
					// Verify the signable bytes of the transaction
					txSigBz, err := txProto.SignableBytes()
					if err != nil {
						return err
					}

					valid, err = kb.Verify(addrHex, txSigBz, sigBz)
					if err != nil {
						return err
					}
				}

				if err := kb.Stop(); err != nil {
					return err
				}

				logger.Global.Info().Str("address", addrHex).Bool("valid", valid).Msg("Signature checked")

				return nil
			},
		},
	}

	// Add --pwd, --input_file and --output_file flags
	applySubcommandOptions(cmds, attachPwdFlagToSubcommands())
	applySubcommandOptions(cmds, attachInputFlagToSubcommands())
	applySubcommandOptions(cmds, attachOutputFlagToSubcommands())
	// Add --keybase flag
	applySubcommandOptions(cmds, attachKeybaseFlagsToSubcommands())

	return cmds
}

func keysSlipCommands() []*cobra.Command {
	cmds := []*cobra.Command{
		{
			Use:     "DeriveChild <parentAddrHex> <index>",
			Short:   "Derive the child key at the given index from a parent key",
			Long:    "Derive the child key at <index> from the parent key provided optionally store it in the keybase with [--store_child]",
			Aliases: []string{"derivechild"},
			Args:    cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				// Unpack CLI args
				parentAddr := args[0]
				idx64, err := strconv.ParseUint(args[1], 10, 32)
				if err != nil {
					return err
				}
				index := uint32(idx64)

				kb, err := keybaseForCLI()
				if err != nil {
					return err
				}

				if !nonInteractive {
					pwd = readPassphrase(pwd)
				}

				kp, err := kb.DeriveChildFromKey(parentAddr, pwd, index, childPwd, childHint, storeChild)
				if err != nil {
					return err
				}

				if err := kb.Stop(); err != nil {
					return err
				}

				logger.Global.Info().Str("address", kp.GetAddressString()).Str("parent", parentAddr).Uint32("index", index).Bool("stored", storeChild).Msg("Child key derived")

				return nil
			},
		},
	}

	// Add --pwd, --store_child, --child_pwd, --child_hint flags
	applySubcommandOptions(cmds, attachPwdFlagToSubcommands())
	applySubcommandOptions(cmds, attachStoreChildFlagToSubcommands())
	applySubcommandOptions(cmds, attachChildPwdFlagToSubcommands())
	applySubcommandOptions(cmds, attachChildHintFlagToSubcommands())
	// Add --keybase flag
	applySubcommandOptions(cmds, attachKeybaseFlagsToSubcommands())

	return cmds
}
