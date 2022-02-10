package go_cypherdsl

import (
	"errors"
	"fmt"
	"strings"
)

//with is basically a sort

type WithConfig struct {
	Parts []WithPart
}

func (w *WithConfig) ToString() (string, error) {
	if w.Parts == nil || len(w.Parts) == 0 {
		return "", errors.New("parts can not be empty")
	}

	query := ""

	for _, part := range w.Parts {
		partQuery, err := part.ToString()
		if err != nil {
			return "", err
		}

		query += fmt.Sprintf("%s, ", partQuery)
	}

	return strings.TrimSuffix(query, ", "), nil
}

//todo distinct
type WithPart struct {
	Function *FunctionConfig
	Name     string
	Field    string
	As       string
}

func (wp *WithPart) ToString() (string, error) {
	query := ""
	var err error

	if wp.Function != nil {
		//make sure nothing else is defined
		if wp.Name != "" || wp.Field != "" {
			return "", errors.New("can not define name or field with a function")
		}

		query, err = wp.Function.ToString()
		if err != nil {
			return "", err
		}
	} else if wp.Name != "" {
		query = wp.Name

		if wp.Field != "" {
			query += fmt.Sprintf(".%s", wp.Field)
		}
	} else {
		return "", errors.New("must define a function or name")
	}

	if wp.As != "" {
		query += fmt.Sprintf(" AS %s", wp.As)
	}

	return query, nil
}
