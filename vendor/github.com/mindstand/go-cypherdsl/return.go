package go_cypherdsl

import (
	"errors"
	"fmt"
	"strings"
)

type ReturnPart struct {
	Name              string
	Type              string
	Alias             string
	Function          *FunctionConfig
	Literal           interface{}
	BooleanExpression WhereQuery
	Path              string
}

func (r *ReturnPart) ToString() (string, error) {
	if r.Function != nil {
		query, err := r.Function.ToString()
		if err != nil {
			return "", err
		}

		if r.Alias != "" {
			query += fmt.Sprintf(" AS %s", r.Alias)
		}

		return query, nil
	}

	//handle literal
	if r.Literal != nil {
		return cypherizeInterface(r.Literal)
	}

	//handle boolean expression
	if r.BooleanExpression != "" {
		return string(r.BooleanExpression), nil
	}

	if r.Path != "" {
		return r.Path, nil
	}

	if r.Name == "" {
		return "", errors.New("name can not be empty")
	}

	//handle standard return
	query := r.Name

	if r.Type != "" {
		query += fmt.Sprintf(".%s", r.Type)
	}

	if r.Alias != "" {
		query += fmt.Sprintf(" AS %s", r.Alias)
	}

	return query, nil
}

func NewReturnClause(distinct bool, parts ...ReturnPart) (ReturnQuery, error) {
	if len(parts) == 0 {
		return "", errors.New("parts can not be empty")
	}

	query := "RETURN "

	if distinct {
		query += "DISTINCT "
	}

	for _, part := range parts {
		partStr, err := part.ToString()
		if err != nil {
			return "", err
		}

		query += fmt.Sprintf("%s, ", partStr)
	}

	return ReturnQuery(strings.TrimSuffix(query, ", ")), nil
}
