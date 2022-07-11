package main

import (
	"fmt"
	"os"
)

func main() {
	// Args[0] is the program path, so use args from Args[1].
	// Args should include:
	// (0) owner name, (1) repo name,
	// (2) base commit SHA, (3) head commit SHA.
	args := os.Args[1:]
	if len(args) != 4 {
		fmt.Println("len of args not equals to 4")
		return
	}

	// Fetch dependency diffs using the GitHub Dependency Review API.
	deps, err := FetchDependencyDiffData(args[0], args[1], args[2], args[3])
	if err != nil {
		fmt.Println(err)
		return
	}
	PrintDependencies(deps)
	// results_0, _ := depdiff.GetDependencyScorecardResults(deps[0])
	// fmt.Println(*results_0)
}
