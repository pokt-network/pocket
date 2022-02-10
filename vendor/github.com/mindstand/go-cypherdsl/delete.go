package go_cypherdsl

import (
	"errors"
	"fmt"
	"strings"
)

//no config obj needed because this is really simple
func deleteToString(detach bool, params ...string) (string, error) {
	if len(params) == 0 {
		return "", errors.New("params can not be empty")
	}

	query := "DELETE"

	if detach {
		query = "DETACH " + query
	}

	for _, v := range params {
		query += fmt.Sprintf(" %s,", v)
	}

	return strings.TrimSuffix(query, ","), nil
}
