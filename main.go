package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aidenwang9867/DependencyDiffVisualizationInAction/depdiff"
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
	owner, repo, base, head := args[0], args[1], args[2], args[3]
	//Fetch dependency diffs using the GitHub Dependency Review API.
	results, err := GetDependencyDiff(owner, repo, base, head)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(*results[0].ScorecardResults)
}

func GetDependencyDiff(owner, repo, base, head string) ([]depdiff.DependencyCheckResult, error) {
	ctx := context.Background()
	// Fetch dependency diffs using the GitHub Dependency Review API.
	deps, err := depdiff.FetchRawDependencyDiffData(ctx, owner, repo, base, head)
	if err != nil {
		return nil, fmt.Errorf("error fetching dependency changes: %w", err)
	}
	return deps, nil
}
