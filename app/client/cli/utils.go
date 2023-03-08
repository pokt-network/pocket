package cli

import (
	"bufio"
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"strings"

	"github.com/pokt-network/pocket/app/client/keybase"
	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/rpc"
	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/utils"
	typesUtil "github.com/pokt-network/pocket/utility/types"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

const (
	KeybaseFile  = "file"
	KeybaseVault = "vault"
)

var (
	keybaseTypeFromCLI           string
	keybaseVaultAddrFromCLI      string
	keybaseVaultTokenFromCLI     string
	keybaseVaultMountPathFromCLI string
)

func parseEd25519PrivateKeyFromReader(reader io.Reader) (pk crypto.Ed25519PrivateKey, err error) {
	if reader == nil {
		return nil, fmt.Errorf("cannot read from reader %v", reader)
	}

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(reader); err != nil {
		return nil, err
	}

	priv := &crypto.Ed25519PrivateKey{}
	err = priv.UnmarshalJSON(buf.Bytes())
	if err != nil {
		return
	}
	pk = priv.Bytes()
	return
}

// credentials reads a password from the prompt and returns the trimmed version
//
// If pwd is provided (via flag to the command), it uses that one instead of asking via prompt
func credentials(pwd string) string {
	if pwd != "" && strings.TrimSpace(pwd) != "" {
		return strings.TrimSpace(pwd)
	}
	bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		logger.Global.Fatal().Err(err).Msg("failed to read password")
	}
	return strings.TrimSpace(string(bytePassword))
}

// confirmation asks the user for a yes/no answer via interactive prompt.
//
// If pwd is provided (via flag to the command), it returns true since it's assumed that a user that provides a password via flag knows what they are doing
func confirmation(pwd string) bool {
	if pwd != "" && strings.TrimSpace(pwd) != "" {
		return true
	}
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("yes | no")
		response, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading string: ", err.Error())
			return false
		}
		response = strings.ToLower(strings.TrimSpace(response))
		if response == "y" || response == "yes" {
			return true
		} else if response == "n" || response == "no" {
			return false
		}
	}
}

// prepareTxBytes wraps a Message into a Transaction and signs it with the provided pk
//
// returns the raw protobuf bytes of the signed transaction
func prepareTxBytes(msg typesUtil.Message, pk crypto.PrivateKey) ([]byte, error) {
	var err error
	anyMsg, err := codec.GetCodec().ToAny(msg)
	if err != nil {
		return nil, err
	}

	tx := &coreTypes.Transaction{
		Msg:   anyMsg,
		Nonce: fmt.Sprintf("%d", crypto.GetNonce()),
	}

	signBytes, err := tx.SignableBytes()
	if err != nil {
		return nil, err
	}

	signature, err := pk.Sign(signBytes)
	if err != nil {
		return nil, err
	}

	tx.Signature = &coreTypes.Signature{
		Signature: signature,
		PublicKey: pk.PublicKey().Bytes(),
	}

	bz, err := codec.GetCodec().Marshal(tx)
	if err != nil {
		return nil, err
	}
	return bz, nil
}

