package cli

import (
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

	getCmds := keysGetCommands()
	exportCmds := keysExportCommands()
	importCmds := keysImportCommands()

	// Add --pwd, --output_file and --export_as flags
	applySubcommandOptions(exportCmds, attachPwdFlagToSubcommands())
	applySubcommandOptions(exportCmds, attachOutputFlagToSubcommands())
	applySubcommandOptions(exportCmds, attachExportFlagToSubcommands())

	// Add --pwd, --hint, --input_file and --import_as flags
	applySubcommandOptions(importCmds, attachPwdFlagToSubcommands())
	applySubcommandOptions(importCmds, attachHintFlagToSubcommands())
	applySubcommandOptions(importCmds, attachInputFlagToSubcommands())
	applySubcommandOptions(importCmds, attachImportFlagToSubcommands())

	cmd.AddCommand(getCmds...)
	cmd.AddCommand(exportCmds...)
	cmd.AddCommand(importCmds...)

	return cmd
}

func keysGetCommands() []*cobra.Command {
	cmds := []*cobra.Command{
		{
			Use:     "List",
			Short:   "List",
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

				pwd := readPassphrase(pwd)

				// Determine correct way to export private key
				var exportString string
				exportAs := strings.ToLower(exportAs)
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

				pwd := readPassphrase(pwd)

				// Determine correct way to import the private key
				importAs := strings.ToLower(importAs)
				if importAs == "json" {
					err := kb.ImportFromJSON(privateKeyString, pwd)
					if err != nil {
						return err
					}
				} else if importAs == "raw" {
					err := kb.ImportFromString(privateKeyString, pwd, hint)
					if err != nil {
						return err
					}
				} else {
					return fmt.Errorf("invalid import format: got %s, want [raw]/[json]", exportAs)
				}

				return kb.Stop()
			},
		},
	}
	return cmds
}
