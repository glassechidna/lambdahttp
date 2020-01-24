package secretenv

import (
	"fmt"
	"strings"
)

func EnvMap(environ []string) map[string]string {
	envMap := make(map[string]string, len(environ))

	for _, keyval := range environ {
		parts := strings.SplitN(keyval, "=", 2)
		key := parts[0]
		val := parts[1]
		envMap[key] = val
	}

	return envMap
}

func EnvSlice(envMap map[string]string) []string {
	envSlice := make([]string, 0, len(envMap))

	for key, val := range envMap {
		envSlice = append(envSlice, fmt.Sprintf("%s=%s", key, val))
	}

	return envSlice
}

func keysWithPrefixedValue(envMap map[string]string, prefix string) []string {
	keys := []string{}

	for key, value := range envMap {
		if strings.HasPrefix(value, prefix) {
			keys = append(keys, key)
		}
	}

	return keys
}
