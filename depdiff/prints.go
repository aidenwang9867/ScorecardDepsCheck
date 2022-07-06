package depdiff

import (
	"fmt"
)

func PrintDependencyToStdOut(d Dependency) {
	result := ""
	if d.IsDirect {
		result += fmt.Sprintf("name: %v \nstatus: %v \nversion: %v \necosys: %v \n",
			d.Name, d.ChangeType, d.Version, d.Ecosystem)
	} else {
		result += fmt.Sprintf("name: %v \nstatus: %v \nversion: %v \necosys: %v \n",
			d.Name, "indirect", d.Version, d.Ecosystem)
	}
	if d.PackageURL != "" {
		result += fmt.Sprintf("pkg_url: %v\n", d.PackageURL)
	}
	if d.SrcRepoURL != "" {
		result += fmt.Sprintf("src_url: %v\n", d.SrcRepoURL)
	}
	if len(d.Vulnerabilities) != 0 {
		result += fmt.Sprintln("vulnerabilities: ")
		for _, v := range d.Vulnerabilities {
			result += fmt.Sprintf("\tvuln_id: %v\n", v.ID)
			result += fmt.Sprintf("\tvuln_url: %v\n\n", v.SourceURL)
			if v.Title != "" {
				result += fmt.Sprintf("\tvuln_summary: %v\n", v.Title)
			}
			if v.GitHubSeverity != "" {
				result += fmt.Sprintf("\tseverity: %v\n", v.GitHubSeverity)
			}
			if v.Score.Valid {
				result += fmt.Sprintf("\tCVSS3Score: %v\n", v.Score.Float64)
			}
			if !v.DisclosedTime.IsZero() {
				result += fmt.Sprintf("\ttime_disclosed: %v\n", v.DisclosedTime)
			}
		}
	}
	fmt.Println(result)
}

// SprintDependencyDiffToMarkDown analyzes the dependency-diff fetched from the GitHub Dependency
// Review API, then parse them and return as a markdown string.
func SprintDependencyDiffToMarkDown(deps []Dependency) string {
	// Divide fetched depdendencies into added, updated, and removed ones.
	added, updated, removed :=
		map[string]Dependency{}, map[string]Dependency{}, map[string]Dependency{}
	for _, d := range deps {
		switch d.ChangeType {
		case Added:
			added[d.Name] = d
		case Removed:
			removed[d.Name] = d
		}
	}
	results := ""
	for dName, d := range added {
		// If a dependency name in the added map is also in the removed map,
		// then it's an updated dependency.
		if _, ok := removed[dName]; ok {
			updated[dName] = d
		} else {
			// Otherwise, it's an added dependency.
			current := changeTypeTag(Added)
			// Add the vulnerble tag for added dependencies if vuln found.
			if len(d.Vulnerabilities) != 0 {
				current += fmt.Sprintf(vulnTag(d))
			}
			current += fmt.Sprintf(
				"%s: %s @ %s\n\n",
				d.Ecosystem, d.Name, d.Version,
			)
			results += current
		}
	}
	for dName, d := range updated {
		current := changeTypeTag(Updated)
		// Add the vulnerble tag for updated dependencies if vuln found.
		if len(d.Vulnerabilities) != 0 {
			current += fmt.Sprintf(vulnTag(d))
		}
		current += fmt.Sprintf(
			" %s: %s @ %s (**new**) (bumped from %s: %s @ %s)\n\n",
			d.Ecosystem, d.Name, d.Version,
			added[dName].Ecosystem, added[dName].Name, added[dName].Version,
		)
		results += current
	}
	for dName, d := range removed {
		// If a dependency name in the removed map is not in the added map,
		// then it's a removed dependency.
		if _, ok := added[dName]; !ok {
			// We don't care vulnerbailities found in the removed dependencies.
			current := fmt.Sprintf(
				changeTypeTag(Removed)+" %s: %s @ %s\n\n",
				d.Ecosystem, d.Name, d.Version,
			)
			results += current + "\n\n"
		}
	}
	if results == "" {
		return fmt.Sprintln("No dependencies changed")
	} else {
		return results
	}
}

// changeTypeTag generates the change type markdown label for the dependency change type.
func changeTypeTag(ct ChangeType) string {
	switch ct {
	case Added, Updated:
		return fmt.Sprintf("**`"+"%s"+"`**", ct)
	case Removed:
		return fmt.Sprintf("~~**`"+"%s"+"`**~~", ct)
	default:
		return ""
	}
}

// vulnTag generates the vulnerable markdown label with a vulnerability reference URL.
func vulnTag(d Dependency) string {
	// TODO: which URL we should give as the vulnerability reference?
	result := fmt.Sprintf("[**`"+"vulnerable"+"`**](%s) ", d.SrcRepoURL)
	return result
}
