package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
)

const (
	QueryVulnerabilitiesBySystemNameVersion = `
		SELECT
			pv.System, pv.Name, pv.Version,
			adv.Source, adv.SourceID, adv.SourceURL, adv.Title, adv.Description,
			CAST(adv.CVSS3Score AS FLOAT64) AS Score, adv.Severity, adv.GitHubSeverity, adv.Disclosed,
			adv.ReferenceURLs
		FROM (
			SELECT 
				System, Name, Version, SourceID
			FROM
				bigquery-public-data.deps_dev_v1.PackageVersions
			INNER JOIN
				UNNEST(Advisories)
			WHERE
				System = "%s" -- Input 1: dependency system
			AND
				Name = "%s" -- Input 2: dependency name
			AND
				Version = "%s" -- Input 3: dependency version
			AND
				SnapshotAt=(SELECT Time FROM bigquery-public-data.deps_dev_v1.Snapshots ORDER BY Time DESC LIMIT 1)
		) AS pv
		INNER JOIN 
			bigquery-public-data.deps_dev_v1.Advisories AS adv
		ON
		pv.SourceID = adv.SourceID
		WHERE
			SnapshotAt=(SELECT Time FROM bigquery-public-data.deps_dev_v1.Snapshots ORDER BY Time DESC LIMIT 1)
		ORDER BY Score
		DESC
		;
	`

	QueryDependencies = `
		SELECT
			Dependency.System, Dependency.Name, Dependency.Version
		FROM
			bigquery-public-data.deps_dev_v1.Dependencies
		WHERE
			System = "%s" -- Input 1: dependency system
		AND
			Name = "%s" -- Input 2: dependency name
		AND
			Version = "%s" -- Input 3: dependency version
		AND
			SnapshotAt=(SELECT Time FROM bigquery-public-data.deps_dev_v1.Snapshots ORDER BY Time DESC LIMIT 1)
		;
	`

	QueryVulnerabilityByAdvID = `
		SELECT 
			adv.Source, adv.SourceID, adv.SourceURL, adv.Title, adv.Description,
			CAST(adv.CVSS3Score AS FLOAT64) AS Score, adv.Severity, adv.GitHubSeverity, adv.Disclosed,
			adv.ReferenceURLs
		FROM
			bigquery-public-data.deps_dev_v1.Advisories AS adv
		WHERE
			adv.SourceID = "%s"
		AND
			SnapshotAt=(SELECT Time FROM bigquery-public-data.deps_dev_v1.Snapshots ORDER BY Time DESC LIMIT 1)
		;
	`
)

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

// GetVulnerabilityByAdvID now is only used for supplementing the vuln result obtained
// from the GitHub API with vuln data retrieved from BQ.
func GetVulnerabilityByAdvID(advID string) (Vulnerability, error) {
	it, err := DoQueryAndGetRowIterator(
		fmt.Sprintf(
			QueryVulnerabilityByAdvID,
			advID,
		),
	)
	if err != nil {
		return Vulnerability{}, err
	}
	vuln := Vulnerability{}
	err = it.Next(&vuln)
	if err != nil {
		return Vulnerability{}, err
	}
	return vuln, nil
}

func GetVulnerabilitiesBySystemNameVersion(system string, name string, version string) ([]Vulnerability, error) {
	it, err := DoQueryAndGetRowIterator(
		fmt.Sprintf(
			QueryVulnerabilitiesBySystemNameVersion,
			strings.ToUpper(system),
			name,
			version,
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

	depDiff := []Dependency{}
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

func GetDependenciesOfDependencyBySystemNameVersion(system string, name string, version string) ([]Dependency, error) {
	it, err := DoQueryAndGetRowIterator(
		fmt.Sprintf(
			QueryDependencies,
			strings.ToUpper(system),
			name,
			version,
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
