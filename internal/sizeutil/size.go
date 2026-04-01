package sizeutil

import (
	"fmt"
	"strconv"
	"strings"
)

var multipliers = map[string]int64{
	"":    1,
	"B":   1,
	"K":   1000,
	"KB":  1000,
	"M":   1000 * 1000,
	"MB":  1000 * 1000,
	"G":   1000 * 1000 * 1000,
	"GB":  1000 * 1000 * 1000,
	"KIB": 1024,
	"MIB": 1024 * 1024,
	"GIB": 1024 * 1024 * 1024,
}

func ParseBytes(s string) (int64, error) {
	s = strings.TrimSpace(strings.ToUpper(s))
	if s == "" {
		return 0, fmt.Errorf("size is required")
	}
	i := 0
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	if i == 0 {
		return 0, fmt.Errorf("invalid size %q", s)
	}
	n, err := strconv.ParseInt(s[:i], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size %q", s)
	}
	unit := strings.TrimSpace(s[i:])
	multiplier, ok := multipliers[unit]
	if !ok {
		return 0, fmt.Errorf("unsupported size suffix %q", unit)
	}
	if n <= 0 {
		return 0, fmt.Errorf("size must be greater than zero")
	}
	return n * multiplier, nil
}
