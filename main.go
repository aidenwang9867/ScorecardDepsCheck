package main

import (
	"fmt"
	"net/http"
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
	ownerName := "ossf"
	repoName := "allstar"
	pat := "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"
	base := "c1e3869cc9ae93775549d5583994ba7f89761964"
	head := "65db129afcbb986e0e3ec4156333892a67594d76"

	// Get direct dependenies with change types (e.g. added/removed) using the GitHub Dependency Review REST API
	// on the two specified code commits.
	deps, err := GetDepDiffByCommitsSHA(pat, ownerName, repoName, base, head)
	if err != nil {
		fmt.Println(err)
		return
	}
	md := depdiff.SprintDependencyDiffToMarkDown(deps)
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
