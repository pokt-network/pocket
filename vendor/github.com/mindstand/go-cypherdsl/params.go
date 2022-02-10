package go_cypherdsl

import (
	"errors"
	"fmt"
	"strings"
)

type Params struct {
	params map[string]string
}

func ParamsFromMap(m map[string]interface{}) (*Params, error) {
	if m == nil || len(m) == 0 {
		return nil, errors.New("map can not be empty or nil")
	}

	p := &Params{}

	for k, v := range m {
		err := p.Set(k, v)
		if err != nil {
			return nil, err
		}
	}

	return p, nil
}

func (p *Params) IsEmpty() bool {
	return p.params == nil || len(p.params) == 0
}

func (p *Params) Set(key string, value interface{}) error {
	if p.params == nil {
		p.params = map[string]string{}
	}

	str, err := cypherizeInterface(value)
	if err != nil {
		return err
	}

	p.params[key] = fmt.Sprintf("%s:%s", key, str)

	return nil
}

func (p *Params) ToCypherMap() string {

	if p.params == nil || len(p.params) == 0 {
		return "{}"
	}

	q := ""

	for _, v := range p.params {
		q += fmt.Sprintf("%s,", v)
	}

	return "{" + strings.TrimSuffix(q, ",") + "}"
}
