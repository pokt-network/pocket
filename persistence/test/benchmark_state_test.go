package test

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"reflect"
	"testing"

	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/persistence/indexer"
	"github.com/pokt-network/pocket/shared/debug"
	"github.com/pokt-network/pocket/shared/modules"
)

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
		{1, 1},
		{100, 1},
		{100, 10},
		{1000, 1},
		{1000, 10},
		{10000, 10},
		{10000, 1000},
	}

	for _, testCase := range testCases {
		numHeights := testCase.numHeights
		numTxPerHeight := testCase.numTxPerHeight
		b.Run(fmt.Sprintf("heights=%d;txPerHeight=%d", numHeights, numTxPerHeight), func(b *testing.B) {
			for h := int64(0); h < numHeights; h++ {
				db := NewTestPostgresContext(b, h)
				helper(db)
				for i := int64(0); i < numTxPerHeight; i++ {
					// TODO: Perform a random operation
					db.StoreTransaction(modules.TxResult(getRandomTxResult(h)))
				}
				db.UpdateAppHash()
				db.Commit([]byte("TODOproposer"), []byte("TODOquorumCert"))
			}
		})
	}
}

func helper(p *persistence.PostgresContext) {
	v := reflect.ValueOf(p)
	// t := reflect.TypeOf(p)
	for m := 0; m < v.NumMethod(); m++ {
		var callArgs []reflect.Value
		method := v.Method(m).Type()
		// methodName := t.Method(m).Name
		for i := 0; i < method.NumIn(); i++ {
			arg := method.In(i)
			switch arg.Kind() {
			case reflect.String:
				v = reflect.ValueOf("123")
			case reflect.Slice:
				v = reflect.ValueOf([]byte("abc"))
			case reflect.Bool:
				v = reflect.ValueOf(false)
			case reflect.Uint8:
				fallthrough
			case reflect.Int32:
				fallthrough
			case reflect.Int64:
				fallthrough
			case reflect.Int:
				v = reflect.ValueOf(0)
			default:
				log.Println("OLSH, not supported", arg.Kind())
			}
			callArgs = append(callArgs, v)
		}
		// fmt.Println(methodName, callArgs)
	}
}

func getRandomTxResult(height int64) *indexer.TxRes {
	return &indexer.TxRes{
		Tx:            getTxBytes(50),
		Height:        height,
		Index:         0,
		ResultCode:    0,
		Error:         "TODO",
		SignerAddr:    "TODO",
		RecipientAddr: "TODO",
		MessageType:   "TODO",
	}
}

func getTxBytes(numBytes int64) []byte {
	bz := make([]byte, numBytes)
	rand.Read(bz)
	return bz
}
