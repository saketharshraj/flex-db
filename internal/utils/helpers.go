package utils

import (
	"fmt"
	"strings"
)

// ParseInt parses a string into int64, trims quotes, and allows an optional default fallback
func ParseInt(s string, defaultVal ...int64) (int64, error) {
	s = strings.Trim(s, "\"")
	if s == "" {
		if len(defaultVal) > 0 {
			return defaultVal[0], nil
		}
		return 0, fmt.Errorf("cannot parse empty string")
	}

	var val int64
	_, err := fmt.Sscanf(s, "%d", &val)
	if err != nil && len(defaultVal) > 0 {
		return defaultVal[0], err
	}
	return val, err
}
