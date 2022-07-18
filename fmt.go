package main

import (
	"fmt"
	"sort"

	"github.com/aidenwang9867/DependencyDiffVisualizationInAction/pkg"
	docs "github.com/ossf/scorecard/v4/docs/checks"
)

type scoreAndDependencyName struct {
	aggregateScore float64
	dependencyName string
}

func PrintDependencies(deps []dependency) {
	for _, d := range deps {
		fmt.Println(*d.Ecosystem, d.Name, *d.Version, *d.ChangeType)
		if d.PackageURL != nil && *d.PackageURL != "" {
			fmt.Println(*d.PackageURL)
		} else {
			fmt.Println("empty package url")
		}
		if d.SourceRepository != nil && *d.SourceRepository != "" {
			fmt.Println(*d.SourceRepository)
		} else {
			fmt.Println("empty src url")
		}
		fmt.Printf("===================\n\n")
	}
}

func SprintDependencyChecksToMarkdown(dChecks []pkg.DependencyCheckResult) (string, error) {
	added := map[string]*pkg.DependencyCheckResult{}
	updated := map[string]*pkg.DependencyCheckResult{}
	removed := map[string]*pkg.DependencyCheckResult{}
	for _, d := range dChecks {
		if d.ChangeType != nil {
			switch *d.ChangeType {
			case pkg.Added:
				added[d.Name] = &d
			case pkg.Removed:
				removed[d.Name] = &d
			}
			// The current data source GitHub Dependency Review won't give the updated dependencies,
			// so we need to find them out manually by checking the added/removed maps.
		}
	}
	for dName := range added {
		if removed[dName] != nil {
			// If the dependency check result in the added map is also in the removed map
			// (removing the old version and adding the new version), move it to the updated map.
			updated[dName] = added[dName]
			added[dName] = nil // Remove it from the added map.
			// Will need its old package info in the removed map, so don't remove it from the removed map.
		}
	}
	// Sort dependencies by their aggregate scores in descending orders.
	addedSortKeys, err := getDependencySortKeys(added)
	if err != nil {
		return "", err
	}
	updatedSortKeys, err := getDependencySortKeys(updated)
	if err != nil {
		return "", err
	}
	removedSortKeys, err := getDependencySortKeys(removed)
	if err != nil {
		return "", err
	}
	sort.SliceStable(
		addedSortKeys,
		func(i, j int) bool { return addedSortKeys[i].aggregateScore > addedSortKeys[j].aggregateScore },
	)
	sort.SliceStable(
		updatedSortKeys,
		func(i, j int) bool { return updatedSortKeys[i].aggregateScore > updatedSortKeys[j].aggregateScore },
	)
	sort.SliceStable(
		removedSortKeys,
		func(i, j int) bool { return removedSortKeys[i].aggregateScore > removedSortKeys[j].aggregateScore },
	)

	results := ""
	for _, key := range addedSortKeys {
		dName := key.dependencyName
		current := fmt.Sprintf("**`" + "added" + "`** ")
		current += fmt.Sprintf(
			"%s: %s @ %s ",
			*added[dName].Ecosystem, added[dName].Name, *added[dName].Version,
		)
		if key.aggregateScore != -1 {
			current += fmt.Sprintf("`Scorecard Score: %.1f`", key.aggregateScore)
		}
		results += current + "\n\n"
	}
	for _, key := range updatedSortKeys {
		dName := key.dependencyName
		current := fmt.Sprintf(
			"**`" + "updated" + "`**")
		current += fmt.Sprintf(
			" %s: %s @ %s (**old**) :arrow_right: %s @ %s @ %s (**new**)",
			*updated[dName].Ecosystem, updated[dName].Name, *updated[dName].Version,
			*removed[dName].Ecosystem, removed[dName].Name, *removed[dName].Version,
		)
		if key.aggregateScore != -1 {
			current += fmt.Sprintf("`Scorecard Score: %.1f`", key.aggregateScore)
		}
		results += current + "\n\n"
	}
	for _, key := range removedSortKeys {
		dName := key.dependencyName
		if _, ok := updated[dName]; !ok {
			current := fmt.Sprintf(
				"~~**`"+"removed"+"`**~~"+" %s: %s @ %s",
				*removed[dName].Ecosystem, removed[dName].Name, *removed[dName].Version,
			)
			if key.aggregateScore != -1 {
				current += fmt.Sprintf("`Scorecard Score: %.1f`", key.aggregateScore)
			}
			results += current + "\n\n"
		}
	}
	return results, nil
}

func getDependencySortKeys(depCheckResults map[string]*pkg.DependencyCheckResult) ([]scoreAndDependencyName, error) {
	checkDocs, err := docs.Read()
	if err != nil {
		return nil, fmt.Errorf("error getting the check docs: %w", err)
	}
	sortKeys := []scoreAndDependencyName{}
	for _, dc := range depCheckResults {
		score := float64(-1)
		if dc.ScorecardResultsWithError.ScorecardResults != nil {
			aggregated, err := dc.ScorecardResultsWithError.ScorecardResults.GetAggregateScore(checkDocs)
			if err != nil {
				continue // We still want aggregate scores of other dependencies.
			}
			score = aggregated
		}
		sortKeys = append(sortKeys, scoreAndDependencyName{
			aggregateScore: score,
			dependencyName: dc.Name,
		})
	}
	return sortKeys, nil
}
