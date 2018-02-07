package core

import (
	"reflect"

	"github.com/go-errors/errors"
)

// extractArguments is an helper function which is used to iterate over
// request arguments and combine its values in array recognized by the engine.
// This format is required by exchange core rpc server.
func extractArguments(s interface{}) ([]interface{}, error) {
	var v reflect.Value
	switch reflect.TypeOf(s).Kind() {
	case reflect.Ptr:
		v = reflect.ValueOf(s).Elem()
	default:
		v = reflect.ValueOf(s)
	}

	switch t := v.Type().Kind(); t {
	case reflect.Struct, reflect.Map:
		var args []interface{}
		for i := 0; i < v.NumField(); i++ {
			f := v.Field(i)
			switch f.Kind() {
			case reflect.Slice:
				for j := 0; j < f.Len(); j++ {
					args = append(args, f.Index(j).Interface())
				}
			default:
				args = append(args, v.Field(i).Interface())
			}
		}
		return args, nil

	case reflect.Array, reflect.Slice:
		args := make([]interface{}, v.Len())

		for i := 0; i < v.Len(); i++ {
			args[i] = v.Index(i).Interface()
		}

		return args, nil
	default:
		return nil, errors.Errorf("unknown type: %v", t)

	}
}
