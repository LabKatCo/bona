package utils_test

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	utils2 "github.com/labkatco/bona/example/__bona_mirror/utils"
	"github.com/labkatco/bona/example/utils"
)

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Bold   = "\033[1m"
)

func TestMe(t *testing.T) {
	// utils2.EnableColor = false
	fmt.Printf("%sHello, %sWorld!%s\n", Red, Bold, Reset)
	fmt.Printf("%sThis is blue.%s\n", Blue, Reset)

	CheckEqual(t, utils2.Format(" fig"), "FIG")
	CheckEqual(t, 1, 1)
	CheckEqual(t, utils2.Sum(1, 2), 3)
}

func TestInsertionSortEmpty(t *testing.T) {
	CheckDeepEqual(t, utils2.InsertionSort([]utils2.Pair{}), [][]utils2.Pair{})
}

func TestInsertionSortFruit(t *testing.T) {
	CheckDeepEqual(t, utils2.InsertionSort([]utils2.Pair{
		{Key: 5, Value: "apple"},
		{Key: 2, Value: "banana"},
		{Key: 9, Value: "cherry"},
	}), [][]utils2.Pair{
		{
			{Key: 5, Value: "apple"},
			{Key: 2, Value: "banana"},
			{Key: 9, Value: "cherry"},
		},
		{
			{Key: 2, Value: "banana"},
			{Key: 5, Value: "apple"},
			{Key: 9, Value: "cherry"},
		},
		{
			{Key: 2, Value: "banana"},
			{Key: 5, Value: "apple"},
			{Key: 9, Value: "cherry"},
		},
	})
}

func TestGeminiInsertionSortFruitNormal(t *testing.T) {
	CheckDeepEqual(t, utils2.GeminiInsertionSort([]utils2.Pair{
		{Key: 5, Value: "apple"},
		{Key: 2, Value: "banana"},
		{Key: 9, Value: "cherry"},
	}), [][]utils2.Pair{
		{
			{Key: 5, Value: "apple"},
			{Key: 2, Value: "banana"},
			{Key: 9, Value: "cherry"},
		},
		{
			{Key: 2, Value: "banana"},
			{Key: 5, Value: "apple"},
			{Key: 9, Value: "cherry"},
		},
		{
			{Key: 2, Value: "banana"},
			{Key: 5, Value: "apple"},
			{Key: 9, Value: "cherry"},
		},
	})
}

func TestInsertionSortFruitNormal(t *testing.T) {
	CheckDeepEqual(t, utils.InsertionSort([]utils.Pair{
		{Key: 5, Value: "apple"},
		{Key: 2, Value: "banana"},
		{Key: 9, Value: "cherry"},
	}), [][]utils.Pair{
		{
			{Key: 5, Value: "apple"},
			{Key: 2, Value: "banana"},
			{Key: 9, Value: "cherry"},
		},
		{
			{Key: 2, Value: "banana"},
			{Key: 5, Value: "apple"},
			{Key: 9, Value: "cherry"},
		},
		{
			{Key: 2, Value: "banana"},
			{Key: 5, Value: "apple"},
			{Key: 9, Value: "cherry"},
		},
	})
}

func TestInsertionSortPets(t *testing.T) {
	CheckDeepEqual(t, utils2.InsertionSort([]utils2.Pair{
		{Key: 3, Value: "cat"},
		{Key: 3, Value: "bird"},
		{Key: 2, Value: "dog"},
	}), [][]utils2.Pair{
		{
			{Key: 3, Value: "cat"},
			{Key: 3, Value: "bird"},
			{Key: 2, Value: "dog"},
		},
		{
			{Key: 3, Value: "cat"},
			{Key: 3, Value: "bird"},
			{Key: 2, Value: "dog"},
		},
		{
			{Key: 2, Value: "dog"},
			{Key: 3, Value: "cat"},
			{Key: 3, Value: "bird"},
		},
	})
}

func TestInsertionSortOther(t *testing.T) {
	CheckDeepEqual(t, utils2.InsertionSort([]utils2.Pair{
		{Key: 3, Value: "cat"},
		{Key: 3, Value: "bird"},
		{Key: 3, Value: "dog"},
	}), [][]utils2.Pair{
		{
			{Key: 3, Value: "cat"},
			{Key: 3, Value: "bird"},
			{Key: 3, Value: "dog"},
		},
		{
			{Key: 3, Value: "cat"},
			{Key: 3, Value: "bird"},
			{Key: 3, Value: "dog"},
		},
		{
			{Key: 3, Value: "cat"},
			{Key: 3, Value: "bird"},
			{Key: 3, Value: "dog"},
		},
	})

	CheckDeepEqual(t, utils2.InsertionSort([]utils2.Pair{
		{Key: 8, Value: "cat"},
		{Key: 7, Value: "bird"},
		{Key: 8, Value: "dog"},
	}), [][]utils2.Pair{
		{
			{Key: 8, Value: "cat"},
			{Key: 7, Value: "bird"},
			{Key: 8, Value: "dog"},
		},
		{
			{Key: 7, Value: "bird"},
			{Key: 8, Value: "cat"},
			{Key: 8, Value: "dog"},
		},
		{
			{Key: 7, Value: "bird"},
			{Key: 8, Value: "cat"},
			{Key: 8, Value: "dog"},
		},
	})
}

func CheckEqual[T comparable](t *testing.T, got, want T) {
	t.Helper() // Corrects line-number reporting in terminal
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func CheckDeepEqual[T any](t *testing.T, got, want T) {
	t.Helper()
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}
}
