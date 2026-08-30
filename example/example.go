package main

import "fmt"

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

func main() {
	CalculateTax(100.0, 0.08)

	usr := &User{Name: "Alice"}
	usr.Greet("Hello")
}
