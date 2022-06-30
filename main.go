package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
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
	directDeps, err := GetDiffFromCommits(args[2], args[0], args[1], args[3], args[4])
	if err != nil {
		fmt.Println(err)
		return
	}
	depsdiff.PrintDependencyChangeInfo(directDeps)

}

func GetDiffFromCommits(authToken, repoOwner string, repoName string,
	base string, head string) ([]Dependency, error) {
	reqURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/dependency-graph/compare/%s...%s",
		repoOwner, repoName, base, head)
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("generate request error: %w", err)

	}
	// To specify the returned type to be JSON so that it's easier to parse.
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	// An access token is required in the request header to be able to use this API.
	req.Header.Set("Authorization", "token "+authToken)

	// Set a ten-seconds timeout to make sure the client can be created correctly.
	myClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := myClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get response error: %w", err)
	}
	defer resp.Body.Close()

	depDiff := []structs.Dependency{}
	err = json.NewDecoder(resp.Body).Decode(&depDiff)
	if err != nil {
		return nil, fmt.Errorf("parse response error: %w", err)
	}
	for i := range depDiff {
		depDiff[i].IsDirect = true
		if depDiff[i].Ecosystem == "pip" {
			depDiff[i].Ecosystem = "pypi"
		}
	}
	return depDiff, nil
}
