package templating

import (
	"errors"
	"fmt"
)

var CollectionFuncs = map[string]any{
	"map":      MakeMap,
	"list":     MakeList,
	"reversed": Reversed,
	"append":   AppendList,
}

// MakeMap takes a slice of keys and values and builds a map out of it.
// The keys must be strings.
func MakeMap(kvs ...any) (map[string]any, error) {
	if len(kvs)%2 != 0 {
		return nil, errors.New(
			"must provide an even number of arguments to make key-value pairs",
		)
	}

	result := make(map[string]any, len(kvs)/2)
	for i := 0; i < len(kvs); i += 2 {
		key, ok := kvs[i].(string)
		if !ok {
			return nil, fmt.Errorf(
				"arg index %d should have been a string, got a %T: %+v",
				i, kvs[i], kvs[i],
			)
		}

		value := kvs[i+1]
		result[key] = value
	}
	return result, nil
}

// MakeList takes any number of arguments and returns a slice containing them.
func MakeList(args ...any) []any { return args }

// Reversed takes a slice and returns a copy of it
// with the items in reverse order.
func Reversed(s []any) []any {
	result := make([]any, len(s))
	i := len(s)
	for _, item := range s {
		i--
		result[i] = item
	}
	return result
}

// AppendList returns the result of appending args to s.
func AppendList(s []any, args ...any) []any { return append(s, args...) }
