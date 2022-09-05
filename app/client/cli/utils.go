package cli

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/pokt-network/pocket/shared/crypto"
	"golang.org/x/crypto/ssh/terminal"
)

// readEd25519PrivateKeyFromFile returns an Ed25519PrivateKey from a file where the file simply encodes it in a string (for now)
// TODO(pocket/issues/150): this is a temporary hack since we don't have yet a keybase, the next step would be to read from an "ArmoredJson" like in V0
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
	buf.ReadFrom(reader)

	priv := &crypto.Ed25519PrivateKey{}
	err = priv.UnmarshalJSON(buf.Bytes())
	if err != nil {
		return
	}
	pk = priv.Bytes()
	return
}

// Credentials reads a password from the prompt and returns the trimmed version
//
// If pwd is provided (via flag to the command), it uses that one instead of asking via prompt
func Credentials(pwd string) string {
	if pwd != "" && strings.TrimSpace(pwd) != "" {
		return strings.TrimSpace(pwd)
	} else {
		bytePassword, err := terminal.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			fmt.Println(err)
		}
		return strings.TrimSpace(string(bytePassword))
	}
}

// Confirmation asks the user for a yes/no answer via interactive prompt.
//
// If pwd is provided (via flag to the command), it returns true since it's assumed that a user that provides a password via flag knows what they are doing
func Confirmation(pwd string) bool {
	if pwd != "" && strings.TrimSpace(pwd) != "" {
		return true
	} else {
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
}

func getNonce() string {
	rand.Seed(time.Now().UTC().UnixNano())
	return fmt.Sprintf("%d", rand.Uint64())
}
