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
	fullSeq := [][]Pair{}

	sortedPairs := []Pair{}

	for i := range pairs {
		currentPair := pairs[i]

		insertSortedIndex := 0
		for insertSortedIndex < len(sortedPairs) {
			if currentPair.Key < sortedPairs[insertSortedIndex].Key {
				break
			}

			insertSortedIndex++
		}

		sortedBeforeCurrent := append([]Pair{}, sortedPairs[:insertSortedIndex]...)
		sortedAfterCurrent := append([]Pair{}, sortedPairs[insertSortedIndex:]...)

		newFirstSortedPairs := append(sortedBeforeCurrent, currentPair)

		newSortedPairs := append(newFirstSortedPairs, sortedAfterCurrent...)

		sortedPairs = newSortedPairs

		restPairs := pairs[i+1:]
		allCurrentPairs := append(newSortedPairs, restPairs...)

		fullSeq = append(fullSeq, allCurrentPairs)
	}

	return fullSeq
}

//bona:pure
func GeminiInsertionSort(pairs []Pair) [][]Pair {
	n := len(pairs)
	if n == 0 {
		return [][]Pair{}
	}

	states := make([][]Pair, 0, n)

	// Make a deep copy of the slice to avoid modifying the input directly.
	arr := make([]Pair, n)
	copy(arr, pairs)

	for i := 0; i < n; i++ {
		keyItem := arr[i]
		j := i - 1

		// Shift elements of arr[0..i-1] that are greater than keyItem.Key
		// to one position ahead of their current position.
		// Using > instead of >= ensures stability for equal keys.
		for j >= 0 && arr[j].Key > keyItem.Key {
			arr[j+1] = arr[j]
			j--
		}
		arr[j+1] = keyItem

		// Save a deep copy of the current state after placing element `i`.
		stateCopy := make([]Pair, n)
		copy(stateCopy, arr)
		states = append(states, stateCopy)
	}

	return states
}
