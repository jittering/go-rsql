package rsql

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var tt = ""

var (
	timeFormats = []string{
		time.ANSIC,
		time.UnixDate,
		time.RubyDate,
		time.RFC822,
		time.RFC822Z,
		time.RFC850,
		time.RFC1123,
		time.RFC1123Z,
		time.RFC3339,
		time.RFC3339Nano,
	}
)

var (
	typeOfTime = reflect.TypeOf(time.Time{})
	typeOfByte = reflect.TypeOf([]byte(nil))
)

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
			// support for sql types
			s := v.Type().String()
			i := reflect.New(v.Type()).Interface()
			if scanner, ok := i.(sql.Scanner); ok {
				var err error
				if strings.Contains(s, "NullTime") {
					val, err := parseTime(value)
					if err != nil {
						return nil, err
					}
					err = scanner.Scan(val)
				} else {
					err = scanner.Scan(value)
				}
				if err != nil {
					return nil, err
				}
				return scanner, nil
			}

			return nil, fmt.Errorf("unsupported data type %v", v.Type())
		}
	}

	return v.Interface(), nil
}

func parseTime(value string) (time.Time, error) {
	for _, f := range timeFormats {
		t, err := time.Parse(f, value)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("failed to parse date/time: %s", value)
}
