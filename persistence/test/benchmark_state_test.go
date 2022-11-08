package test

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"testing"

	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/persistence/indexer"
	"github.com/pokt-network/pocket/shared/debug"
	"github.com/pokt-network/pocket/shared/modules"
)

// var isModifierRe = regexp.MustCompile(`^(Insert|Update|Set|Add|Subtract)`)
var isModifierRe = regexp.MustCompile(`^(Insert|Set|Add|Subtract)`)

func BenchmarkStateHash(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	defer log.SetOutput(os.Stderr)

	b.Cleanup(func() {
		if err := testPersistenceMod.ReleaseWriteContext(); err != nil {
			log.Fatalf("Error releasing write context: %v\n", err)
		}
		if err := testPersistenceMod.HandleDebugMessage(&debug.DebugMessage{
			Action:  debug.DebugMessageAction_DEBUG_PERSISTENCE_CLEAR_STATE,
			Message: nil,
		}); err != nil {
			log.Fatalf("Error clearing state: %v\n", err)
		}

	})

	// Rather than using `b.N` and the `-benchtime` flag, we use a fixed number of iterations
	testCases := []struct {
		numHeights     int64
		numTxPerHeight int64
	}{
		// {1, 1},
		{1, 100},
		// {100, 1},
		// {100, 10},
		// {1000, 1},
		// {1000, 10},
		// {10000, 10},
		// {10000, 1000},
	}

	for _, testCase := range testCases {
		numHeights := testCase.numHeights
		numTxPerHeight := testCase.numTxPerHeight
		b.Run(fmt.Sprintf("heights=%d;txPerHeight=%d", numHeights, numTxPerHeight), func(b *testing.B) {
			for h := int64(0); h < numHeights; h++ {
				db := NewTestPostgresContext(b, h)
				for i := int64(0); i < numTxPerHeight; i++ {
					callRandomDatabaseModifierFunc(db, h, false)
					db.StoreTransaction(modules.TxResult(getRandomTxResult(h)))
				}
				db.UpdateAppHash()
				db.Commit([]byte("TODOproposer"), []byte("TODOquorumCert"))
			}
		})
	}
}

// Calls a random database modifier function on the given persistence context
func callRandomDatabaseModifierFunc(
	p *persistence.PostgresContext,
	height int64,
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
				v = reflect.ValueOf(getRandomIntString(1000000))
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
			case reflect.Pointer:
				fallthrough
			default:
				continue MethodLoop // IMPROVE: Other types not supported yet
			}
			callArgs = append(callArgs, v)
		}
		res := reflect.ValueOf(p).MethodByName(method.Name).Call(callArgs)
		var err error
		if v := res[0].Interface(); v != nil {
			if mustSucceed {
				fmt.Println("OLSH SKIP")
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
	return strconv.Itoa(rand.Intn(n))
}

func getRandomString(numChars int64) string {
	return string(getRandomBytes(numChars))
}

func getRandomBytes(numBytes int64) []byte {
	bz := make([]byte, numBytes)
	rand.Read(bz)
	return []byte(hex.EncodeToString(bz))
}