// postRawTx posts a signed transaction
func postRawTx(ctx context.Context, pk crypto.PrivateKey, j []byte) (*rpc.PostV1ClientBroadcastTxSyncResponse, error) {
	client, err := rpc.NewClientWithResponses(remoteCLIURL)
	if err != nil {
		return nil, err
	}
	req := rpc.RawTXRequest{
		Address:     pk.Address().String(),
		RawHexBytes: hex.EncodeToString(j),
	}

	resp, err := client.PostV1ClientBroadcastTxSyncWithResponse(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func readPassphrase(currPwd string) string {
	if strings.TrimSpace(currPwd) == "" {
		fmt.Println("Enter Passphrase: ")
	} else {
		fmt.Println("Using Passphrase provided via flag")
	}

	return credentials(currPwd)
}

func readPassphraseMessage(currPwd, prompt string) string {
	if strings.TrimSpace(currPwd) == "" {
		fmt.Println(prompt)
	} else {
		fmt.Println("Using Passphrase provided via flag")
	}

	return credentials(currPwd)
}

func validateStakeAmount(amount string) error {
	am, err := utils.StringToBigInt(amount)
	if err != nil {
		return err
	}

	sr := big.NewInt(stakingRecommendationAmount)
	if utils.BigIntLessThan(am, sr) {
		fmt.Printf("The amount you are staking for is below the recommendation of %d POKT, would you still like to continue? y|n\n", sr.Div(sr, oneMillion).Int64())
		if !confirmation(pwd) {
			return fmt.Errorf("aborted")
		}
	}
	return nil
}

func applySubcommandOptions(cmds []*cobra.Command, cmdOptions []cmdOption) {
	for _, cmd := range cmds {
		for _, opt := range cmdOptions {
			opt(cmd)
		}
	}
}

func attachPwdFlagToSubcommands() []cmdOption {
	return []cmdOption{func(c *cobra.Command) {
		c.Flags().StringVar(&pwd, "pwd", "", "passphrase used by the cmd, non empty usage bypass interactive prompt")
	}}
}

func attachNewPwdFlagToSubcommands() []cmdOption {
	return []cmdOption{func(c *cobra.Command) {
		c.Flags().StringVar(&pwd, "new_pwd", "", "new passphrase for private key, non empty usage bypass interactive prompt")
	}}
}

func attachOutputFlagToSubcommands() []cmdOption {
	return []cmdOption{func(c *cobra.Command) {
		c.Flags().StringVar(&outputFile, "output_file", "", "output file to write results to")
	}}
}

func attachInputFlagToSubcommands() []cmdOption {
	return []cmdOption{func(c *cobra.Command) {
		c.Flags().StringVar(&inputFile, "input_file", "", "input file to read data from")
	}}
}

func attachExportFlagToSubcommands() []cmdOption {
	return []cmdOption{func(c *cobra.Command) {
		c.Flags().StringVar(&exportAs, "export_format", "json", "export the private key in the specified format")
	}}
}

func attachImportFlagToSubcommands() []cmdOption {
	return []cmdOption{func(c *cobra.Command) {
		c.Flags().StringVar(&importAs, "import_format", "raw", "import the private key from the specified format")
	}}
}

func attachHintFlagToSubcommands() []cmdOption {
	return []cmdOption{func(c *cobra.Command) {
		c.Flags().StringVar(&hint, "hint", "", "hint for the passphrase of the private key")
	}}
}

func attachStoreChildFlagToSubcommands() []cmdOption {
	return []cmdOption{func(c *cobra.Command) {
		c.Flags().BoolVar(&storeChild, "store_child", true, "store the derived child key in the keybase")
	}}
}

func attachChildHintFlagToSubcommands() []cmdOption {
	return []cmdOption{func(c *cobra.Command) {
		c.Flags().StringVar(&childHint, "child_hint", "", "hint for the passphrase of the derived child's private key")
	}}
}

func attachChildPwdFlagToSubcommands() []cmdOption {
	return []cmdOption{func(c *cobra.Command) {
		c.Flags().StringVar(&childPwd, "child_pwd", "", "passphrase for the derived child's private key")
	}}
}

func attachKeybaseFlagsToSubcommands() []cmdOption {
	return []cmdOption{func(c *cobra.Command) {
		c.Flags().StringVar(&keybaseTypeFromCLI, "keybase", "file", "keybase type used by the cmd, options are: file, vault")
		c.Flags().StringVar(&keybaseVaultAddrFromCLI, "vault-addr", "", "Vault address used by the cmd. Defaults to https://127.0.0.1:8200 or VAULT_ADDR env var")
		c.Flags().StringVar(&keybaseVaultTokenFromCLI, "vault-token", "", "Vault token used by the cmd. Defaults to VAULT_TOKEN env var")
		c.Flags().StringVar(&keybaseVaultMountPathFromCLI, "vault-mount", "", "Vault mount path used by the cmd. Defaults to secret")
	}}
}

func keybaseForCLI() (keybase.Keybase, error) {
	keybaseType := keybase.KeybaseTypeFile
	keybaseOptions := keybase.KeybaseOptions{}

	switch keybaseTypeFromCLI {
	case KeybaseFile:
		keybaseType = keybase.KeybaseTypeFile
		// Open the keybase at the specified directory
		pocketDir := strings.TrimSuffix(dataDir, "/")
		keybasePath, err := filepath.Abs(pocketDir + keybaseSuffix)
		if err != nil {
			return nil, err
		}
		keybaseOptions.KeybasePath = keybasePath
	case KeybaseVault:
		keybaseType = keybase.KeybaseTypeVault
		keybaseOptions.VaultAddr = keybaseVaultAddrFromCLI
		keybaseOptions.VaultToken = keybaseVaultTokenFromCLI
		keybaseOptions.VaultMountPath = keybaseVaultMountPathFromCLI
	default:
		logger.Global.Fatal().Msgf("invalid keybase type: %s", keybaseTypeFromCLI)
	}
	return keybase.NewKeybase(keybaseType, &keybaseOptions)
}

func unableToConnectToRpc(err error) error {
	fmt.Printf("❌ Unable to connect to the RPC @ %s\n\nError: %s", boldText(remoteCLIURL), err)
	return nil
}

func rpcResponseCodeUnhealthy(statusCode int, response []byte) error {
	fmt.Printf("❌ RPC reporting unhealthy status HTTP %d @ %s\n\n%s", statusCode, boldText(remoteCLIURL), response)
	return nil
}

func boldText[T string | []byte](s T) string {
	return fmt.Sprintf("\033[1m%s\033[0m", s)
}

func setValueInCLIContext(cmd *cobra.Command, key cliContextKey, value any) {
	cmd.SetContext(context.WithValue(cmd.Context(), key, value))
}

func getValueFromCLIContext[T any](cmd *cobra.Command, key cliContextKey) (T, bool) {
	value, ok := cmd.Context().Value(key).(T)
	return value, ok
}

// confirmPassphrase should be used when a new key is being created or a raw unarmored key is being imported
func confirmPassphrase(currPwd string) {
	confirm := readPassphraseMessage("", "Confirm passphrase: ")
	if currPwd != confirm {
		logger.Global.Fatal().Msg("❌ Passphrases do not match")
	}
}
