package main

import (
	"fmt"
	"sort"

	"github.com/aidenwang9867/depdiffvis/pkg"
	"github.com/ossf/scorecard/v4/checker"
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

func SprintDependencyChecksToMarkdown(dChecks []pkg.DependencyCheckResult) (*string, error) {
	// Use maps to reduce lookup times. Use pointers as values to save space.
	added := map[string]*pkg.DependencyCheckResult{}
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
			// so we need to find them manually by checking the added/removed maps.
		}
	}
	// Sort dependencies by their aggregate scores in descending orders.
	addedSortKeys, err := getDependencySortKeys(added)
	if err != nil {
		return nil, err
	}
	removedSortKeys, err := getDependencySortKeys(removed)
	if err != nil {
		return nil, err
	}
	sort.SliceStable(
		addedSortKeys,
		func(i, j int) bool { return addedSortKeys[i].aggregateScore > addedSortKeys[j].aggregateScore },
	)
	sort.SliceStable(
		removedSortKeys,
		func(i, j int) bool { return removedSortKeys[i].aggregateScore > removedSortKeys[j].aggregateScore },
	)
	results := ""
	for _, key := range addedSortKeys {
		dName := key.dependencyName
		new, old := added[dName], removed[dName]
		if new == nil {
			continue
		}
		current := addedTag()
		if old != nil {
			// Dependency in the added map also found in the removed map, indicating an updated one.
			current += updatedTag()
		}
		current += scoreTag(key.aggregateScore)
		current += fmt.Sprintf(
			" %s @ %s (new) ",
			new.Name, *new.Version,
		)
		if old != nil {
			current += fmt.Sprintf(
				" ~~%s @ %s (removed)~~ ",
				old.Name, *old.Version,
			)
		}
		results += current + "\n\n"
	}
	for _, key := range removedSortKeys {
		dName := key.dependencyName
		new, old := added[dName], removed[dName]
		if old == nil || new != nil {
			// Skip updated ones.
			continue
		}
		current := removedTag()
		current += scoreTag(key.aggregateScore)
		current += fmt.Sprintf(
			" ~~%s @ %s~~ ",
			old.Name, *old.Version,
		)
		results += current + "\n\n"
	}
	return &results, nil
}

func getDependencySortKeys(dcMap map[string]*pkg.DependencyCheckResult) ([]scoreAndDependencyName, error) {
	checkDocs, err := docs.Read()
	if err != nil {
		return nil, fmt.Errorf("error getting the check docs: %w", err)
	}
	sortKeys := []scoreAndDependencyName{}
	for _, v := range dcMap {
		score := float64(checker.InconclusiveResultScore)
		if v.ScorecardResultsWithError.ScorecardResults != nil {
			aggregated, err := v.ScorecardResultsWithError.ScorecardResults.GetAggregateScore(checkDocs)
			if err == nil {
				score = aggregated
			}
			// Don't return the err immediately since we still want aggregate scores of other dependencies.
		}
		sortKeys = append(sortKeys, scoreAndDependencyName{
			aggregateScore: score,
			dependencyName: v.Name,
		})
	}
	return sortKeys, nil
}

func addedTag() string {
	return fmt.Sprintf("**`" + "added" + "`**")
}

func updatedTag() string {
	return fmt.Sprintf("**`" + "updated" + "`**")
}

func removedTag() string {
	return fmt.Sprintf("~~**`" + "removed" + "`**~~")
}

func scoreTag(score float64) string {
	switch score {
	case float64(checker.InconclusiveResultScore):
		return fmt.Sprintf("`Scorecard Score: N/A`")
	default:
		return fmt.Sprintf("`Scorecard Score: %.1f`", score)
	}
}
