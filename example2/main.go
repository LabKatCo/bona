package main

import (
	"fmt"

	"github.com/labkatco/bona/example2/utils"
)

func main() {
	sum := utils.Sum(2, 3)
	name := utils.Format(" homie")
	fmt.Println(sum, name)
}
