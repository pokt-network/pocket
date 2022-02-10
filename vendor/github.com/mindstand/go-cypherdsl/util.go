package go_cypherdsl

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

func cypherizeInterface(i interface{}) (string, error) {
	if i == nil {
		return "NULL", nil
	}

	//get interface kind
	t := reflect.TypeOf(i)
	k := t.Kind()

	if t == reflect.TypeOf(ParamString("")) {
		s := i.(ParamString)
		return string(s), nil
	}

	//check string
	if k == reflect.String {
		return fmt.Sprintf("'%s'", i.(string)), nil
	}

	//check if primitive numeric type
	if k == reflect.Int || k == reflect.Int8 || k == reflect.Int16 || k == reflect.Int32 || k == reflect.Int64 ||
		k == reflect.Uint || k == reflect.Uint8 || k == reflect.Uint16 || k == reflect.Uint32 || k == reflect.Uint64 ||
		k == reflect.Float32 || k == reflect.Float64 {
		return fmt.Sprintf("%v", i), nil
	}

	//check bool
	if k == reflect.Bool {
		return fmt.Sprintf("%t", i.(bool)), nil
	}

	if k == reflect.Slice {
		q := ""

		is, ok := i.([]interface{})
		if !ok {
			return "", errors.New("kind check failed, this should not have happened")
		}

		for _, iface := range is {
			s, err := cypherizeInterface(iface)
			if err != nil {
				return "", nil
			}

			q += fmt.Sprintf("%s,", s)
		}

		return fmt.Sprintf("[%s]", strings.TrimSuffix(q, ",")), nil
	}

	return "", errors.New("invalid type " + k.String())
}

func RowsToStringArray(data [][]interface{}) ([]string, error) {
	//check to make sure its not empty
	if data == nil || len(data) == 0 || len(data[0]) == 0 {
		return []string{}, nil
	}

	_, ok := data[0][0].(string)
	if !ok {
		return nil, errors.New("does not contain array of strings")
	}

	toReturn := make([]string, len(data))
	for i, v := range data {
		if len(v) == 0 {
			return nil, errors.New("index %v is empty")
		}

		temp, ok := v[0].(string)
		if !ok {
			return nil, errors.New("unable to cast to string")
		}

		toReturn[i] = temp
	}

	return toReturn, nil
}

func RowsTo2dStringArray(data [][]interface{}) ([][]string, error) {
	if len(data) != 0 && len(data[0]) != 0 {
		toReturn := make([][]string, len(data))

		var ok bool

		for i, v := range data {

			toReturn[i] = make([]string, len(v))

			for j, v1 := range v {
				toReturn[i][j], ok = v1.(string)
				if !ok {
					return nil, errors.New("failed to cast value to string")
				}
			}
		}
		return toReturn, nil
	} else {
		return [][]string{}, nil
	}
}
