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
		DISTINCT
			t1.System, t1.Name, t1.Version,
			t2.Source, t2.SourceID, t2.SourceURL, t2.Title, t2.Description,
			CAST(t2.CVSS3Score AS FLOAT64) AS Score, t2.Severity, t2.GitHubSeverity, t2.Disclosed,
			-- t2.ReferenceURLs

		FROM (
				SELECT DISTINCT pv.System, pv.Name, pv.Version, pv_adv.SourceID
				FROM ` +
		"`bigquery-public-data.deps_dev_v1.PackageVersions`" + `AS pv
				INNER JOIN 
					UNNEST(pv.Advisories) AS pv_adv
				WHERE 
					pv.System = "%s" -- Input 1: dependency system
				AND
					pv.Name = "%s"  -- Input 2: dependency name
				AND
					pv.Version = "%s"  -- Input 3: dependency version
			) AS t1

				INNER JOIN` +
		"`bigquery-public-data.deps_dev_v1.Advisories`" + `
				AS t2
				ON
					t1.SourceID = t2.SourceID

		ORDER BY Score
		DESC
	`

	QueryDependencies = `
		SELECT
		DISTINCT
			d.System, d.Name, d.Version
		FROM (
			SELECT
			dep.Dependency AS d
			FROM ` +
		"`bigquery-public-data.deps_dev_v1.Dependencies`" + ` AS dep
			WHERE
			dep.System = "%s" -- Input 1: dependency system
			AND
			dep.Name = "%s" -- Input 2: dependency name
			AND
			dep.Version = "%s" -- Input 3: dependency version
		)
	`

	QueryVulnerabilityByAdvID = `
		SELECT 
			adv.Source, adv.SourceID, adv.SourceURL, adv.Title, adv.Description,
			CAST(adv.CVSS3Score AS FLOAT64) AS Score, adv.Severity, adv.GitHubSeverity, adv.Disclosed,
			adv.ReferenceURLs
		FROM ` +
		"`bigquery-public-data.deps_dev_v1.Advisories`" + ` AS adv
		WHERE
		adv.SourceID = "%s"
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
