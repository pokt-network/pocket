package go_cypherdsl

import (
	"errors"
	"fmt"
)

type RemoveConfig struct {
	Name   string
	Field  string
	Labels []string
}

func (r *RemoveConfig) ToString() (string, error) {
	if r.Name == "" {
		return "", errors.New("name must be defined")
	}

	if r.Field == "" && (r.Labels == nil || len(r.Labels) == 0) {
		return "", errors.New("field or labels has to be defined")
	}

	if (r.Labels != nil && len(r.Labels) > 0) && r.Field != "" {
		return "", errors.New("labels and field can not both be defined")
	}

	query := r.Name

	if r.Field != "" {
		return query + fmt.Sprintf(".%s", r.Field), nil
	} else {
		for _, label := range r.Labels {
			query += fmt.Sprintf(":%s", label)
		}

		return query, nil
	}
}
