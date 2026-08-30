package utils

import (
	"strings"
)

//avoud:pure
func Sum(a, b int) int {
	return a + b
}

//avoud:pure
func Format(txt string) string {
	upper := strings.ToUpper(txt)
	return strings.TrimSpace(upper)
}
