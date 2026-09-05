package main

import "fmt"

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
