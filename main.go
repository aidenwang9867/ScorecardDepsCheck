package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
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

	// Get direct dependeny changes using the GitHub Dependency Review REST API
	directDeps, err := GetDiffFromCommits(args[2], args[0], args[1], args[3], args[4])
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(len(directDeps))
	// Get dependencies of direct dependencies, i.e. indirect dependencies.
	for i, d := range directDeps {
		indirectDeps, err := GetDependenciesOfDependency(d)
		if err != nil {
			fmt.Println(err)
			return
		}
		directDeps[i].Dependencies = indirectDeps
		fmt.Println(indirectDeps)
	}
	// Retrieve vulnerabilities of all dependencies.
	// Since we only have (1) direct dependencies and (2) one layer of indirect dependencies,
	// we only use two iterations here to traverse over all nodes.
	// This might be a graph traversal in the future if more indirect dependency layers are added.
	for i, d := range directDeps {
		if *d.ChangeType == "removed" {
			continue
		}
		vuln, err := GetVulnerabilitiesOfDependency(d)
		if err != nil {
			fmt.Println(err)
			return
		}
		directDeps[i].Vulnerabilities = append(
			directDeps[i].Vulnerabilities,
			vuln...,
		)
		// PrintDependencyToStdOut(d)

		// Handle vulnerabilities in indirect dependencies.
		for j, dd := range directDeps[i].Dependencies {
			v, err := GetVulnerabilitiesOfDependency(dd)
			if err != nil {
				fmt.Println(err)
				return
			}
			directDeps[i].Dependencies[j].Vulnerabilities = append(
				directDeps[i].Dependencies[j].Vulnerabilities,
				v...,
			)
			// PrintDependencyToStdOut(dd)
			if len(directDeps[i].Dependencies[j].Vulnerabilities) != 0 {
				fmt.Println("Vulnerbaility found in indirect dependencies")
			}
		}

		if len(directDeps[i].Vulnerabilities) != 0 {
			fmt.Println("Vulnerbaility found in direct dependencies")
		}
	}
}

func DoQueryAndGetRowIterator(queryStr string) (*bigquery.RowIterator, error) {
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
	return it, nil
}

func GetVulnerabilitiesOfDependency(d Dependency) ([]Vulnerability, error) {
	it, err := DoQueryAndGetRowIterator(
		fmt.Sprintf(
			QueryGetVulnerabilities,
			strings.ToUpper(d.Ecosystem),
			d.Name,
			d.Version,
		),
	)
	if err != nil {
		return nil, err
	}
	vuln := []Vulnerability{}
	for {
		row := Vulnerability{}
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		vuln = append(vuln, row)
	}
	return vuln, nil
}

func GetDiffFromCommits(authToken, repoOwner string, repoName string,
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
	return depDiff, nil
}

func GetDependenciesOfDependency(d Dependency) ([]Dependency, error) {
	// A nit for the name of the Python package manager?
	if d.Ecosystem == "pip" {
		d.Ecosystem = "pypi"
	}
	it, err := DoQueryAndGetRowIterator(
		fmt.Sprintf(
			QueryGetDependencies,
			strings.ToUpper(d.Ecosystem),
			d.Name,
			d.Version,
		),
	)
	if err != nil {
		return nil, err
	}
	dep := []Dependency{}
	for {
		row := Dependency{}
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		dep = append(dep, row)
	}
	return dep, nil
}
