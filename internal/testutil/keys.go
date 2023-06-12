package testutil

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"runtime"

	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/stretchr/testify/require"
)

var (
	privKeyManifestKeyRegex = regexp.MustCompile(`\s+"\d+":\s+(\w+)\s+`)
)

func LoadLocalnetPrivateKeys(t require.TestingT, keyCount int) (privKeys []cryptoPocket.PrivateKey) {
	_, filename, _, _ := runtime.Caller(0)
	pkgDir := filepath.Dir(filename)
	relativePathToKeys := filepath.Join(pkgDir, "..", "..", "build", "localnet", "manifests", "private-keys.yaml")

	privKeyManifest, err := os.Open(relativePathToKeys)
	require.NoError(t, err)

	privKeys = make([]cryptoPocket.PrivateKey, 0, keyCount)

	// scan through file & extract private keys
	scanner := bufio.NewScanner(privKeyManifest)
	scanner.Split(bufio.ScanLines)

	for i, done := 0, false; i < keyCount && !done; {
		done = !scanner.Scan()
		line := scanner.Text()
		matches := privKeyManifestKeyRegex.FindStringSubmatch(line)
		if len(matches) > 0 {
			privKey, err := cryptoPocket.NewPrivateKey(matches[1])
			require.NoError(t, err)

			privKeys = append(privKeys, privKey)
			i++
		}
	}
	return privKeys
}
