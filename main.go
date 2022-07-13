package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aidenwang9867/DependencyDiffVisualizationInAction/pkg"
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
	results, err := GetDependencyDiffResults(owner, repo, base, head)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(results[0].Name, *results[0].Version)
	fmt.Println(*results[0].ScorecardResults)
}

func GetDependencyDiffResults(ownerName, repoName, baseSHA, headSHA string) ([]pkg.DependencyCheckResult, error) {
	ctx := context.Background()
	// Fetch dependency diffs using the GitHub Dependency Review API.
	return fetchRawDependencyDiffData(ctx, ownerName, repoName, baseSHA, headSHA)
}
