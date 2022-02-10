package go_cypherdsl

import (
	"errors"
	"fmt"
)

type MergeConfig struct {
	//the path its merging on
	Path string

	//what it does if its creating the node
	OnCreate *MergeSetConfig

	//what it does if its matching the node
	OnMatch *MergeSetConfig
}

func (m *MergeConfig) ToString() (string, error) {
	if m.Path == "" {
		return "", errors.New("path can not be empty")
	}

	query := m.Path

	if m.OnCreate != nil {
		str, err := m.OnCreate.ToString()
		if err != nil {
			return "", err
		}

		query += fmt.Sprintf(" ON CREATE SET %s", str)
	}

	if m.OnMatch != nil {
		str, err := m.OnMatch.ToString()
		if err != nil {
			return "", err
		}

		query += fmt.Sprintf(" ON MATCH SET %s", str)
	}

	return query, nil
}

type MergeSetConfig struct {
	//variable name
	Name string

	//member variable of node
	Member string

	//new value
	Target interface{}

	//new value if its a function, do not include
	TargetFunction *FunctionConfig
}

func (m *MergeSetConfig) ToString() (string, error) {
	if m.Name == "" {
		return "", errors.New("name can not be empty")
	}

	if m.Member == "" {
		return "", errors.New("member can not be empty")
	}

	if m.Target == nil && m.TargetFunction == nil {
		return "", errors.New("target or target function has to be defined")
	}

	if m.Target != nil && m.TargetFunction != nil {
		return "", errors.New("target and target function can not both be defined")
	}

	query := fmt.Sprintf("%s.%s = ", m.Name, m.Member)

	if m.Target != nil {
		str, err := cypherizeInterface(m.Target)
		if err != nil {
			return "", err
		}

		return query + str, nil
	} else {
		str, err := m.TargetFunction.ToString()
		if err != nil {
			return "", err
		}

		return query + str, nil
	}
}
