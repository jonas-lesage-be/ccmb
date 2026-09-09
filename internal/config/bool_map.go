package config

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-viper/mapstructure/v2"
)

var errInvalidBoolMapType = errors.New("expected string for bool map")

func toBoolMapHook() mapstructure.DecodeHookFuncType {
	target := reflect.TypeFor[map[string]bool]()

	return func(from, to reflect.Type, data any) (any, error) {
		if to != target {
			return data, nil
		}

		switch from.Kind() {
		case reflect.Map:
			return toBoolMap(data), nil
		case reflect.Slice, reflect.Array:
			return sliceToBoolMap(data), nil
		case reflect.String:
			return stringToBoolMap(data)
		default:
			return data, nil
		}
	}
}

func toBoolMap(data any) map[string]bool {
	v := reflect.ValueOf(data)
	if v.Len() == 0 {
		return nil
	}

	res := make(map[string]bool, v.Len())
	for _, key := range v.MapKeys() {
		keyStr := fmt.Sprintf("%v", key.Interface())
		val := v.MapIndex(key).Interface()

		if b, ok := val.(bool); ok {
			res[keyStr] = b
			continue
		}
		res[keyStr] = true
	}

	if len(res) == 0 {
		return nil
	}

	return res
}

func sliceToBoolMap(data any) map[string]bool {
	v := reflect.ValueOf(data)
	if v.Len() == 0 {
		return nil
	}

	var res map[string]bool

	for i := range v.Len() {
		val := v.Index(i).Interface()
		if m, ok := val.(map[string]any); ok {
			if res == nil {
				res = make(map[string]bool)
			}

			mergeRawMap(res, m)
			continue
		}

		keyStr := strings.TrimSpace(fmt.Sprintf("%v", val))
		if keyStr != "" {
			if res == nil {
				res = make(map[string]bool)
			}

			res[keyStr] = true
		}
	}
	return res
}

func mergeRawMap(dest map[string]bool, src map[string]any) {
	for key, val := range src {
		if b, ok := val.(bool); ok {
			dest[key] = b
			continue
		}
		dest[key] = true
	}
}

func stringToBoolMap(data any) (map[string]bool, error) {
	str, ok := data.(string)
	if !ok {
		return nil, fmt.Errorf("%w: got %T", errInvalidBoolMapType, data)
	}

	trimmed := strings.TrimSpace(str)
	if trimmed == "" || trimmed == "null" {
		return map[string]bool(nil), nil
	}

	return parseBoolMap(trimmed), nil
}

func parseBoolMap(value string) map[string]bool {
	if value == "nil" {
		return nil
	}

	set := make(map[string]bool)
	for entry := range strings.SplitSeq(value, ",") {
		if entry = strings.TrimSpace(entry); entry != "" {
			set[entry] = true
		}
	}

	return set
}
