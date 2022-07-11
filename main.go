package main

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/aidenwang9867/DependencyDiffVisualizationInAction/depdiff"

	gogh "github.com/google/go-github/v38/github"
)

func main() {
	// Args[0] is the program path, so use args from Args[1].
	// Args should include:
	// (0) owner name, (1) repo name,
	// (2) GitHub Access Token, (3) base commit SHA, (4) head commit SHA.
	args := os.Args[1:]
	if len(args) != 5 {
		fmt.Println("len of args not equals to 5")
		return
	}

	ctx := depdiff.DepDiffContext{
		OwnerName:   args[0],
		RepoName:    args[1],
		BaseSHA:     args[3],
		HeadSHA:     args[4],
		AccessToken: args[2],
	}

	// Fetch dependency diffs using the GitHub Dependency Review API.
	deps, err := FetchDependencyDiffData(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}
	PrintDependencies(deps)
	// results_0, _ := depdiff.GetDependencyScorecardResults(deps[0])
	// fmt.Println(*results_0)
}

// Get the depednency-diffs between two specified code commits.
func FetchDependencyDiffData(ctx depdiff.DepDiffContext) ([]depdiff.Dependency, error) {
	// Currently, the GitHub Dependency Review
	// (https://docs.github.com/en/rest/dependency-graph/dependency-review) API is used.
	// Set a ten-seconds timeout to make sure the client can be created correctly.
	client := gogh.NewClient(&http.Client{Timeout: 10 * time.Second})
	reqURL := path.Join(
		"repos", ctx.OwnerName, ctx.RepoName, "dependency-graph", "compare",
		ctx.BaseSHA+"..."+ctx.HeadSHA,
	)
	req, err := client.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("request for dependency-diff failed with %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	// An access token is required in the request header to be able to use this API.
	req.Header.Set("Authorization", "token "+ctx.AccessToken)

	depDiff := []depdiff.Dependency{}
	_, err = client.Do(req.Context(), req, &depDiff)
	if err != nil {
		return nil, fmt.Errorf("get response error: %w", err)
	}
	return depDiff, nil
}

func PrintDependencies(deps []depdiff.Dependency) {
	for _, d := range deps {
		fmt.Println(*d.Ecosystem, d.Name, *d.Version, *d.ChangeType)
		if d.PackageURL != nil && *d.PackageURL != "" {
			fmt.Println(*d.PackageURL)
		} else {
			fmt.Println("empty package url")
		}
		if d.SrcRepoURL != nil && *d.SrcRepoURL != "" {
			fmt.Println(*d.SrcRepoURL)
		} else {
			fmt.Println("empty src url")
		}
		fmt.Printf("===================\n\n")
	}
}
