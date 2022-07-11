package main

import (
	"fmt"
	"net/http"
	"path"
	"time"

	"github.com/aidenwang9867/DependencyDiffVisualizationInAction/depdiff"
	gogh "github.com/google/go-github/v38/github"
)

// Get the depednency-diffs between two specified code commits.
func FetchDependencyDiffData(owner, repo, base, head string) ([]depdiff.Dependency, error) {
	// Currently, the GitHub Dependency Review
	// (https://docs.github.com/en/rest/dependency-graph/dependency-review) API is used.
	// Set a ten-seconds timeout to make sure the client can be created correctly.
	client := gogh.NewClient(&http.Client{Timeout: 10 * time.Second})
	reqURL := path.Join(
		"repos", owner, repo, "dependency-graph", "compare", base+"..."+head,
	)
	req, err := client.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("request for dependency-diff failed with %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	depDiff := []depdiff.Dependency{}
	_, err = client.Do(req.Context(), req, &depDiff)
	if err != nil {
		return nil, fmt.Errorf("get response error: %w", err)
	}
	return depDiff, nil
}
