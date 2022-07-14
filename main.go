package main

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/aidenwang9867/DependencyDiffVisualizationInAction/pkg"
	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks"
	"github.com/ossf/scorecard/v4/clients"
	"github.com/ossf/scorecard/v4/clients/githubrepo"
	"github.com/ossf/scorecard/v4/log"

	scpkg "github.com/ossf/scorecard/v4/pkg"
)

func main() {
	// Args[0] is the program path, so use args from Args[1].
	// Args should include:
	// (0) owner name, (1) repo name,
	// (2) base commit SHA, (3) head commit SHA.
	args := os.Args[1:]
	if len(args) != 4 {
		fmt.Println("len of args not equals to 4")
		return
	}
	owner, repo, base, head := args[0], args[1], args[2], args[3]
	//Fetch dependency diffs using the GitHub Dependency Review API.
	checksToRun := []string{
		// checks.CheckCodeReview,
		// checks.CheckSAST,
		checks.CheckCIIBestPractices,
		checks.CheckFuzzing,
		checks.CheckVulnerabilities,
	}
	results, err := GetDependencyDiffResults(
		context.Background(), owner, repo, base, head, checksToRun, log.NewLogger(log.InfoLevel),
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(results[0].Name, *results[0].Version)
	fmt.Println(*&results[0].ScorecardResults.Checks)
}

// GetDependencyDiffResults gets dependency changes between two given code commits BASE and HEAD
// along with the Scorecard check results of the dependencies, and returns a slice of DependencyCheckResult.
// TO use this API, an access token must be set following https://github.com/ossf/scorecard#authentication.
func GetDependencyDiffResults(
	ctx context.Context, ownerName, repoName, baseSHA, headSHA string,
	scorecardChecksNames []string, logger *log.Logger) ([]pkg.DependencyCheckResult, error) {
	// Fetch the raw dependency diffs.
	deps, err := fetchRawDependencyDiffData(
		ctx,
		ownerName, repoName, baseSHA, headSHA,
		scorecardChecksNames,
		logger,
	)
	// Initialize the checks to run from the caller's input.
	checksToRun := initScorecardChecks(scorecardChecksNames)
	// Initialize the repo and client(s) corresponding to the checks to run.
	ghRepo, ghRepoClient, ossFuzzClient, vulnsClient, ciiClient, err := initRepoAndClient(
		ownerName, repoName, ctx, logger, checksToRun,
	)
	if err != nil {
		return nil, fmt.Errorf("error initializing repo and client: %w", err)
	}
	results := []pkg.DependencyCheckResult{}
	for _, d := range deps {
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
			if d.ChangeType != nil && *d.ChangeType != pkg.Removed {
				// Run scorecard on those added/updated dependencies with valid srcRepo URLs and fetch the result.
				// TODO: use the Scorecare REST API to retrieve the Scorecard result statelessly.
				scorecardResult, err := scpkg.RunScorecards(
					ctx,
					ghRepo,
					// TODO: In future versions, ideally, this should be
					// the commitSHA corresponding to d.Version instead of HEAD.
					clients.HeadSHA,
					checksToRun,
					ghRepoClient,
					ossFuzzClient,
					ciiClient,
					vulnsClient,
				)
				// If the run fails, we leave the current dependency scorecard result empty
				// rather than letting the entire API return nil since we still expect results for other dependencies.
				if err != nil {
					logger.Error(
						fmt.Errorf("error running scorecard checks: %w", err),
						fmt.Sprintf("The scorecard checks running for dependency %s failed.", d.Name),
					)
				} else { // Otherwise, we record the scorecard check results for this dependency.
					depCheckResult.ScorecardResults = &scorecardResult
				}
			}
		}
		results = append(results, depCheckResult)
		break
	}
	return results, nil
}

func initRepoAndClient(owner, repo string, ctx context.Context, logger *log.Logger, c checker.CheckNameToFnMap) (
	clients.Repo, clients.RepoClient, clients.RepoClient,
	clients.VulnerabilitiesClient, clients.CIIBestPracticesClient, error) {
	// Create the repo and the corresponding client if its check needs to run.
	ghRepo, err := githubrepo.MakeGithubRepo(path.Join(owner, repo))
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("error creating the github repo: %w", err)
	}
	ghRepoClient := githubrepo.CreateGithubRepoClient(ctx, logger)
	// Initialize these three clients as nil at first.
	var ossFuzzClient clients.RepoClient
	var vulnsClient clients.VulnerabilitiesClient
	var ciiClient clients.CIIBestPracticesClient
	for cn := range c {
		switch cn {
		case checks.CheckFuzzing:
			ossFuzzClient, err = githubrepo.CreateOssFuzzRepoClient(ctx, logger)
			if err != nil {
				return nil, nil, nil, nil, nil, fmt.Errorf("error initializing the oss fuzz repo client: %w", err)
			}
		case checks.CheckVulnerabilities:
			vulnsClient = clients.DefaultVulnerabilitiesClient()
		case checks.CheckCIIBestPractices:
			ciiClient = clients.DefaultCIIBestPracticesClient()
		}
	}
	return ghRepo, ghRepoClient, ossFuzzClient, vulnsClient, ciiClient, nil
}

func initScorecardChecks(checkNames []string) checker.CheckNameToFnMap {
	checksToRun := checker.CheckNameToFnMap{}
	if checkNames == nil && len(checkNames) == 0 {
		// If no check names are provided, we run all the checks for the caller.
		checksToRun = checks.AllChecks
	} else {
		for _, cn := range checkNames {
			checksToRun[cn] = checks.AllChecks[cn]
		}
	}
	return checksToRun
}
