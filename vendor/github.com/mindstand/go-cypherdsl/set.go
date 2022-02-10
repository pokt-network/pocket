package go_cypherdsl

import (
	"errors"
	"fmt"
)

type SetOperation string

const (
	SetEqualTo = "="
	SetMutate  = "+="
)

type SetConfig struct {
	//name is  required, just the var name
	Name string

	//member is used to set the specific member of a node
	Member string

	//defines whether a variable is being set equal to or mutated
	Operation SetOperation

	//used to set the label of a node, (can give a node many labels)
	Label []string

	//if the target is a literal or a reference to another variable
	Target interface{}

	//if the target is a map
	TargetMap *Params

	//if the target is a function
	TargetFunction *FunctionConfig

	//if the set is being done on a condition
	Condition ConditionOperator
}

func (s *SetConfig) ToString() (string, error) {
	if s.Name == "" {
		return "", errors.New("name can not be empty")
	}

	//validate that we aren't trying to do something invalid
	if s.Condition != nil && ((s.Label != nil && len(s.Label) > 0) || s.Operation == SetMutate) {
		return "", errors.New("can not use mutate operator or change labels on a conditional")
	}

	//start query
	query := ""

	//check if its a conditional set
	if s.Condition == nil {
		query = s.Name
	} else {
		str, err := s.Condition.Build()
		if err != nil {
			return "", err
		}

		query = fmt.Sprintf("(CASE WHEN %s THEN %s END)", str, s.Name)
	}

	//if were just setting some labels, we're done
	if (s.Label != nil && len(s.Label) > 0) && s.Operation == "" {
		for _, label := range s.Label {
			query += ":" + label
		}

		return query, nil
	}

	//past the label stuff, an operation must be present
	if s.Operation == "" {
		return "", errors.New("operation must be defined")
	}

	//validate that we only have one type of target set
	c := 0

	if s.TargetFunction != nil {
		c++
	}

	if s.Target != nil {
		c++
	}

	if s.TargetMap != nil {
		c++
	}

	if c != 1 {
		return "", fmt.Errorf("must set exactly one target type, found (%v)", c)
	}

	//handle mutate operation
	if s.Operation == SetMutate {
		if !(s.Label != nil && len(s.Label) > 0) {
			if s.Member == "" {
				if s.TargetMap != nil {
					return query + fmt.Sprintf(" %s %s", s.Operation, s.TargetMap.ToCypherMap()), nil
				} else {
					return "", errors.New("TargetMap must be defined if trying to mutate")
				}
			} else {
				return "", errors.New("member must not be defined if trying to mutate")
			}
		} else {
			return "", errors.New("labels can not be defined for mutate operation")
		}

	}

	//by now we should only be setting stuff
	//already validated that only one target type is set

	//validate target node combo
	if s.Member == "" && (s.Target != nil || s.TargetFunction != nil) {
		//cant do this kind of operation directly on a node
		return "", errors.New("can only set node equal to a map")
	}

	if s.Member != "" {
		query += fmt.Sprintf(".%s", s.Member)
	}

	query += fmt.Sprintf(" %s ", s.Operation)

	if s.Target != nil {
		str, err := cypherizeInterface(s.Target)
		if err != nil {
			return "", err
		}

		return query + str, nil
	} else if s.TargetFunction != nil {
		str, err := s.TargetFunction.ToString()
		if err != nil {
			return "", err
		}

		return query + str, nil
	} else {
		return query + s.TargetMap.ToCypherMap(), nil
	}
}
