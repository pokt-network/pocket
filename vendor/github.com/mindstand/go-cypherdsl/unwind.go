package go_cypherdsl

import (
	"errors"
	"fmt"
	"strings"
)

type UnwindConfig struct {
	Slice []interface{}
	As    string
}

func (u *UnwindConfig) ToString() (string, error) {
	if u.Slice == nil || len(u.Slice) == 0 {
		return "", errors.New("slice in unwind can not be empty")
	}

	if u.As == "" {
		return "", errors.New("AS has to be defined")
	}

	query := "["

	for _, i := range u.Slice {
		str, err := cypherizeInterface(i)
		if err != nil {
			return "", err
		}

		query += fmt.Sprintf("%s,", str)
	}

	query = strings.TrimSuffix(query, ",")

	return query + fmt.Sprintf("] AS %s", u.As), nil
}
