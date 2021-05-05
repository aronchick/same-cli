package utils

import (
	"fmt"
	"strings"
)

func ContainsString(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func ContainsInt(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func JoinMapKeysValues(s map[string]string) (string, error) {
	values := make([]string, 0, len(s))
	for k, v := range s {
		values = append(values, fmt.Sprintf("%v='%v'", k, v))
	}
	return strings.Join(values, ", "), nil
}
