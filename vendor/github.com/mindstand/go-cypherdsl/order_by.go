package go_cypherdsl

import (
	"errors"
	"fmt"
)

type OrderByConfig struct {
	Name   string
	Member string
	Desc   bool
}

func (o *OrderByConfig) ToString() (string, error) {
	if o.Name == "" || o.Member == "" {
		return "", errors.New("name and member have to be defined")
	}

	if o.Desc {
		return fmt.Sprintf("%s.%s DESC", o.Name, o.Member), nil
	}
	return fmt.Sprintf("%s.%s", o.Name, o.Member), nil
}
