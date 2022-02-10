package go_cypherdsl

import (
	"errors"
	"fmt"
	"strings"
)

func NewNode(builder *PathBuilder) (CreateQuery, error) {
	if builder == nil {
		return "", errors.New("builder can not be nil")
	}

	query, err := builder.ToCypher()
	if err != nil {
		return "", err
	}

	if query == "" {
		return "", errors.New("query can not be empty")
	}

	return CreateQuery(query), nil
}

type IndexConfig struct {
	Type   string
	Fields []string
}

func NewIndex(index *IndexConfig) (CreateQuery, error) {
	if index == nil {
		return "", errors.New("index can not be nil")
	}

	if index.Type == "" {
		return "", errors.New("type can not be empty")
	}

	if index.Fields == nil {
		return "", errors.New("fields can not be nil")
	}

	if len(index.Fields) == 0 {
		return "", errors.New("fields can not be empty")
	}

	query := fmt.Sprintf("INDEX ON :%s(", index.Type)

	for _, field := range index.Fields {
		query += fmt.Sprintf("%s,", field)
	}

	return CreateQuery(strings.TrimSuffix(query, ",") + ")"), nil
}

type ConstraintConfig struct {
	//specify the name of the variable for the constraint
	Name string

	//specify the type the action takes place on
	Type string

	//specify the field the action takes place on
	Field string

	//require field to be unique
	Unique bool

	//require field to show up
	Exists bool
}

func NewConstraint(constraint *ConstraintConfig) (CreateQuery, error) {
	if constraint == nil {
		return "", errors.New("constraint can not be nil")
	}

	if constraint.Name == "" || constraint.Type == "" || constraint.Field == "" {
		return "", errors.New("name, type and field can not be empty")
	}

	if constraint.Unique == constraint.Exists || (!constraint.Unique && !constraint.Exists) {
		return "", errors.New("can only be unique or exists per call")
	}

	root := fmt.Sprintf("CONSTRAINT ON (%s:%s) ASSERT ", constraint.Name, constraint.Type)

	if constraint.Unique {
		root += fmt.Sprintf("%s.%s IS UNIQUE", constraint.Name, constraint.Field)
	} else {
		root += fmt.Sprintf("exists(%s.%s)", constraint.Name, constraint.Field)
	}

	return CreateQuery(root), nil
}
