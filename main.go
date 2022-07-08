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

	// Get direct dependenies with change types (e.g. added/removed) using the GitHub Dependency Review REST API
	// on the two specified code commits.
	directDeps, err := GetDepDiffByCommitsSHA(args[2], args[0], args[1], args[3], args[4])
	if err != nil {
		fmt.Println(err)
		return
	}
	md := depdiff.SprintDependencyDiffToMarkDown(directDeps)
	fmt.Println(md)

}

// Get the depednency-diff using the GitHub Dependency Review
// (https://docs.github.com/en/rest/dependency-graph/dependency-review) API
func GetDepDiffByCommitsSHA(authToken, repoOwner string, repoName string,
	base string, head string) ([]depdiff.Dependency, error) {
	// Set a ten-seconds timeout to make sure the client can be created correctly.
	client := gogh.NewClient(&http.Client{Timeout: 10 * time.Second})
	reqURL := path.Join(
		"repos", repoOwner, repoName, "dependency-graph", "compare", base+"..."+head,
	)
	req, err := client.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("request for dependency-diff failed with %w", err)

	}
	// To specify the return type to be JSON.
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	// An access token is required in the request header to be able to use this API.
	req.Header.Set("Authorization", "token "+authToken)

	depDiff := []depdiff.Dependency{}
	_, err = client.Do(req.Context(), req, &depDiff)
	if err != nil {
		return nil, fmt.Errorf("get response error: %w", err)
	}
	for i := range depDiff {
		depDiff[i].IsDirect = true
	}
	return depDiff, nil
}
