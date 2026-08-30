package main_test

import (
	"testing"

	"github.com/labkatco/bona/example2/utils"

	utils2 "github.com/labkatco/bona/example2/__avoud_mirror/utils"
)

func TestMe(t *testing.T) {
	CheckEqual(t, utils2.Format(" fig"), "FIG")

	CheckEqual(t, 1, 1)
	CheckEqual(t, utils.Sum(1, 2), 3)
	CheckEqual(t, utils2.Sum(1, 2), 33)
}

func CheckEqual[T comparable](t *testing.T, got, want T) {
	t.Helper() // Corrects line-number reporting in terminal
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}
