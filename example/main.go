package main

import (
	"fmt"

	"github.com/labkatco/bona/example/utils"
)

func main() {
	name := utils.Format(" homie")

	sum := utils.Sum(2, 3)
	fmt.Println(sum, name)

	utils.CalculateTax(100.0, 0.08)

	usr := &utils.User{Name: "Alice"}
	usr.Greet("Hello")
}
