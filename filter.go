package rsql

import (
	"errors"
	"fmt"
	"io"
	"reflect"
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
				return fmt.Errorf("invalid field to filter: %s", tkn1.Value)
			}

			if _, ok := f.Tag.Lookup("filter"); !ok {
				return fmt.Errorf("invalid field to filter: %s", tkn1.Value)
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
