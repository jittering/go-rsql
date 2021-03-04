package rsql

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/timtadh/lexmachine"
)

var (
	typeOfTime = reflect.TypeOf(time.Time{})
	typeOfByte = reflect.TypeOf([]byte(nil))
)

// Filter :
type Filter struct {
	Name     string
	Operator Expr
	Value    interface{}
}

func (p *RSQL) parseFilter(values map[string]string, params *Params) error {
	val, ok := values[p.FilterTag]
	if !ok || len(val) < 1 {
		return nil
	}

	scan, err := p.lexer.Scanner([]byte(val))
	if err != nil {
		return err
	}

loop:
	for {
		// lhs - field name
		tkn1, err := nextToken(scan)
		if err != nil {
			if err == io.EOF {
				break loop
			}
			return err
		}

		switch tkn1.Value {
		case "(", ")":
			continue
		}

		var f *StructField
		var ok bool
		if p.codec != nil {
			f, ok = p.codec.Names[tkn1.Value]
			if !ok {
				return fmt.Errorf("invalid field to filter")
			}

			if _, ok := f.Tag.Lookup("filter"); !ok {
				return fmt.Errorf("invalid field to filter")
			}
		}

		name := tkn1.Value
		if f != nil {
			if v, ok := f.Tag.Lookup("column"); ok {
				name = v
			}
		}

		var allows []string
		if f != nil {
			allows = getAllows(f.Type)
			if v, ok := f.Tag.Lookup("allow"); ok {
				allows = strings.Split(v, "|")
			}
		} else {
			allows = allowAll
		}

		// operator
		tkn2, err := nextToken(scan)
		if err != nil {
			return err
		}

		op := operators[tkn2.Value]
		if Strings(allows).IndexOf(op.String()) < 0 {
			return errors.New("operator not support for this field")
		}

		// rhs - value
		tkn3, err := nextToken(scan)
		if err != nil {
			return err
		}

		var value interface{}
		if f != nil {
			v := reflect.New(f.Type).Elem()
			tkn3.Value = strings.Trim(tkn3.Value, `"'`)
			value, err = convertValue(v, tkn3.Value)
			if err != nil {
				return err
			}
		} else {
			// just always use strings, for now
			// postgres and mysql, at least, will try to coerce the types
			// automatically based on the field we are comparing with.
			value = tkn3.Value
		}

		params.Filters = append(params.Filters, &Filter{
			Name:     name,
			Operator: op,
			Value:    value,
		})

		tkn, err := nextToken(scan)
		if err != nil {
			if err == io.EOF {
				break loop
			}
			return err
		}

		switch tkn.Value {
		case ";", ",":
		case "(", ")":
		default:
			return errors.New("unexpected char")
		}
	}

	// for _, f := range params.Filters {
	// 	log.Println("Each :", f, reflect.TypeOf(f.Value), reflect.TypeOf(f))
	// }
	return nil
}

func nextToken(scan *lexmachine.Scanner) (*Token, error) {
	it, err, eof := scan.Next()
	if eof {
		return nil, io.EOF
	}
	if err != nil {
		return nil, err
	}
	return it.(*Token), nil
}

// convertValue string to the correct type for the db field represented by v
func convertValue(v reflect.Value, value string) (interface{}, error) {
	value = strings.TrimSpace(value)

	switch v.Type() {
	case typeOfTime:
		t, err := time.Parse(time.RFC3339, value)
		if err != nil {
			return nil, err
		}
		v.Set(reflect.ValueOf(t))

	case typeOfByte:
		x, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			return nil, err
		}
		v.SetBytes(x)

	default:
		switch v.Kind() {
		case reflect.String:
			v.SetString(value)

		case reflect.Bool:
			x, err := strconv.ParseBool(value)
			if err != nil {
				return nil, err
			}
			v.SetBool(x)

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			x, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return nil, err
			}
			if v.OverflowInt(x) {
				return nil, errors.New("int overflow")
			}
			v.SetInt(x)

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			x, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return nil, err
			}
			if v.OverflowUint(x) {
				return nil, errors.New("unsigned int overflow")
			}
			v.SetUint(x)

		case reflect.Float32, reflect.Float64:
			x, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return nil, err
			}
			if v.OverflowFloat(x) {
				return nil, errors.New("float overflow")
			}
			v.SetFloat(x)

		case reflect.Ptr:
			if value == "null" {
				zero := reflect.Zero(v.Type())
				return zero.Interface(), nil
			}
			return convertValue(v.Elem(), value)

		default:
			return nil, fmt.Errorf("unsupported data type %v", v.Type())
		}
	}

	return v.Interface(), nil
}
