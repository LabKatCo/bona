package main_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	utils2 "github.com/labkatco/bona/example/__bona_mirror/utils"
)

func TestMe(t *testing.T) {
	CheckEqual(t, utils2.Format(" fig"), "FIG")
	CheckEqual(t, 1, 1)
	CheckEqual(t, utils2.Sum(1, 2), 3)
}

func TestInsertionSort(t *testing.T) {
	CheckDeepEqual(t, utils2.InsertionSort([]utils2.Pair{}), [][]utils2.Pair{{}})

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
