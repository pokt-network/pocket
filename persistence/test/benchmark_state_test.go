package test

import (
	crand "crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"testing"

	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/persistence/indexer"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/stretchr/testify/require"
)

const (
	maxStringAmount = 1000000000000000000
)

var isModifierRe = regexp.MustCompile(`^(Insert|Set|Add|Subtract)`) // Add Update?

// INVESTIGATE: This benchmark can be used to experiment with different Merkle Tree implementations
// and key-value stores.
// IMPROVE(#361): Improve the output of this benchmark to be more informative and human readable.
func BenchmarkStateHash(b *testing.B) {
	log.SetOutput(io.Discard)
	defer log.SetOutput(os.Stderr)

	clearAllState()
	b.Cleanup(clearAllState)

	// NOTE: The idiomatic way to run Go benchmarks is to use `b.N` and the `-benchtime` flag,
	// to specify how long the benchmark should take. However, the code below is non-idiomatic
	// since our goal is to test a specific we use a fixed number of iterations
	testCases := []struct {
		numHeights     int64
		numTxPerHeight int
		numOpsPerTx    int
	}{
		{1, 1, 1},
		{1, 1, 10},
		{1, 10, 10},

		{10, 1, 1},
		{10, 1, 10},
		{10, 10, 10},

		// This takes a VERY long time to run
		// {100, 1, 1},
		// {100, 1, 100},
		// {100, 100, 100},
	}

	for _, testCase := range testCases {
		numHeights := testCase.numHeights
		numTxPerHeight := testCase.numTxPerHeight
		numOpsPerTx := testCase.numOpsPerTx

		// Since this is a benchmark, errors are not
		b.Run(fmt.Sprintf("height=%d;txs=%d,ops=%d", numHeights, numTxPerHeight, numOpsPerTx), func(b *testing.B) {
			for height := int64(0); height < numHeights; height++ {
				db := NewTestPostgresContext(b, height)
				for txIdx := 0; txIdx < numTxPerHeight; txIdx++ {
					for opIdx := 0; opIdx < numOpsPerTx; opIdx++ {
						_, _, err := callRandomDatabaseModifierFunc(db, false)
						require.NoError(b, err)
					}
					err := db.IndexTransaction(modules.TxResult(getRandomTxResult(height)))
					require.NoError(b, err)
				}
				_, err := db.ComputeStateHash()
				require.NoError(b, err)
				err = db.Commit([]byte("placeholderProposerAddr"), []byte("placeholderQuorumCert"))
				require.NoError(b, err)
				err = db.Release()
				require.NoError(b, err)
			}
		})
	}
}

// Calls a random database modifier function on the given persistence context
//
//nolint:gosec // G404 - Weak random source is okay in unit tests
func callRandomDatabaseModifierFunc(
	p *persistence.PostgresContext,
	mustSucceed bool,
) (string, []reflect.Value, error) {
	t := reflect.TypeOf(modules.PersistenceWriteContext(p))
	numMethods := t.NumMethod()

	// Select a random method and loops until a successful invocation takes place
MethodLoop:
	for {
		method := t.Method(rand.Intn(numMethods))
		methodName := method.Name
		numArgs := method.Type.NumIn()

		// Preliminary filter to determine which functions we're interested in trying to call
		if !isModifierRe.MatchString(methodName) {
			continue
		}

		// Build a random set of arguments to pass to the function being called
		var callArgs []reflect.Value
		for i := 1; i < numArgs; i++ {
			var v reflect.Value
			arg := method.Type.In(i)
			switch arg.Kind() {
			case reflect.String:
				// String values in modifier functions are usually amounts
				v = reflect.ValueOf(getRandomIntString(maxStringAmount))
			case reflect.Slice:
				switch arg.Elem().Kind() {
				case reflect.Uint8:
					v = reflect.ValueOf([]uint8{0})
				case reflect.String:
					v = reflect.ValueOf([]string{"abc"})
				default:
					continue MethodLoop // IMPROVE: Slices of other types not supported yet
				}
			case reflect.Bool:
				v = reflect.ValueOf(rand.Intn(2) == 1)
			case reflect.Uint8:
				v = reflect.ValueOf(uint8(rand.Intn(2 ^ 8 - 1)))
			case reflect.Int32:
				v = reflect.ValueOf(rand.Int31())
			case reflect.Int64:
				v = reflect.ValueOf(rand.Int63())
			case reflect.Int:
				v = reflect.ValueOf(rand.Int())
			default:
				continue MethodLoop // IMPROVE: Other types not supported yet
			}
			callArgs = append(callArgs, v)
		}
		res := reflect.ValueOf(p).MethodByName(method.Name).Call(callArgs)
		var err error
		if v := res[0].Interface(); v != nil {
			if mustSucceed {
				continue MethodLoop
			}
			err = v.(error)
		}
		return methodName, callArgs, err
	}
}

func getRandomTxResult(height int64) *indexer.TxRes {
	return &indexer.TxRes{
		Tx:            getRandomBytes(50),
		Height:        height,
		Index:         0,
		ResultCode:    0,
		Error:         "TODO",
		SignerAddr:    "TODO",
		RecipientAddr: "TODO",
		MessageType:   "TODO",
	}
}

func getRandomIntString(n int) string {
	return strconv.Itoa(rand.Intn(n)) //nolint:gosec // G404 - Weak random source is okay in unit tests
}

func getRandomBytes(numBytes int64) []byte {
	bz := make([]byte, numBytes)
	_, err := crand.Read(bz)
	if err != nil {
		panic(err)
	}
	return []byte(hex.EncodeToString(bz))
}
