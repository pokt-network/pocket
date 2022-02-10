package go_cypherdsl

import (
	"errors"
	"strings"
)

type CreateQuery string

func (c *CreateQuery) ToString() string {
	return string(*c)
}

type WhereQuery string

func (c *WhereQuery) ToString() string {
	return string(*c)
}

type MergeQuery string

func (c *MergeQuery) ToString() string {
	return string(*c)
}

type ReturnQuery string

func (c *ReturnQuery) ToString() string {
	return string(*c)
}

type DeleteQuery string

func (c *DeleteQuery) ToString() string {
	return string(*c)
}

type SetQuery string

func (c *SetQuery) ToString() string {
	return string(*c)
}

type RemoveQuery string

func (c *RemoveQuery) ToString() string {
	return string(*c)
}

type ParamString string

func (p *ParamString) ToString() string {
	return string(*p)
}

type FunctionConfig struct {
	Name   string
	Params []interface{}
}

func (f *FunctionConfig) ToString() (string, error) {
	if f.Name == "" {
		return "", errors.New("name can not be nil")
	}

	fu := f.Name + "("

	if f.Params != nil && len(f.Params) != 0 {
		for i := range f.Params {
			str, err := cypherizeInterface(f.Params[i])
			if err != nil {
				return "", err
			}

			fu += str + ","
		}

		fu = strings.TrimSuffix(fu, ",")
	}

	return fu + ")", nil
}
