package main

import (
	"fmt"

	"github.com/labkatco/bona/example/utils"
)

func main() {
	name := utils.Format(" homie")

	sum := utils.Sum(2, 3)
	fmt.Println(sum, name)

	CalculateTax(100.0, 0.08)

	usr := &User{Name: "Alice"}
	usr.Greet("Hello")
}
