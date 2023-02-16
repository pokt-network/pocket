package cli

import (
	"encoding/hex"
	"fmt"
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
	deleteCmds := keysDeleteCommands()
	updateCmds := keysUpdateCommands()
	getCmds := keysGetCommands()
	exportCmds := keysExportCommands()
	importCmds := keysImportCommands()
	signCmds := keysSignCommands()

	// Add --pwd and --hint flags
	applySubcommandOptions(createCmds, attachPwdFlagToSubcommands())
	applySubcommandOptions(createCmds, attachHintFlagToSubcommands())

	// Add --pwd flag
	applySubcommandOptions(deleteCmds, attachPwdFlagToSubcommands())

	// Add --pwd and --hint flags
	applySubcommandOptions(updateCmds, attachPwdFlagToSubcommands())
	applySubcommandOptions(updateCmds, attachNewPwdFlagToSubcommands())
	applySubcommandOptions(updateCmds, attachHintFlagToSubcommands())

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
	applySubcommandOptions(signCmds, attachPwdFlagToSubcommands())

	cmd.AddCommand(createCmds...)
	cmd.AddCommand(deleteCmds...)
	cmd.AddCommand(getCmds...)
	cmd.AddCommand(exportCmds...)
	cmd.AddCommand(importCmds...)
	cmd.AddCommand(signCmds...)

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

				pwd = readPassphrase(pwd)

				kp, err := kb.Create(pwd, hint)
				if err != nil {
					return err
				}

				fmt.Printf("Key created: %s\n", kp.GetAddressString())

				return kb.Stop()
			},
		},
	}
	return cmds
}

func keysUpdateCommands() []*cobra.Command {
	cmds := []*cobra.Command{
		{
			Use:     "Update <addrHex>",
			Short:   "Update <addrHex>",
			Long:    "Updates the passphrase of <addrHex> in the keybase",
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

				pwd = readPassphrase(pwd)

				newPwd = readPassphraseMessage(newPwd, "New passphrase: ")

				err = kb.UpdatePassphrase(addrHex, pwd, newPwd, hint)
				if err != nil {
					return err
				}

				fmt.Printf("Updated key: %s\n", addrHex)

				return kb.Stop()
			},
		},
	}
	return cmds
}

func keysDeleteCommands() []*cobra.Command {
	cmds := []*cobra.Command{
		{
			Use:     "Delete <addrHex>",
			Short:   "Delete <addrHex>",
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

				pwd = readPassphrase(pwd)

				err = kb.Delete(addrHex, pwd)
				if err != nil {
					return err
				}

				fmt.Printf("Key deleted: %s\n", addrHex)

				return kb.Stop()
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

				for _, addr := range addresses {
					fmt.Println(addr)
				}

				return kb.Stop()
			},
		},
		{
			Use:     "Get <addrHex>",
			Short:   "Get <addrHex>",
			Long:    "Get the keypair <addrHex> from the keybase",
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

				fmt.Printf("Address: %s\nPublic Key: %s\n", addrHex, kp.GetPublicKey().String())

				return kb.Stop()
			},
		},
	}
	return cmds
}

func keysExportCommands() []*cobra.Command {
	cmds := []*cobra.Command{
		{
			Use:     "Export <addrHex>",
			Short:   "Export <addrHex>",
			Long:    "Exports the private key of <addrHex> as a raw or JSON encoded string",
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

				pwd = readPassphrase(pwd)

				// Determine correct way to export private key
				var exportString string
				exportAs = strings.ToLower(exportAs)
				if exportAs == "json" {
					exportString, err = kb.ExportPrivJSON(addrHex, pwd)
					if err != nil {
						return err
					}
				} else if exportAs == "raw" {
					exportString, err = kb.ExportPrivString(addrHex, pwd)
					if err != nil {
						return err
					}
				} else {
					return fmt.Errorf("invalid export format: got %s, want [raw]/[json]", exportAs)
				}

				if err := kb.Stop(); err != nil {
					return err
				}

				// Write to stdout or file
				if outputFile == "" {
					fmt.Println(exportString)
					return nil
				}

				return writeOutput(exportString, outputFile)
			},
		},
	}
	return cmds
}

func keysImportCommands() []*cobra.Command {
	cmds := []*cobra.Command{
		{
			Use:     "Import <privateKeyString>",
			Short:   "Import <privateKeyString>",
			Long:    "Imports <privateKeyString> into the keybase",
			Aliases: []string{"import"},
			Args:    cobra.MaximumNArgs(1),
			RunE: func(cmd *cobra.Command, args []string) (err error) {
				// Get import string
				var privateKeyString string
				if len(args) == 1 {
					privateKeyString = args[0]
				} else if inputFile != "" {
					privateKeyString, err = readInput(inputFile)
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

				pwd = readPassphrase(pwd)

				// Determine correct way to import the private key
				importAs = strings.ToLower(importAs)
				if importAs == "json" {
					kp, err := kb.ImportFromJSON(privateKeyString, pwd)
					if err != nil {
						return err
					}
					fmt.Printf("Key imported: %s\n", kp.GetAddressString())
				} else if importAs == "raw" {
					kp, err := kb.ImportFromString(privateKeyString, pwd, hint)
					if err != nil {
						return err
					}
					fmt.Printf("Key imported: %s\n", kp.GetAddressString())
				} else {
					return fmt.Errorf("invalid import format: got %s, want [raw]/[json]", exportAs)
				}

				return kb.Stop()
			},
		},
	}
	return cmds
}

func keysSignCommands() []*cobra.Command {
	cmds := []*cobra.Command{
		{
			Use:     "Sign <addrHex> <messageHex>",
			Short:   "Sign <addrHex> <messageHex>",
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

				pwd = readPassphrase(pwd)

				sigBz, err := kb.Sign(addrHex, pwd, msgBz)
				if err != nil {
					return err
				}

				sigHex := hex.EncodeToString(sigBz)

				fmt.Printf("Signature: %s\n", sigHex)

				return kb.Stop()
			},
		},
		{
			Use:     "Verify <addrHex> <messageHex> <signatureHex>",
			Short:   "Verify <addrHex> <messageHex> <signatureHex>",
			Long:    "Verify that <signatureHex> is a valid signature of <messageHex> signed by <addrHex>",
			Aliases: []string{"sign"},
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

				fmt.Printf("Valid signature: %v", valid)

				return kb.Stop()
			},
		},
	}
	return cmds
}
