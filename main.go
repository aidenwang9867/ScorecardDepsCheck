package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
)

func main() {
	args := os.Args[1:] // Args[0] is the program path.
	// Args should include: (0) owner name, (1) repo name,
	// (2) GH PAT, (3) base commit SHA, (4) head commit SHA, .
	if len(args) != 5 {
		fmt.Println("len of args not equals to 5")
		return
	}
	GetGitHubDepReviewResult(args[2], args[0], args[1], args[3], args[4])
	records, e := DoQuery(fmt.Sprintf(QueryGetVulnerabilities, "tensorflow"))
	if e != nil {
		fmt.Println(e)
	}
	fmt.Println(records)
}

func DoQuery(queryStr string) ([][]bigquery.Value, error) {
	ctx := context.Background()
	c, err := bigquery.NewClient(ctx, "ossf-malware-analysis")
	if err != nil {
		return nil, fmt.Errorf("%w when creating the big query client", err)
	}
	q := c.Query(queryStr)
	// Execute the query.
	it, err := q.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w when reading the context", err)
	}
	records := [][]bigquery.Value{}
	// Iterate through the results.
	for {
		row := []bigquery.Value{}
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("%w when iterating the query results", err)
		}
		records = append(records, row)
	}
	return records, nil
}

func GetVulnerabilitiesAndParse(deps []Dependency) ([]string, error) {

	return nil, nil
}

func GetGitHubDepReviewResult(authToken, repoOwner string, repoName string,
	base string, head string) ([]Dependency, error) {
	reqURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/dependency-graph/compare/%s...%s",
		repoOwner, repoName, base, head)
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("generate request error: %w", err)

	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Authorization", "token "+authToken)

	myClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := myClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get response error: %w", err)
	}
	defer resp.Body.Close()

	depDiff := []Dependency{}
	err = json.NewDecoder(resp.Body).Decode(&depDiff)
	if err != nil {
		return nil, fmt.Errorf("parse response error: %w", err)
	}
	PrintDepDiffToStdOut(depDiff)
	return depDiff, nil
}

func IsVersionAffected(version string, affectedVersions string) bool {
	return false
}
