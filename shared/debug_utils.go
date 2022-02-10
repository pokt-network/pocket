package shared

import (
	"encoding/json"
	"log"
	"runtime/debug"
)

func DebugDeferStackTrackPrint() {
	defer func() {
		if r := recover(); r != nil {
			log.Println("[ERROR] stacktrace from panic: \n" + string(debug.Stack()))
		}
	}()
}

func DebugPrintStruct(s interface{}) {
	str, _ := json.MarshalIndent(s, "", "\t")
	log.Println("[DEBUG]", string(str))
}
