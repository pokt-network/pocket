package cli

import (
	"bufio"
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"math/big"
	"os"
	"strings"

	"github.com/pokt-network/pocket/rpc"
	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/converters"
	"github.com/pokt-network/pocket/shared/crypto"
	typesUtil "github.com/pokt-network/pocket/utility/types"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// readEd25519PrivateKeyFromFile returns an Ed25519PrivateKey from a file where the file simply encodes it in a string (for now)
// HACK(#150): this is a temporary hack since we don't have yet a keybase, the next step would be to read from an "ArmoredJson" like in V0
func readEd25519PrivateKeyFromFile(pkPath string) (pk crypto.Ed25519PrivateKey, err error) {
	pkFile, err := os.Open(pkPath)
	if err != nil {
		return
	}
	defer pkFile.Close()
	pk, err = parseEd25519PrivateKeyFromReader(pkFile)
	return
}

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
		log.Fatalf(err.Error())
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
func prepareTxBytes(msg typesUtil.Message, pk crypto.Ed25519PrivateKey) ([]byte, error) {
	var err error
	anyMsg, err := codec.GetCodec().ToAny(msg)
	if err != nil {
		return nil, err
	}

	tx := &typesUtil.Transaction{
		Msg:   anyMsg,
		Nonce: fmt.Sprintf("%d", crypto.GetNonce()),
	}

	signBytes, err := tx.SignBytes()
	if err != nil {
		return nil, err
	}

	signature, err := pk.Sign(signBytes)
	if err != nil {
		return nil, err
	}

	tx.Signature = &typesUtil.Signature{
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
func postRawTx(ctx context.Context, pk crypto.Ed25519PrivateKey, j []byte) (*rpc.PostV1ClientBroadcastTxSyncResponse, error) {
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

func validateStakeAmount(amount string) error {
	am, err := converters.StringToBigInt(amount)
	if err != nil {
		return err
	}

	sr := big.NewInt(stakingRecommendationAmount)
	if typesUtil.BigIntLessThan(am, sr) {
		fmt.Printf("The amount you are staking for is below the recommendation of %d POKT, would you still like to continue? y|n\n", sr.Div(sr, oneMillion).Int64())
		if !confirmation(pwd) {
			return fmt.Errorf("aborted")
		}
	}
	return nil
}

func applySubcommandOptions(cmds []*cobra.Command, cmdDef actorCmdDef) {
	for _, cmd := range cmds {
		for _, opt := range cmdDef.Options {
			opt(cmd)
		}
	}
}

func attachPwdFlagToSubcommands() []cmdOption {
	return []cmdOption{func(c *cobra.Command) {
		c.Flags().StringVar(&pwd, "pwd", "", "passphrase used by the cmd, non empty usage bypass interactive prompt")
	}}
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
