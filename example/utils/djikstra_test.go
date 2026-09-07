package utils_test

import (
	"testing"

	utils2 "github.com/labkatco/bona/example/__bona_mirror/utils"
)

// {{0:0}, {1:7}, {2:3}, {3:9}, {4:5}}
func TestShortestPath(t *testing.T) {
	CheckDeepEqual(t,
		utils2.ShortestPath(
			5,
			[][]int{
				{0, 1, 10},
				{0, 2, 3},
				{1, 3, 2},
				{2, 1, 4},
				{2, 3, 8},
				{2, 4, 2},
				{3, 4, 5}},
			0,
		),
		map[int]int{0: 0, 1: 7, 2: 3, 3: 9, 4: 5},
	)
}
