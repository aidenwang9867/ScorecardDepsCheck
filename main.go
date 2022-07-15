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
		checks.CheckBranchProtection,
	}
	changeTypeToCheck := map[pkg.ChangeType]bool{pkg.Added: true}
	results, err := GetDependencyDiffResults(
		context.Background(), owner, repo, base, head, checksToRun, changeTypeToCheck,
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(results)
}

// Depdiff is the exported name for dependency-diff.
const Depdiff = "Dependency-diff"

type dependencydiffContext struct {
	logger                                *log.Logger
	ownerName, repoName, baseSHA, headSHA string
	ctx                                   context.Context
	ghRepo                                clients.Repo
	ghRepoClient                          clients.RepoClient
	ossFuzzClient                         clients.RepoClient
	vulnsClient                           clients.VulnerabilitiesClient
	ciiClient                             clients.CIIBestPracticesClient
	changeTypesToCheck                    map[pkg.ChangeType]bool
	checkNamesToRun                       []string
	dependencydiffs                       []dependency
	results                               []pkg.DependencyCheckResult
}

// GetDependencyDiffResults gets dependency changes between two given code commits BASE and HEAD
// along with the Scorecard check results of the dependencies, and returns a slice of DependencyCheckResult.
// TO use this API, an access token must be set following https://github.com/ossf/scorecard#authentication.
func GetDependencyDiffResults(
	ctx context.Context, ownerName, repoName, baseSHA, headSHA string, scorecardChecksNames []string,
	changeTypesToCheck map[pkg.ChangeType]bool) ([]pkg.DependencyCheckResult, error) {
	// Fetch the raw dependency diffs.
	dCtx := dependencydiffContext{
		logger:             log.NewLogger(log.InfoLevel),
		ownerName:          ownerName,
		repoName:           repoName,
		baseSHA:            baseSHA,
		headSHA:            headSHA,
		ctx:                ctx,
		changeTypesToCheck: changeTypesToCheck,
		checkNamesToRun:    scorecardChecksNames,
	}
	err := fetchRawDependencyDiffData(&dCtx)
	if err != nil {
		return nil, fmt.Errorf("error in fetchRawDependencyDiffData: %w", err)
	}

	// Initialize the repo and client(s) corresponding to the checks to run.
	err = initRepoAndClientByChecks(&dCtx)
	if err != nil {
		return nil, fmt.Errorf("error in initRepoAndClientByChecks: %w", err)
	}
	getScorecardCheckResults(&dCtx)
	return dCtx.results, nil
}

func initRepoAndClientByChecks(dCtx *dependencydiffContext) error {
	// Create the repo and the corresponding client if its check needs to run.
	ghRepo, err := githubrepo.MakeGithubRepo(path.Join(dCtx.ownerName, dCtx.repoName))
	if err != nil {
		return fmt.Errorf("error creating the github repo: %w", err)
	}
	dCtx.ghRepo = ghRepo
	dCtx.ghRepoClient = githubrepo.CreateGithubRepoClient(dCtx.ctx, dCtx.logger)
	// Initialize these three clients as nil at first.
	var ossFuzzClient clients.RepoClient
	for _, cn := range dCtx.checkNamesToRun {
		switch cn {
		case checks.CheckFuzzing:
			ossFuzzClient, err = githubrepo.CreateOssFuzzRepoClient(dCtx.ctx, dCtx.logger)
			if err != nil {
				return fmt.Errorf("error initializing the oss fuzz repo client: %w", err)
			}
			dCtx.ossFuzzClient = ossFuzzClient
		case checks.CheckVulnerabilities:
			dCtx.vulnsClient = clients.DefaultVulnerabilitiesClient()
		case checks.CheckCIIBestPractices:
			dCtx.ciiClient = clients.DefaultCIIBestPracticesClient()
		}
	}
	return nil
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

func getScorecardCheckResults(dCtx *dependencydiffContext) {
	// Initialize the checks to run from the caller's input.
	checksToRun := initScorecardChecks(dCtx.checkNamesToRun)
	for _, d := range dCtx.dependencydiffs {
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
		// TODO: use the BigQuery dataset to supplement null source repo URLs to fetch the Scorecard results for them.
		if d.SourceRepository != nil && *d.SourceRepository != "" {
			if d.ChangeType != nil && (dCtx.changeTypesToCheck[*d.ChangeType] || dCtx.changeTypesToCheck == nil) {
				// Run scorecard on those types of dependencies that the caller would like to check.
				// If the input map changeTypesToCheck is empty, by default, we run checks for all valid types.
				// TODO: use the Scorecare REST API to retrieve the Scorecard result statelessly.
				scorecardResult, err := scpkg.RunScorecards(
					dCtx.ctx,
					dCtx.ghRepo,
					// TODO: In future versions, ideally, this should be
					// the commitSHA corresponding to d.Version instead of HEAD.
					clients.HeadSHA,
					checksToRun,
					dCtx.ghRepoClient,
					dCtx.ossFuzzClient,
					dCtx.ciiClient,
					dCtx.vulnsClient,
				)
				// If the run fails, we leave the current dependency scorecard result empty and record the error
				// rather than letting the entire API return nil since we still expect results for other dependencies.
				if err != nil {
					depCheckResult.ScorecardResultsWithError.Error = fmt.Errorf("error running the scorecard checks: %w", err)
				} else { // Otherwise, we record the scorecard check results for this dependency.
					depCheckResult.ScorecardResultsWithError.ScorecardResults = &scorecardResult
				}
			}
		}
		dCtx.results = append(dCtx.results, depCheckResult)
	}
}
