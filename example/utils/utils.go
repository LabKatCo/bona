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

type Pair struct {
	Key   int
	Value string
}

//bona:pure
func InsertionSort(pairs []Pair) [][]Pair {
	orderSeq := [][]Pair{pairs}

	for i := range len(pairs) - 1 {
		pairToInsert := pairs[i]

		otherPairs := []Pair{}

		for _, pair := range orderSeq[len(orderSeq)-1] {
			if pair.Key == pairToInsert.Key && pair.Value == pairToInsert.Value {
				continue
			}

			otherPairs = append(otherPairs, pair)
		}

		afterPairs := []Pair{}
		inserted := false

		for _, otherPair := range otherPairs {
			if pairToInsert.Key <= otherPair.Key {
				if !inserted {
					afterPairs = append(afterPairs, pairToInsert)
					inserted = true
				}
			}

			afterPairs = append(afterPairs, otherPair)
		}

		orderSeq = append(orderSeq, afterPairs)
	}

	return orderSeq
}
