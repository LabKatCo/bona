package utils

import (
	"fmt"
	"strings"
)

//bona:pure
func ShortestPath(n int, edges [][]int, src int) map[int]int {
	basePathsToWeights := map[string]int{}

	// for all edges, map start-end path to weight
	for _, edge := range edges {
		start := edge[0]
		end := edge[1]
		weight := edge[2]

		basePathsToWeights[fmt.Sprintf("%v-%v", start, end)] = weight
	}

	pathsToWeights := map[string]int{}
	for k := range basePathsToWeights {
		pathsToWeights[k] = basePathsToWeights[k]
	}

	lastPathCount := len(pathsToWeights)
	pathCount := 0

	for pathCount == 0 || lastPathCount != pathCount {
		lastPathCount = len(pathsToWeights)

		for path, weight := range pathsToWeights {
			pathStart := string(path[0])
			pathEnd := string(path[len(path)-1])

			for possibleNextPath, nextWeight := range basePathsToWeights {
				if strings.HasPrefix(possibleNextPath, pathEnd) && !strings.Contains(possibleNextPath, pathStart) {
					fullPath := path + "-" + string(possibleNextPath[len(possibleNextPath)-1])
					pathsToWeights[fullPath] = weight + nextWeight
				}
			}
		}

		pathCount = len(pathsToWeights)
	}

	shortestPaths := map[int]int{src: 0}

	srcIndex := fmt.Sprintf("%v", src)

	for i := range n {
		nIndex := fmt.Sprintf("%v", i)

		for fullPath, weight := range pathsToWeights {
			if strings.HasPrefix(fullPath, srcIndex) && strings.HasSuffix(fullPath, nIndex) {
				currentShortest, found := shortestPaths[i]

				if !found {
					shortestPaths[i] = weight
				} else if weight < currentShortest {
					shortestPaths[i] = weight
				}
			}
		}

		_, found := shortestPaths[i]
		if !found {
			shortestPaths[i] = -1
		}
	}

	return shortestPaths
}
