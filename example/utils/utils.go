package utils

import (
	"fmt"
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

//avoud:pure
func CalculateTax(amount float64, rate float64) float64 {
	return amount * rate
}

type User struct {
	Name string
}

//avoud:pure
func (u *User) Greet(greeting string) (string, error) {
	return fmt.Sprintf("%s, %s!", greeting, u.Name), nil
}
