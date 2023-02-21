package cli

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/converters"
	"github.com/pokt-network/pocket/utility/types"
	"path/filepath"
	"strings"

	"github.com/pokt-network/pocket/app/client/keybase"
	"github.com/spf13/cobra"
)

var (
	outputFile string
	inputFile  string
	exportAs   string
	importAs   string
	hint       string
	newPwd     string
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

	createCmds := keysCreateCommands()
	updateCmds := keysUpdateCommands()
	deleteCmds := keysDeleteCommands()
	getCmds := keysGetCommands()
	exportCmds := keysExportCommands()
	importCmds := keysImportCommands()
	signMsgCmds := keysSignMsgCommands()
	signTxCmds := keysSignTxCommands()

	// Add --pwd and --hint flags
	applySubcommandOptions(createCmds, attachPwdFlagToSubcommands())
	applySubcommandOptions(createCmds, attachHintFlagToSubcommands())

	// Add --pwd, --new_pwd and --hint flags
	applySubcommandOptions(updateCmds, attachPwdFlagToSubcommands())
	applySubcommandOptions(updateCmds, attachNewPwdFlagToSubcommands())
	applySubcommandOptions(updateCmds, attachHintFlagToSubcommands())

	// Add --pwd flag
	applySubcommandOptions(deleteCmds, attachPwdFlagToSubcommands())

	// Add --pwd, --output_file and --export_format flags
	applySubcommandOptions(exportCmds, attachPwdFlagToSubcommands())
	applySubcommandOptions(exportCmds, attachOutputFlagToSubcommands())
	applySubcommandOptions(exportCmds, attachExportFlagToSubcommands())

	// Add --pwd, --hint, --input_file and --import_format flags
	applySubcommandOptions(importCmds, attachPwdFlagToSubcommands())
	applySubcommandOptions(importCmds, attachHintFlagToSubcommands())
	applySubcommandOptions(importCmds, attachInputFlagToSubcommands())
	applySubcommandOptions(importCmds, attachImportFlagToSubcommands())

	// Add --pwd flag
	applySubcommandOptions(signMsgCmds, attachPwdFlagToSubcommands())

	// Add --pwd, --input_file and --output_file flags
	applySubcommandOptions(signTxCmds, attachPwdFlagToSubcommands())
	applySubcommandOptions(signTxCmds, attachInputFlagToSubcommands())
	applySubcommandOptions(signTxCmds, attachOutputFlagToSubcommands())

	cmd.AddCommand(createCmds...)
	cmd.AddCommand(updateCmds...)
	cmd.AddCommand(deleteCmds...)
	cmd.AddCommand(getCmds...)
	cmd.AddCommand(exportCmds...)
	cmd.AddCommand(importCmds...)
	cmd.AddCommand(signMsgCmds...)
	cmd.AddCommand(signTxCmds...)

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
				// Open the debug keybase at the specified path
				pocketDir := strings.TrimSuffix(dataDir, "/")
				keybasePath, err := filepath.Abs(pocketDir + keybaseSuffix)
				if err != nil {
					return err
				}
				kb, err := keybase.NewKeybase(keybasePath)
				if err != nil {
					return err
				}

				if !nonInteractive {
					pwd = readPassphrase(pwd)
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

				// Open the debug keybase at the specified path
				pocketDir := strings.TrimSuffix(dataDir, "/")
				keybasePath, err := filepath.Abs(pocketDir + keybaseSuffix)
				if err != nil {
					return err
				}
				kb, err := keybase.NewKeybase(keybasePath)
				if err != nil {
					return err
				}

				if !nonInteractive {
					pwd = readPassphrase(pwd)
					newPwd = readPassphraseMessage(newPwd, "New passphrase: ")
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

				// Open the debug keybase at the specified path
				pocketDir := strings.TrimSuffix(dataDir, "/")
				keybasePath, err := filepath.Abs(pocketDir + keybaseSuffix)
				if err != nil {
					return err
				}
				kb, err := keybase.NewKeybase(keybasePath)
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
				// Open the debug keybase at the specified path
				pocketDir := strings.TrimSuffix(dataDir, "/")
				keybasePath, err := filepath.Abs(pocketDir + keybaseSuffix)
				if err != nil {
					return err
				}
				kb, err := keybase.NewKeybase(keybasePath)
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

				for _, addr := range addresses {
					fmt.Println(addr)
				}

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

				// Open the debug keybase at the specified path
				pocketDir := strings.TrimSuffix(dataDir, "/")
				keybasePath, err := filepath.Abs(pocketDir + keybaseSuffix)
				if err != nil {
					return err
				}
				kb, err := keybase.NewKeybase(keybasePath)
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

				// Open the debug keybase at the specified path
				pocketDir := strings.TrimSuffix(dataDir, "/")
				keybasePath, err := filepath.Abs(pocketDir + keybaseSuffix)
				if err != nil {
					return err
				}
				kb, err := keybase.NewKeybase(keybasePath)
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

				return converters.WriteOutput(exportString, outputFile)
			},
		},
	}
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
					privateKeyBz, err := converters.ReadInput(inputFile)
					privateKeyString = string(privateKeyBz)
					if err != nil {
						return err
					}
				} else {
					return fmt.Errorf("no input file provided")
				}

				// Open the debug keybase at the specified path
				pocketDir := strings.TrimSuffix(dataDir, "/")
				keybasePath, err := filepath.Abs(pocketDir + keybaseSuffix)
				if err != nil {
					return err
				}
				kb, err := keybase.NewKeybase(keybasePath)
				if err != nil {
					return err
				}

				if !nonInteractive {
					pwd = readPassphrase(pwd)
				}

				// Determine correct way to import the private key
				switch strings.ToLower(importAs) {
				case "json":
					kp, err := kb.ImportFromJSON(privateKeyString, pwd)
					if err != nil {
						return err
					}
					logger.Global.Info().Str("address", kp.GetAddressString()).Msg("Key imported")
				case "raw":
					kp, err := kb.ImportFromString(privateKeyString, pwd, hint)
					if err != nil {
						return err
					}
					logger.Global.Info().Str("address", kp.GetAddressString()).Msg("Key imported")
				default:
					return fmt.Errorf("invalid import format: got %s, want [raw]/[json]", exportAs)
				}

				if err := kb.Stop(); err != nil {
					return err
				}

				return nil
			},
		},
	}
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

				// Open the debug keybase at the specified path
				pocketDir := strings.TrimSuffix(dataDir, "/")
				keybasePath, err := filepath.Abs(pocketDir + keybaseSuffix)
				if err != nil {
					return err
				}
				kb, err := keybase.NewKeybase(keybasePath)
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

				// Open the debug keybase at the specified path
				pocketDir := strings.TrimSuffix(dataDir, "/")
				keybasePath, err := filepath.Abs(pocketDir + keybaseSuffix)
				if err != nil {
					return err
				}
				kb, err := keybase.NewKeybase(keybasePath)
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

				// Open the debug keybase at the specified path
				pocketDir := strings.TrimSuffix(dataDir, "/")
				keybasePath, err := filepath.Abs(pocketDir + keybaseSuffix)
				if err != nil {
					return err
				}
				kb, err := keybase.NewKeybase(keybasePath)
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
				txBz, err := converters.ReadInput(inputFile)
				if err != nil {
					return err
				}
				txProto := new(types.Transaction)
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

				sig := new(types.Signature)
				sig.PublicKey = privKey.PublicKey().Bytes()
				sig.Signature = sigBz
				txProto.Signature = sig

				// Re-serealise the transaction and write to output_file
				txBz, err = codec.GetCodec().Marshal(txProto)
				if err != nil {
					return err
				}

				if err := converters.WriteOutput(txBz, outputFile); err != nil {
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

				// Open the debug keybase at the specified path
				pocketDir := strings.TrimSuffix(dataDir, "/")
				keybasePath, err := filepath.Abs(pocketDir + keybaseSuffix)
				if err != nil {
					return err
				}
				kb, err := keybase.NewKeybase(keybasePath)
				if err != nil {
					return err
				}

				pubKey, err := kb.GetPubKey(addrHex)
				if err != nil {
					return err
				}

				// Unmarshal Tx from input file
				txBz, err := converters.ReadInput(inputFile)
				if err != nil {
					return err
				}
				txProto := new(types.Transaction)
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
	return cmds
}
