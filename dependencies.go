package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/google/go-github/github"
	"github.com/ossf/scorecard/v4/checks"
	"github.com/ossf/scorecard/v4/clients"
	"github.com/ossf/scorecard/v4/clients/githubrepo"
	"github.com/ossf/scorecard/v4/clients/githubrepo/roundtripper"
	"github.com/ossf/scorecard/v4/pkg"
)

// Dependency is a raw dependency fetched from the GitHub Dependency Review API.
type dependency struct {
	// Package URL is a short link for a package.
	PackageURL *string `json:"package_url"`

	// SourceRepository is the source repository URL of the dependency.
	SourceRepository *string `json:"source_repository_url"`

	// ChangeType indicates whether the dependency is added, updated, or removed.
	ChangeType *pkg.ChangeType `json:"change_type"`

	// ManifestPath is the path of the manifest file of the dependency, such as go.mod for Go.
	ManifestPath *string `json:"manifest"`

	// Ecosystem is the name of the package management system, such as NPM, GO, PYPI.
	Ecosystem *string `json:"ecosystem"`

	// Version is the package version of the dependency.
	Version *string `json:"version"`

	// Name is the name of the dependency.
	Name string `json:"name"`
}

// fetchRawDependencyDiffData fetches the dependency-diffs between the two code commits
// using the GitHub Dependency Review API, and returns a slice of Dependency.
func fetchRawDependencyDiffData(ctx context.Context, owner, repo, base, head string) ([]pkg.DependencyCheckResult, error) {
	reqPath := url.PathEscape(path.Join(
		"repos", owner, repo, "dependency-graph", "compare", base+"..."+head,
	))
	req, err := http.NewRequestWithContext(ctx, "GET", reqPath, nil)
	if err != nil {
		return nil, fmt.Errorf("request for dependency-diff failed with %w", err)
	}
	ghClient := github.NewClient(
		&http.Client{
			Transport: roundtripper.NewTransport(ctx, nil),
			Timeout:   10 * time.Second,
		},
	)
	deps := []dependency{}
	resp, err := ghClient.Do(ctx, req, &deps)
	if err != nil {
		return nil, fmt.Errorf("error receiving the http reponse: %w with resp status code %v", err, resp.StatusCode)
	}

	ghRepo, err := githubrepo.MakeGithubRepo(path.Join(owner, repo))
	if err != nil {
		return nil, fmt.Errorf("error creating the github repo: %w", err)
	}
	// Initialize the clients to be used in the checks.
	ghRepoClient := githubrepo.CreateGithubRepoClient(ctx, nil)
	ossFuzzRepoClient, err := githubrepo.CreateOssFuzzRepoClient(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating the oss fuzz repo client: %w", err)
	}
	vulnsClient := clients.DefaultVulnerabilitiesClient()
	ciiClient := clients.DefaultCIIBestPracticesClient()

	results := []pkg.DependencyCheckResult{}
	for _, d := range deps {
		// Skip removed dependencies and don't run scorecard checks on them.
		if *d.ChangeType == pkg.Removed {
			continue
		}
		depCheckResult := pkg.DependencyCheckResult{
			PackageURL:       d.PackageURL,
			SourceRepository: d.SourceRepository,
			ChangeType:       d.ChangeType,
			ManifestPath:     d.ManifestPath,
			Ecosystem:        d.Ecosystem,
			Version:          d.Version,
			Name:             d.Name,
		}
		// For now we skip those without source repo urls.
		// TODO: use the BigQuery dataset to supplement null source repo URLs
		// so that we can fetch the Scorecard results for them.
		if d.SourceRepository != nil && *d.SourceRepository != "" {
			// If the srcRepo is valid, run scorecard on this dependency and fetch the result.
			// TODO: use the Scorecare REST API to retrieve the Scorecard result statelessly.
			scorecardResult, err := pkg.RunScorecards(
				ctx,
				ghRepo,
				clients.HeadSHA, /* TODO: In future versions, ideally, this should be commitSHA corresponding to d.Version instead of HEAD. */
				checks.AllChecks,
				ghRepoClient,
				ossFuzzRepoClient,
				ciiClient,
				vulnsClient,
			)
			// "err==nil" suggests the run succeeds, so that we record the scorecard check results for this dependency.
			// Otherwise, it indicates the run fails and we leave the current dependency scorecard result empty
			// rather than letting the entire API return nil since we still expect results for other dependencies.
			if err == nil {
				depCheckResult.ScorecardResults = &scorecardResult
			}
		}
		results = append(results, depCheckResult)
	}
	return results, nil
}
