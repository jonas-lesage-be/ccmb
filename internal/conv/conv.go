package conv

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var (
	// ErrEmptyString is returned when an empty string is provided for conversion.
	ErrEmptyString = errors.New("empty string")

	// ErrUnknownUnit is returned when an unknown size unit is encountered.
	ErrUnknownUnit = errors.New("unknown size unit")

	// ErrUnsupportedType is returned when a type is not supported for conversion.
	ErrUnsupportedType = errors.New("unsupported type for conversion")
)

// ConvertValue converts a string value to the specified type T,
// using the provided converter for unit conversions.
// If the conversion fails or the value is empty,
// it returns the provided default value.
func ConvertValue[T any](v string, def T) (T, bool) {
	t := reflect.TypeOf(def)
	if t == nil {
		return def, false
	}

	kind := t.Kind()

	if kind == reflect.String {
		if val, ok := reflect.ValueOf(v).Convert(t).Interface().(T); ok {
			return val, true
		}
		return def, false
	}

	if kind < reflect.Bool || kind > reflect.Complex128 {
		return parseComplexType(v, t, def)
	}

	resPtr := reflect.New(t)
	if err := parsePrimitive(v, t, kind, resPtr); err != nil {
		return def, false
	}

	if val, ok := resPtr.Elem().Interface().(T); ok {
		return val, true
	}
	return def, false
}

func parsePrimitive(
	v string,
	t reflect.Type,
	kind reflect.Kind,
	resPtr reflect.Value,
) error {
	if t == reflect.TypeFor[time.Duration]() {
		return parseDuration(v, resPtr)
	}

	var err error

	switch {
	case kind == reflect.Bool:
		err = parseBool(v, resPtr)
	case kind >= reflect.Int && kind <= reflect.Int64:
		err = parseInt(v, bitsOrZero(t, kind), resPtr)
	case kind >= reflect.Uint && kind <= reflect.Uintptr:
		err = parseUint(v, bitsOrZero(t, kind), resPtr)
	default:
		err = parseFloating(v, t, kind, resPtr)
	}

	if err != nil {
		return fmt.Errorf("%w (type: %s)", err, kind.String())
	}
	return nil
}

func parseDuration(v string, resPtr reflect.Value) error {
	d, err := time.ParseDuration(v)
	if err != nil {
		return fmt.Errorf("failed to parse duration: %w", err)
	}

	resPtr.Elem().Set(reflect.ValueOf(d))
	return nil
}

func parseBool(v string, resPtr reflect.Value) error {
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fmt.Errorf("failed to parse boolean: %w", err)
	}

	resPtr.Elem().SetBool(b)
	return nil
}

func parseInt(v string, bits int, resPtr reflect.Value) error {
	v = strings.TrimSpace(v)
	if v == "" {
		return ErrEmptyString
	}

	unit := strings.TrimLeftFunc(v, func(r rune) bool {
		return r >= '0' && r <= '9' || r == '-' || r == '+'
	})

	if unit = strings.TrimSpace(strings.ToLower(unit)); unit != "" {
		numStr := v[:len(v)-len(unit)]
		numStr = strings.TrimSpace(numStr)

		num, err := strconv.ParseInt(numStr, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse size number: %w", err)
		}

		mult, err := multiplier(unit)
		if err != nil {
			return fmt.Errorf("failed to parse size unit: %w", err)
		}
		resPtr.Elem().SetInt(num * mult)

		return nil
	}

	i, err := strconv.ParseInt(v, 10, bits)
	if err != nil {
		return fmt.Errorf("failed to parse integer: %w", err)
	}

	resPtr.Elem().SetInt(i)
	return nil
}

func parseUint(v string, bits int, resPtr reflect.Value) error {
	u, err := strconv.ParseUint(v, 10, bits)
	if err != nil {
		return fmt.Errorf("failed to parse unsigned integer: %w", err)
	}

	resPtr.Elem().SetUint(u)
	return nil
}

func parseFloating(v string, t reflect.Type, kind reflect.Kind, resPtr reflect.Value) error {
	if kind == reflect.Float32 || kind == reflect.Float64 {
		f, err := strconv.ParseFloat(v, bitsOrZero(t, kind))
		if err != nil {
			return fmt.Errorf("failed to parse float: %w", err)
		}

		resPtr.Elem().SetFloat(f)
		return nil
	}

	if kind == reflect.Complex64 || kind == reflect.Complex128 {
		c, err := strconv.ParseComplex(v, bitsOrZero(t, kind))
		if err != nil {
			return fmt.Errorf("failed to parse complex number: %w", err)
		}

		resPtr.Elem().SetComplex(c)
		return nil
	}

	return ErrUnsupportedType
}

func bitsOrZero(t reflect.Type, kind reflect.Kind) int {
	if kind == reflect.Int || kind == reflect.Uint || kind == reflect.Uintptr {
		return 0
	}
	return t.Bits()
}

func parseComplexType[T any](v string, t reflect.Type, def T) (T, bool) {
	newPtr := reflect.New(t)
	if err := json.Unmarshal([]byte(v), newPtr.Interface()); err != nil {
		return def, false
	}

	if val, ok := newPtr.Elem().Interface().(T); ok {
		return val, true
	}
	return def, false
}

func multiplier(unit string) (int64, error) {
	switch unit {
	case "b":
		return B, nil
	case "kb":
		return KB, nil
	case "mb":
		return MB, nil
	case "gb":
		return GB, nil
	case "tb":
		return TB, nil
	case "pb":
		return PB, nil
	case "eb":
		return EB, nil
	default:
		return binaryMultiplier(unit)
	}
}

func binaryMultiplier(unit string) (int64, error) {
	switch unit {
	case "kib":
		return KiB, nil
	case "mib":
		return MiB, nil
	case "gib":
		return GiB, nil
	case "tib":
		return TiB, nil
	case "pib":
		return PiB, nil
	case "eib":
		return EiB, nil
	default:
		return 0, fmt.Errorf("%w: %s", ErrUnknownUnit, unit)
	}
}
