package utils

import (
	"fmt"
	"strings"
)

//bona:pure
func Sum(a, b int) int {
	return a + b
}

//bona:pure
func Format(txt string) string {
	upper := strings.ToUpper(txt)
	return strings.TrimSpace(upper)
}

//bona:pure
func CalculateTax(amount float64, rate float64) float64 {
	return amount * rate
}

type User struct {
	Name string
}

//bona:pure
func (u *User) Greet(greeting string) (string, error) {
	return fmt.Sprintf("%s, %s!", greeting, u.Name), nil
}
