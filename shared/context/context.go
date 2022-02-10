package context

import (
	"context"
	"log"
)

type PocketContext struct {
	Ctx     context.Context
	Handler func(...interface{}) (interface{}, error)
}

func EmptyPocketContext() *PocketContext {
	return &PocketContext{
		Ctx: context.Background(),
		Handler: func(...interface{}) (interface{}, error) {
			log.Println("[DEBUG] NOOP HANDLER ")
			return nil, nil
		},
	}
}
