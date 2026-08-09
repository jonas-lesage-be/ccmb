package config

import (
	"os"
	"strings"

	"ccmb/internal/conv"
)

// EnvOr gets an environment variable and converts it to the correct type.
// If the variable is not set, it returns the default value.
func EnvOr[T any](key string, def T) T {
	v := os.Getenv(key)
	if v == "" {
		return def
	}

	val, ok := conv.ConvertValue(v, def)
	if !ok {
		return def
	}
	return val
}

// EnvMapOr gets an environment variable and converts it to a map.
// If the variable is not set, it returns the default value.
func EnvMapOr(envVar, defaultValue string) map[string]bool {
	mapRaw := EnvOr(envVar, defaultValue)
	if mapRaw == "nil" {
		return nil
	}

	mapEntries := make(map[string]bool)
	for ext := range strings.SplitSeq(mapRaw, ",") {
		if trimmed := strings.TrimSpace(ext); trimmed != "" {
			mapEntries[trimmed] = true
		}
	}

	return mapEntries
}
