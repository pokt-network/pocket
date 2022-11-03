package test

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"reflect"
	"regexp"
	"testing"

	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/persistence/indexer"
	"github.com/pokt-network/pocket/shared/debug"
	"github.com/pokt-network/pocket/shared/modules"
)

var re = regexp.MustCompile(`^[Insert|Update|Set|Add|Subtract]`)

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
					callRandomModifierFunc(db, h)
					db.StoreTransaction(modules.TxResult(getRandomTxResult(h)))
				}
				db.UpdateAppHash()
				db.Commit([]byte("TODOproposer"), []byte("TODOquorumCert"))
			}
		})
	}
}

func callRandomModifierFunc(p *persistence.PostgresContext, height int64) error {
	t := reflect.TypeOf(modules.PersistenceWriteContext(p))

MethodLoop:
	for m := 0; m < t.NumMethod(); m++ {
		method := t.Method(m)
		methodName := method.Name

		if !re.MatchString(methodName) {
			continue
		}

		var callArgs []reflect.Value
		for i := 1; i < method.Type.NumIn(); i++ {
			var v reflect.Value
			arg := method.Type.In(i)
			switch arg.Kind() {
			case reflect.String:
				v = reflect.ValueOf(getRandomString(50))
			case reflect.Slice:
				switch arg.Elem().Kind() {
				case reflect.Uint8:
					v = reflect.ValueOf([]uint8{0})
				case reflect.String:
					v = reflect.ValueOf([]string{"abc"})
				default:
					continue MethodLoop
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
				continue MethodLoop
			}
			callArgs = append(callArgs, v)
		}
		// fmt.Println(methodName, "~~~", method.Type.NumIn(), callArgs)
		// return reflect.ValueOf(p).MethodByName(method.Name).Call(callArgs)
		reflect.ValueOf(p).MethodByName(method.Name).Call(callArgs)
	}
	return nil
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

func getRandomString(numChars int64) string {
	return string(getRandomBytes(numChars))
}

func getRandomBytes(numBytes int64) []byte {
	bz := make([]byte, numBytes)
	rand.Read(bz)
	return bz
}
